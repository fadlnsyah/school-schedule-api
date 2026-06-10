package main

import (
	"os"

	"github.com/gin-gonic/gin"
)

// @title School Schedule API
// @version 1.0
// @description REST API for school schedule management.
// @BasePath /
func main() {
	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "School Schedule API is running",
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = os.Getenv("APP_PORT")
	}
	if port == "" {
		port = "8080"
	}

	if err := router.Run(":" + port); err != nil {
		panic(err)
	}
}
