package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/waltertaya/doctor-appointment/internal/database"
	"github.com/waltertaya/doctor-appointment/internal/models"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables.")
	}

	database.ConnectDB()
}

func main() {
	database.DB.AutoMigrate(&models.Doctor{}, &models.Patient{}, &models.WorkingHours{}, &models.Appointment{})

	log.Println("Database migration completed successfully!")
}
