package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost               string
	DBUser               string
	DBPassword           string
	DBName               string
	DBPort               string
	JWTSecret            string
	RedisAddr            string
	RedisPass            string
	RedisDB              int
	ServicePort          string
	BookingServiceAPIURL string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, assuming environment variables are set.")
	}

	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		log.Printf("Invalid REDIS_DB value, defaulting to 0: %v", err)
		redisDB = 0
	}

	return &Config{
		DBHost:               getEnv("DB_HOST", "localhost"),
		DBUser:               getEnv("DB_USER", "root"),
		DBPassword:           getEnv("DB_PASSWORD", "password"),
		DBName:               getEnv("DB_NAME", "payment_db"),
		DBPort:               getEnv("DB_PORT", "3306"),
		JWTSecret:            getEnv("JWT_SECRET", "anothersecretkey"),
		RedisAddr:            getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPass:            getEnv("REDIS_PASSWORD", ""),
		RedisDB:              redisDB,
		ServicePort:          getEnv("PORT", "8082"),
		BookingServiceAPIURL: getEnv("BOOKING_SERVICE_API_URL", "http://localhost:8081/api/v1"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
