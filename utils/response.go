package utils

import "github.com/gin-gonic/gin"

func Success(c *gin.Context, status int, message string, data any) {
	c.JSON(status, gin.H{"message": message, "data": data})
}

func SuccessWithMeta(c *gin.Context, status int, message string, data any, meta any) {
	c.JSON(status, gin.H{"message": message, "data": data, "meta": meta})
}

func Error(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}

func ValidationError(c *gin.Context, details []string) {
	c.JSON(400, gin.H{"error": "Validation failed", "details": details})
}

func ConflictError(c *gin.Context, details []string) {
	c.JSON(409, gin.H{"error": "Schedule conflict detected", "details": details})
}
