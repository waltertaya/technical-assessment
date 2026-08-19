package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// ConnectDB establishes and pings our connection pool
func ConnectDB(connStr string) (*sql.DB, error) {
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection pool: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL database!")
	return db, nil
}

// RunMigrations runs our initialization SQL script directly
func RunMigrations(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS doctors (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		specialty VARCHAR(100) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS patients (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS working_hours (
		id SERIAL PRIMARY KEY,
		doctor_id INT REFERENCES doctors(id) ON DELETE CASCADE,
		day_of_week INT CHECK (day_of_week BETWEEN 0 AND 6),
		start_time TIME NOT NULL,
		end_time TIME NOT NULL,
		CONSTRAINT chk_times CHECK (start_time < end_time)
	);

	CREATE TABLE IF NOT EXISTS appointments (
		id SERIAL PRIMARY KEY,
		doctor_id INT REFERENCES doctors(id) ON DELETE CASCADE,
		patient_id INT REFERENCES patients(id) ON DELETE CASCADE,
		appointment_date DATE NOT NULL,
		start_time TIME NOT NULL,
		end_time TIME NOT NULL,
		status VARCHAR(20) DEFAULT 'booked' CHECK (status IN ('booked', 'cancelled')),
		CONSTRAINT chk_app_times CHECK (start_time < end_time)
	);`

	_, err := db.Exec(query)
	return err
}
