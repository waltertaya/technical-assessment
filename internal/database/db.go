package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

var (
	openPostgres = func(dsn string) (*gorm.DB, error) {
		return gorm.Open(postgres.Open(dsn), &gorm.Config{})
	}
	logFatal = func(v ...any) {
		log.Fatal(v...)
	}
	logPrintln = func(v ...any) {
		log.Println(v...)
	}
)

// ConnectDB establishes and pings our connection pool
func ConnectDB() {
	var err error

	dsn := os.Getenv("DATABASE_URL")
	DB, err = openPostgres(dsn)
	if err != nil {
		logFatal(fmt.Errorf("failed to open connection pool: %w", err))
	}

	logPrintln("Successfully connected to PostgreSQL database!")
}
