package middlewares

import (
	"context"
	"log"
	"net/http"
	"time"

	"backend/payment-service/config"
	"backend/payment-service/utils"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func InitRedis(cfg *config.Config) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPass,
		DB:       cfg.RedisDB,
	})

	var err error
	var counts uint8 = 1
	const maxRetries = 10
	const retryDelay = 3 * time.Second

	for counts <= maxRetries {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err = client.Ping(ctx).Result()
		cancel()

		if err != nil {
			log.Printf("Attempt %d/%d: Failed to connect to Redis: %v. Retrying in %s...", counts, maxRetries, err, retryDelay)
			time.Sleep(retryDelay)
			counts++
			continue
		} else {
			utils.LogInfo("Connected to Redis successfully for rate limiting!")
			RedisClient = client
			return
		}
	}

	if err != nil {
		log.Fatalf("Failed to connect to Redis after %d retries: %v", maxRetries, err)
	}
}

func RateLimitMiddleware(maxRequests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := "ratelimit:" + ip

		ctx := context.Background()

		count, err := RedisClient.Incr(ctx, key).Result()
		if err != nil {
			utils.LogError("Redis INCR error for rate limit: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Rate limiting service error"})
			c.Abort()
			return
		}

		if count == 1 {
			RedisClient.Expire(ctx, key, window)
		}

		if count > int64(maxRequests) {
			utils.LogWarning("Rate limit exceeded for IP: %s (Limit: %d requests/%s)", ip, maxRequests, window)
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests. Please try again later."})
			c.Abort()
			return
		}

		c.Next()
	}
}
