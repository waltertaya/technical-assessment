package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/waltertaya/doctor-appointment/internal/service"
)

type Handler struct {
	Booking *service.BookingService
}

func NewHandler(booking *service.BookingService) *Handler {
	return &Handler{Booking: booking}
}

type BookRequest struct {
	DoctorID        uuid.UUID `json:"doctor_id" binding:"required"`
	PatientID       uuid.UUID `json:"patient_id" binding:"required"`
	AppointmentDate string    `json:"appointment_date" binding:"required"`
	StartTime       string    `json:"start_time" binding:"required"`
}

type CancelRequest struct {
	Reason string `json:"reason" binding:"required"`
}

type RescheduleRequest struct {
	NewDate      string `json:"new_date" binding:"required"`
	NewStartTime string `json:"new_start_time" binding:"required"`
}

func (h *Handler) BookAppointment(c *gin.Context) {
	var req BookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}
	var err error
	var appointmentDate time.Time
	appointmentDate, err = time.Parse("2006-01-02", req.AppointmentDate)
	if err != nil || !appointmentDate.After(time.Now()) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Appointment date must be in the future"})
		return
	}
	if _, err := time.Parse("15:04", req.StartTime); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start time format"})
		return
	}

	err = h.Booking.Book(c.Request.Context(), service.BookInput{
		DoctorID:        req.DoctorID,
		PatientID:       req.PatientID,
		AppointmentDate: req.AppointmentDate,
		StartTime:       req.StartTime,
	})
	if errors.Is(err, service.ErrDoctorNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Doctor not found"})
		return
	}
	if errors.Is(err, service.ErrOutsideWorkingHours) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Requested slot is outside doctor's working hours"})
		return
	}
	if errors.Is(err, service.ErrSlotBooked) {
		c.JSON(http.StatusConflict, gin.H{"error": "This time slot is already booked"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Booking transaction failed"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "booked", "message": "Appointment successfully booked"})
}

func (h *Handler) GetAvailability(c *gin.Context) {
	doctorID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid doctor ID"})
		return
	}
	dateStr := c.Query("date")
	if _, err := time.Parse("2006-01-02", dateStr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}
	slots, err := h.Booking.Availability(c.Request.Context(), doctorID, dateStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve availability"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"date": dateStr, "available_slots": slots})
}

func (h *Handler) CancelAppointment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid appointment ID"})
		return
	}
	var req CancelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cancellation reason is required"})
		return
	}
	err = h.Booking.Cancel(c.Request.Context(), id)
	if errors.Is(err, service.ErrAppointmentNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
		return
	}
	if errors.Is(err, service.ErrAppointmentCancelled) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Appointment is already cancelled"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel appointment"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "cancelled", "message": "Appointment cancelled successfully"})
}

func (h *Handler) RescheduleAppointment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid appointment ID"})
		return
	}
	var req RescheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload: " + err.Error()})
		return
	}
	var newDate time.Time
	newDate, err = time.Parse("2006-01-02", req.NewDate)
	if err != nil || !newDate.After(time.Now()) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "New appointment date must be in the future"})
		return
	}
	if _, err := time.Parse("15:04", req.NewStartTime); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid new start time format"})
		return
	}

	err = h.Booking.Reschedule(c.Request.Context(), id, req.NewDate, req.NewStartTime)
	if errors.Is(err, service.ErrAppointmentNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
		return
	}
	if errors.Is(err, service.ErrAppointmentCancelled) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Cannot reschedule a cancelled appointment"})
		return
	}
	if errors.Is(err, service.ErrNewOutsideWorkingHours) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "New slot is outside doctor's working hours"})
		return
	}
	if errors.Is(err, service.ErrNewSlotBooked) {
		c.JSON(http.StatusConflict, gin.H{"error": "New slot is already booked"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Rescheduling failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Appointment successfully rescheduled"})
}
