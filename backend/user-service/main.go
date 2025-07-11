package main

import (
	"time"

	"backend/user-service/config"
	"backend/user-service/controllers"
	"backend/user-service/database"
	"backend/user-service/middlewares"
	"backend/user-service/repositories"
	"backend/user-service/services"
	"backend/user-service/utils"

	"github.com/gin-gonic/gin"

	_ "backend/user-service/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Concert Ticket Booking System - User Service API
// @version 1.0
// @description This is the User Service for a concert ticket booking system, handling user authentication and profile management.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1/users
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description "Type 'Bearer YOUR_TOKEN' to authenticate. Example: 'Bearer eyJhbGciOiJIUzI1Ni...'"
func main() {

	cfg := config.LoadConfig()
	utils.InitJWT(cfg)

	database.ConnectDB(cfg)

	middlewares.InitRedis(cfg)

	userRepo := repositories.NewUserRepository(database.DB)
	userService := services.NewUserService(userRepo)
	userController := controllers.NewUserController(userService)

	router := gin.Default()

	router.Use(middlewares.CORSMiddleware())
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.Use(middlewares.RateLimitMiddleware(100, 1*time.Minute))

	v1 := router.Group("/api/v1/users")
	{

		v1.POST("/register", userController.Register)
		v1.POST("/login", userController.Login)
		v1.POST("/logout", userController.Logout)

		authenticated := v1.Group("/")
		authenticated.Use(middlewares.AuthMiddleware())
		{
			authenticated.GET("/profile", userController.GetProfile)
			authenticated.PUT("/profile", userController.UpdateProfile)

		}
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	utils.LogInfo("User Service running on port %s", cfg.ServicePort)
	if err := router.Run(":" + cfg.ServicePort); err != nil {
		utils.LogError("Failed to start user service: %v", err)
	}
}
