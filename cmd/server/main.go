package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/joho/godotenv"
	"github.com/waltertaya/doctor-appointment/internal/api"
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
	h := handlers.NewHandler(service.NewBookingService(database.DB))
	r := api.SetupRoutes(h)

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))

	log.Println("Starting Gin server on :8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
