package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// ConnectDB establishes and pings our connection pool
func ConnectDB() {
	var err error

	dsn := os.Getenv("DATABASE_URL")
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(fmt.Errorf("failed to open connection pool: %w", err))
	}

	log.Println("Successfully connected to PostgreSQL database!")
}
