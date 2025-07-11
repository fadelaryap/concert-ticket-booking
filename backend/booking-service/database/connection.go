package database

import (
	"fmt"
	"log"
	"time"

	"backend/booking-service/config"
	"backend/booking-service/models"

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

	log.Println("Database connection established!")

	if !DB.Migrator().HasTable("booking_seats") {
		log.Println("Creating booking_seats join table...")
		err = DB.Exec(`
			CREATE TABLE IF NOT EXISTS booking_seats (
				booking_id VARCHAR(36) NOT NULL, -- Changed to VARCHAR(36)
				seat_id BIGINT UNSIGNED NOT NULL,
				PRIMARY KEY (booking_id, seat_id),
				CONSTRAINT fk_booking_seats_booking FOREIGN KEY (booking_id) REFERENCES bookings (id) ON DELETE CASCADE,
				CONSTRAINT fk_booking_seats_seat FOREIGN KEY (seat_id) REFERENCES seats (id) ON DELETE CASCADE
			) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
		`).Error
		if err != nil {
			log.Fatalf("Failed to create booking_seats join table: %v", err)
		}
	}

	err = DB.AutoMigrate(&models.Buyer{}, &models.TicketHolder{})
	if err != nil {
		log.Fatalf("Failed to auto migrate Buyer/TicketHolder tables: %v", err)
	}
	log.Println("Buyer and TicketHolder tables migrated successfully!")
}
