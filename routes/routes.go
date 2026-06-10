package routes

import (
	"school-schedule-api/controllers"
	"school-schedule-api/middleware"
	"school-schedule-api/services"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	router := gin.Default()
	scheduleService := services.NewScheduleService(db)
	scheduleController := controllers.NewScheduleController(scheduleService)

	router.GET("/health", controllers.Health)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := router.Group("/api")
	api.Use(middleware.APIKeyMiddleware())
	{
		api.POST("/schedules", scheduleController.Create)
		api.GET("/schedules", scheduleController.FindAll)
		api.GET("/schedules/student", scheduleController.Student)
		api.GET("/schedules/teacher", scheduleController.Teacher)
		api.GET("/schedules/report/rekap-jp", scheduleController.RecapJP)
		api.POST("/schedules/upload", scheduleController.Upload)
		api.GET("/schedules/export", scheduleController.Export)
		api.GET("/schedules/:id", scheduleController.FindByID)
		api.PUT("/schedules/:id", scheduleController.Update)
		api.DELETE("/schedules/:id", scheduleController.Delete)
	}

	return router
}
