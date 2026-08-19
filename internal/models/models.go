package models

import (
	"time"

	"github.com/google/uuid"
)

type Doctor struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey" validate:"omitempty,uuid4"`
	Name      string    `json:"name" gorm:"type:varchar(100);not null" validate:"required,max=100"`
	Specialty string    `json:"specialty" gorm:"type:varchar(100);not null" validate:"required,max=100"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

type Patient struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey" validate:"omitempty,uuid4"`
	Name      string    `json:"name" gorm:"type:varchar(100);not null" validate:"required,max=100"`
	Email     string    `json:"email" gorm:"type:varchar(100);uniqueIndex;not null" validate:"required,email,max=100"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

type WorkingHours struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey" validate:"omitempty,uuid4"`
	DoctorID  uuid.UUID `json:"doctor_id" gorm:"type:uuid;not null" validate:"required,uuid4"`
	DayOfWeek int       `json:"day_of_week" gorm:"not null" validate:"min=0,max=6"` // 0 = Sunday, 1 = Monday, ...
	StartTime string    `json:"start_time" gorm:"type:time;not null" validate:"required,datetime=15:04:05"`
	EndTime   string    `json:"end_time" gorm:"type:time;not null" validate:"required,datetime=15:04:05"`
}

type Appointment struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey" validate:"omitempty,uuid4"`
	DoctorID        uuid.UUID `json:"doctor_id" gorm:"type:uuid;not null" validate:"required,uuid4"`
	PatientID       uuid.UUID `json:"patient_id" gorm:"type:uuid;not null" validate:"required,uuid4"`
	AppointmentDate time.Time `json:"appointment_date" gorm:"type:date;not null" validate:"required"`             // YYYY-MM-DD
	StartTime       string    `json:"start_time" gorm:"type:time;not null" validate:"required,datetime=15:04:05"` // HH:MM:SS
	EndTime         string    `json:"end_time" gorm:"type:time;not null" validate:"required,datetime=15:04:05"`
	Status          string    `json:"status" gorm:"type:varchar(20);default:booked;not null" validate:"required,oneof=booked cancelled"` // booked, cancelled
}
