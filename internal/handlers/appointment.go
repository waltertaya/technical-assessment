package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Defines the input payload
type BookAppointmentRequest struct {
	DoctorID        uuid.UUID `json:"doctor_id" binding:"required"`
	PatientID       uuid.UUID `json:"patient_id" binding:"required"`
	AppointmentDate string    `json:"appointment_date" binding:"required"` // Format: YYYY-MM-DD
	StartTime       string    `json:"start_time" binding:"required"`       // Format: HH:MM
}

// Handles booking logic inside our Gin controller
func (h *Handler) BookAppointment(c *gin.Context) {
	var req BookAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	targetDate, err := time.Parse("2006-01-02", req.AppointmentDate)
	if err != nil || !targetDate.After(time.Now()) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Appointment date must be in the future"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Appointment booked successfully"})
}
