package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"backend/booking-service/utils"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RateLimitMiddleware(maxRequests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s", clientIP)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		count, err := utils.RedisClient.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			utils.LogError("Redis error in rate limit middleware for IP %s: %v", clientIP, err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		if err == redis.Nil {

			_, err = utils.RedisClient.Set(ctx, key, 1, window).Result()
			if err != nil {
				utils.LogError("Redis error setting rate limit for IP %s: %v", clientIP, err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
				return
			}
		} else {

			_, err = utils.RedisClient.Incr(ctx, key).Result()
			if err != nil {
				utils.LogError("Redis error incrementing rate limit for IP %s: %v", clientIP, err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
				return
			}
		}

		if count >= maxRequests {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			return
		}

		c.Next()
	}
}
