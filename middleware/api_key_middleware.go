package middleware

import (
	"net/http"

	"school-schedule-api/config"

	"github.com/gin-gonic/gin"
)

func APIKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		expectedAPIKey := config.GetEnv("API_KEY", "")
		if expectedAPIKey == "" || c.GetHeader("x-api-key") != expectedAPIKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		c.Next()
	}
}
