package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/waltertaya/doctor-appointment/internal/database"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables.")
	}

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5432/clinic_booking?sslmode=disable"
	}

	db, err := database.ConnectDB(connStr)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()

	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Database migrations failed: %v", err)
	}
	log.Println("Database migrations completed successfully.")
}

func main() {
	r := gin.Default()

	api := r.Group("/api/v1")
	{
		// api.POST("/appointments", handlers.BookAppointment)
		// api.GET("/doctors/:id/availability", handlers.GetAvailability)
		// api.PATCH("/appointments/:id/cancel", handlers.CancelAppointment)
		// api.PATCH("/appointments/:id/reschedule", handlers.RescheduleAppointment)
	}

	log.Println("Starting Gin server on :8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
