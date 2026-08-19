package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/waltertaya/doctor-appointment/internal/models"
)

func (h *Handler) CreatePatient(c *gin.Context) {
	var patient models.Patient
	if err := c.ShouldBindJSON(&patient); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid patient: " + err.Error()})
		return
	}

	if err := h.Booking.DB.WithContext(c.Request.Context()).Create(&patient).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create patient"})
		return
	}
	c.JSON(http.StatusCreated, patient)
}
