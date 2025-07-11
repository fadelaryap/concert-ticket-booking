package middlewares

import (
	"net/http"

	"backend/payment-service/utils"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("token")
		if err != nil {
			if err == http.ErrNoCookie {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: No token cookie found"})
				c.Abort()
				return
			}
			utils.LogError("Error getting token cookie: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error reading token"})
			c.Abort()
			return
		}

		claims, err := utils.ParseJWT(tokenString)
		if err != nil {
			utils.LogError("JWT parsing error from cookie: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}

func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		AuthMiddleware()(c)

		if c.IsAborted() {
			return
		}

		role, exists := c.Get("role")
		if !exists || role.(string) != "admin" {
			utils.LogWarning("Unauthorized access attempt: User %s (ID: %d) tried to access admin route", c.GetString("username"), c.GetUint("userID"))
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Requires admin role"})
			c.Abort()
			return
		}
		c.Next()
	}
}
