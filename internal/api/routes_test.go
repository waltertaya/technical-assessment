package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/waltertaya/doctor-appointment/internal/handlers"
	"github.com/waltertaya/doctor-appointment/internal/service"
	"gorm.io/gorm"
)

func TestSetupRoutesRegistersAPIEndpoints(t *testing.T) {
	router := SetupRoutes(handlers.NewHandler(service.NewBookingService(&gorm.DB{})))

	expected := map[string]bool{
		"GET /api/v1/health":                        false,
		"GET /api/v1/doc-test":                      false,
		"POST /api/v1/doctors":                      false,
		"POST /api/v1/doctors/:id/working-hours":    false,
		"POST /api/v1/patients":                     false,
		"POST /api/v1/appointments":                 false,
		"GET /api/v1/doctors/:id/availability":      false,
		"PATCH /api/v1/appointments/:id/cancel":     false,
		"PATCH /api/v1/appointments/:id/reschedule": false,
	}

	for _, route := range router.Routes() {
		key := route.Method + " " + route.Path
		if _, ok := expected[key]; ok {
			expected[key] = true
		}
	}

	for route, registered := range expected {
		if !registered {
			t.Errorf("expected route %s to be registered", route)
		}
	}
}

func TestSetupRoutesDocumentationEndpoint(t *testing.T) {
	router := SetupRoutes(handlers.NewHandler(service.NewBookingService(&gorm.DB{})))
	request := httptest.NewRequest(http.MethodGet, "/api/v1/doc-test", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if body := response.Body.String(); !strings.Contains(body, "<title>Appointment API</title>") {
		t.Fatalf("expected documentation page, got %q", body)
	}
}

func TestSetupRoutesHealthEndpoint(t *testing.T) {
	router := SetupRoutes(handlers.NewHandler(service.NewBookingService(&gorm.DB{})))
	request := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if body := response.Body.String(); body != `{"status":"healthy"}` {
		t.Fatalf("expected healthy response, got %q", body)
	}
}
