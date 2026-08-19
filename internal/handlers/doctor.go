package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/waltertaya/doctor-appointment/internal/models"
	"gorm.io/gorm"
)

type WorkingHoursRequest struct {
	DayOfWeek int    `json:"day_of_week" binding:"min=0,max=6"`
	StartTime string `json:"start_time" binding:"required,datetime=15:04:05"`
	EndTime   string `json:"end_time" binding:"required,datetime=15:04:05"`
}

func (h *Handler) CreateDoctor(c *gin.Context) {
	var doctor models.Doctor
	if err := c.ShouldBindJSON(&doctor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid doctor: " + err.Error()})
		return
	}

	if err := h.Booking.DB.WithContext(c.Request.Context()).Create(&doctor).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create doctor"})
		return
	}
	c.JSON(http.StatusCreated, doctor)
}

func (h *Handler) CreateWorkingHours(c *gin.Context) {
	doctorID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid doctor ID"})
		return
	}

	var req WorkingHoursRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid working hours: " + err.Error()})
		return
	}
	start, _ := time.Parse("15:04:05", req.StartTime)
	end, _ := time.Parse("15:04:05", req.EndTime)
	if !start.Before(end) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Start time must be before end time"})
		return
	}

	db := h.Booking.DB.WithContext(c.Request.Context())
	var doctor models.Doctor
	if err := db.First(&doctor, "id = ?", doctorID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Doctor not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find doctor"})
		return
	}

	shift := models.WorkingHours{
		DoctorID:  doctorID,
		DayOfWeek: req.DayOfWeek,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}
	if err := db.Create(&shift).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create working hours"})
		return
	}
	c.JSON(http.StatusCreated, shift)
}
