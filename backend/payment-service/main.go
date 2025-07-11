package main

import (
	"time"

	"backend/payment-service/config"
	"backend/payment-service/controllers"
	"backend/payment-service/database"
	"backend/payment-service/middlewares"
	"backend/payment-service/repositories"
	"backend/payment-service/services"
	"backend/payment-service/utils"

	_ "backend/payment-service/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Concert Ticket Booking System - Payment Service API
// @version 1.0
// @description This is the Payment Service for a concert ticket booking system, handling payment processing and status updates.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8082
// @BasePath /api/v1
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description "Type 'Bearer YOUR_TOKEN' to authenticate. Example: 'Bearer eyJhbGciOiJIUzI1Ni...'"
func main() {
	cfg := config.LoadConfig()
	utils.InitJWT(cfg)
	database.ConnectDB(cfg)
	middlewares.InitRedis(cfg)

	paymentRepo := repositories.NewPaymentRepository(database.DB)

	bookingServiceAPIURL := cfg.BookingServiceAPIURL

	paymentService := services.NewPaymentService(paymentRepo, bookingServiceAPIURL)

	paymentController := controllers.NewPaymentController(paymentService)

	router := gin.Default()
	router.Use(middlewares.CORSMiddleware())
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.Use(middlewares.RateLimitMiddleware(200, 1*time.Minute))

	v1 := router.Group("/api/v1")
	{
		payments := v1.Group("/payments")

		payments.Use(middlewares.AuthMiddleware())
		{
			payments.POST("/", paymentController.ProcessPayment)
			payments.GET("/:id", paymentController.GetPaymentByID)

		}
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	utils.LogInfo("Payment Service running on port %s", cfg.ServicePort)
	if err := router.Run(":" + cfg.ServicePort); err != nil {
		utils.LogError("Failed to start payment service: %v", err)
	}
}
