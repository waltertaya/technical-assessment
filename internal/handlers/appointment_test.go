package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/waltertaya/doctor-appointment/internal/service"
	"gorm.io/gorm"
)

func TestBookAppointmentValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(service.NewBookingService(&gorm.DB{}))

	tests := []struct {
		name     string
		body     string
		expected int
	}{
		{name: "malformed JSON", body: `{`, expected: http.StatusBadRequest},
		{name: "missing required fields", body: `{}`, expected: http.StatusBadRequest},
		{name: "invalid date", body: `{"doctor_id":"00000000-0000-0000-0000-000000000001","patient_id":"00000000-0000-0000-0000-000000000002","appointment_date":"tomorrow","start_time":"09:00"}`, expected: http.StatusUnprocessableEntity},
		{name: "invalid start time", body: `{"doctor_id":"00000000-0000-0000-0000-000000000001","patient_id":"00000000-0000-0000-0000-000000000002","appointment_date":"2099-01-01","start_time":"9am"}`, expected: http.StatusBadRequest},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := gin.New()
			router.POST("/appointments", handler.BookAppointment)
			request := httptest.NewRequest(http.MethodPost, "/appointments", strings.NewReader(test.body))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			if response.Code != test.expected {
				t.Fatalf("expected status %d, got %d: %s", test.expected, response.Code, response.Body.String())
			}
		})
	}
}

func TestCancelAppointmentValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(service.NewBookingService(&gorm.DB{}))
	router := gin.New()
	router.PATCH("/appointments/:id/cancel", handler.CancelAppointment)

	request := httptest.NewRequest(http.MethodPatch, "/appointments/not-an-id/cancel", strings.NewReader(`{}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}
