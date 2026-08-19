package api

import (
	"github.com/gin-gonic/gin"
	"github.com/waltertaya/doctor-appointment/internal/handlers"
)

func SetupRoutes(h *handlers.Handler) *gin.Engine {
	r := gin.Default()

	api := r.Group("/api/v1")
	{
		api.POST("/doctors", h.CreateDoctor)
		api.POST("/doctors/:id/working-hours", h.CreateWorkingHours)
		api.POST("/patients", h.CreatePatient)
		api.POST("/appointments", h.BookAppointment)
		api.GET("/doctors/:id/availability", h.GetAvailability)
		api.PATCH("/appointments/:id/cancel", h.CancelAppointment)
		api.PATCH("/appointments/:id/reschedule", h.RescheduleAppointment)
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "healthy"})
		})
	}

	return r
}
