package database

import (
	"fmt"
	"log"
	"time"

	"backend/user-service/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB(cfg *config.Config) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	var dbErr error
	var counts uint8 = 1
	const maxRetries = 20
	const retryDelay = 5 * time.Second

	for counts <= maxRetries {
		DB, dbErr = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if dbErr != nil {
			log.Printf("Attempt %d/%d: Failed to connect to database: %v. Retrying in %s...", counts, maxRetries, dbErr, retryDelay)
			time.Sleep(retryDelay)
			counts++
			continue
		} else {
			log.Println("Database connected successfully!")
			break
		}
	}

	if dbErr != nil {
		log.Fatalf("Failed to connect to database after %d retries: %v", maxRetries, dbErr)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("Failed to get generic database object: %v", err)
	}

	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	log.Println("Database connected successfully!")

	log.Println("Database migrations completed!")
}
