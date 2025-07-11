package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/booking-service/config"
	"backend/booking-service/controllers"
	"backend/booking-service/database"
	"backend/booking-service/middlewares"
	"backend/booking-service/models"

	"backend/booking-service/repositories"
	"backend/booking-service/services"
	"backend/booking-service/utils"

	_ "backend/booking-service/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Concert Ticket Booking System - Booking Service API
// @version 1.0
// @description This is the Booking Service for a concert ticket booking system, handling concert information and ticket reservations.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8081
// @BasePath /api/v1
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description "Type 'Bearer' followed by a space and JWT token."
func main() {
	cfg := config.LoadConfig()
	utils.InitJWT(cfg)

	database.ConnectDB(cfg)

	if err := utils.InitRedis(cfg.RedisAddr, cfg.RedisPass, cfg.RedisDB); err != nil {
		log.Fatalf("Failed to initialize Redis for rate limiting: %v", err)
	}
	defer utils.CloseRedisConnection()

	if err := utils.InitRabbitMQ(cfg.RabbitMQURL); err != nil {
		log.Fatalf("Failed to initialize RabbitMQ: %v", err)
	}
	defer utils.CloseRabbitMQConnection()
	concertRepo := repositories.NewConcertRepository(database.DB)
	seatRepo := repositories.NewSeatRepository(database.DB)
	bookingRepo := repositories.NewBookingRepository(database.DB)
	ticketClassRepo := repositories.NewTicketClassRepository(database.DB)
	buyerRepo := repositories.NewBuyerRepository(database.DB)
	ticketHolderRepo := repositories.NewTicketHolderRepository(database.DB)

	concertService := services.NewConcertService(concertRepo, seatRepo, ticketClassRepo)

	bookingService := services.NewBookingService(bookingRepo, concertRepo, seatRepo, ticketClassRepo, buyerRepo, ticketHolderRepo, cfg.PaymentServiceAPIURL)

	go func() {
		msgs, err := utils.ConsumeMessages(utils.SeatCreationQueue())
		if err != nil {
			utils.LogError("Failed to start consuming from seat creation queue: %v", err)
			return
		}
		for d := range msgs {
			if err := concertService.ProcessSeatCreationMessage(d.Body); err != nil {
				utils.LogError("Error processing seat creation message: %v", err)
			}
		}
	}()

	go func() {
		msgs, err := utils.ConsumeMessages(utils.BookingCancellationQueue())
		if err != nil {
			utils.LogError("Failed to start consuming from booking cancellation queue: %v", err)
			return
		}
		for d := range msgs {
			utils.LogInfo("Received a message on booking_cancellation_queue: %s", d.Body)

		}
	}()

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			utils.LogInfo("Running scheduled task: Checking for expired pending bookings")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			err := bookingService.CancelExpiredPendingBookings(ctx)
			if err != nil {
				utils.LogError("Error during automatic booking cancellation: %v", err)
			}
			cancel()
		}
	}()

	go func() {

		time.Sleep(5 * time.Second)
		utils.LogInfo("Checking for concerts with 'pending_seat_creation' status on startup...")

		pendingConcerts, err := concertRepo.GetConcertsByStatus(models.ConcertStatusPendingSeatCreation)
		if err != nil {
			utils.LogError("Failed to get pending_seat_creation concerts on startup: %v", err)
			return
		}

		for _, c := range pendingConcerts {
			utils.LogInfo("Found pending_seat_creation concert (ID: %d, Name: %s). Attempting to trigger seat creation message...", c.ID, c.Name)

			var tcMessages []models.TicketClassMessage
			for _, tc := range c.TicketClasses {
				tcMessages = append(tcMessages, models.TicketClassMessage{
					TicketClassID:     tc.ID,
					Name:              tc.Name,
					TotalSeatsInClass: tc.TotalSeatsInClass,
				})
			}
			msg := models.SeatCreationMessage{
				ConcertID:     c.ID,
				TotalSeats:    c.TotalSeats,
				TicketClasses: tcMessages,
			}
			msgBody, err := json.Marshal(msg)
			if err != nil {
				utils.LogError("Failed to marshal seat creation message for concert %d on startup: %v", c.ID, err)
				continue
			}

			if err := utils.PublishMessage("", utils.SeatCreationQueue(), msgBody); err != nil {
				utils.LogError("Failed to publish seat creation message for concert %d on startup to RabbitMQ: %v", c.ID, err)

				continue
			}
			utils.LogInfo("Successfully published seat creation message for concert %d on startup", c.ID)
		}
	}()

	concertController := controllers.NewConcertController(concertService)
	bookingController := controllers.NewBookingController(bookingService)

	router := gin.Default()
	router.RedirectTrailingSlash = false

	router.Use(middlewares.CORSMiddleware())
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middlewares.RateLimitMiddleware(100, 1*time.Minute))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.NewHandler()))

	v1 := router.Group("/api/v1")
	{

		v1.GET("/concerts", concertController.GetConcerts)
		v1.GET("/concerts/:id", concertController.GetConcertByID)
		v1.GET("/concerts/:id/seats", concertController.GetConcertSeats)

		adminConcerts := v1.Group("/admin/concerts")
		adminConcerts.Use(middlewares.AuthMiddleware())
		adminConcerts.Use(middlewares.AdminAuthMiddleware())
		{
			adminConcerts.POST("/", concertController.CreateConcert)
		}

		bookings := v1.Group("/bookings")
		bookings.Use(middlewares.AuthMiddleware())
		{
			bookings.POST("/", bookingController.CreateBooking)
			bookings.GET("/my", bookingController.GetMyBookings)
			bookings.GET("/:id", bookingController.GetBookingByID)
			bookings.PUT("/:id/cancel", bookingController.CancelBooking)
		}

		v1.PUT("/internal/bookings/:id/status", bookingController.UpdateBookingStatusInternal)
	}

	go func() {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8081"
		}
		log.Printf("Booking Service running on port %s", port)
		if err := router.Run(fmt.Sprintf(":%s", port)); err != nil {
			log.Fatalf("Booking Service failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down Booking Service gracefully...")
	log.Println("Booking Service stopped.")
}
