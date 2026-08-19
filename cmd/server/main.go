package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/waltertaya/doctor-appointment/internal/database"
	"github.com/waltertaya/doctor-appointment/internal/handlers"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables.")
	}
}

func main() {
	r := gin.Default()

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("DATABASE_URL environment variable is not set.")
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

	// initialize handlers with db connection
	h := &handlers.Handler{DB: db}

	api := r.Group("/api/v1")
	{
		api.POST("/appointments", h.BookAppointment)
		// api.GET("/doctors/:id/availability", h.GetAvailability)
		// api.PATCH("/appointments/:id/cancel", h.CancelAppointment)
		// api.PATCH("/appointments/:id/reschedule", h.RescheduleAppointment)
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
