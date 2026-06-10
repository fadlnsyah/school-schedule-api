package main

import (
	"log"

	"school-schedule-api/config"
	"school-schedule-api/models"
	"school-schedule-api/routes"
)

// @title School Schedule API
// @version 1.0
// @description REST API for school schedule management.
// @BasePath /
func main() {
	config.LoadEnv()

	db, err := config.ConnectDatabase()
	if err != nil {
		log.Fatal("failed to connect database")
	}
	if err := db.AutoMigrate(&models.Schedule{}); err != nil {
		log.Fatal("failed to migrate database")
	}

	router := routes.SetupRouter(db)
	if err := router.Run(":" + config.AppPort()); err != nil {
		log.Fatal("failed to start server")
	}
}
