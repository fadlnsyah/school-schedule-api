package config

import (
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	_ = godotenv.Load()
}

func GetEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func AppPort() string {
	port := os.Getenv("PORT")
	if port != "" {
		return port
	}

	return GetEnv("APP_PORT", "8080")
}
