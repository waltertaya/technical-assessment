package models

import (
	"time"

	"github.com/google/uuid"
)

type Doctor struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Specialty string    `json:"specialty"`
	CreatedAt time.Time `json:"created_at"`
}

type Patient struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type WorkingHours struct {
	ID        uuid.UUID `json:"id"`
	DoctorID  uuid.UUID `json:"doctor_id"`
	DayOfWeek int       `json:"day_of_week"` // 0 = Sunday, 1 = Monday, ...
	StartTime string    `json:"start_time"`  // DB format HH:MM:SS mapped to string
	EndTime   string    `json:"end_time"`
}

type Appointment struct {
	ID              uuid.UUID `json:"id"`
	DoctorID        uuid.UUID `json:"doctor_id"`
	PatientID       uuid.UUID `json:"patient_id"`
	AppointmentDate time.Time `json:"appointment_date"` // YYYY-MM-DD
	StartTime       string    `json:"start_time"`       // HH:MM:SS
	EndTime         string    `json:"end_time"`
	Status          string    `json:"status"` // booked, cancelled
}
