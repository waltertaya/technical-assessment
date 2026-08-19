package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/waltertaya/doctor-appointment/internal/database"
	"github.com/waltertaya/doctor-appointment/internal/handlers"
	"github.com/waltertaya/doctor-appointment/internal/service"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables.")
	}

	database.ConnectDB()
}

func main() {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))

	h := handlers.NewHandler(service.NewBookingService(database.DB))
	api := r.Group("/api/v1")
	{
		api.POST("/appointments", h.BookAppointment)
		api.GET("/doctors/:id/availability", h.GetAvailability)
		api.PATCH("/appointments/:id/cancel", h.CancelAppointment)
		api.PATCH("/appointments/:id/reschedule", h.RescheduleAppointment)
	}

	// health check endpoint
	r.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	log.Println("Starting Gin server on :8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
