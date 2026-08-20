package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBookingServiceBookRejectsInvalidDates(t *testing.T) {
	service := NewBookingService(nil)
	input := BookInput{
		DoctorID:        uuid.New(),
		PatientID:       uuid.New(),
		AppointmentDate: "not-a-date",
		StartTime:       "09:00",
	}

	err := service.Book(context.Background(), input)
	if !errors.Is(err, ErrOutsideWorkingHours) {
		t.Fatalf("expected ErrOutsideWorkingHours, got %v", err)
	}
}

func TestBookingServiceBookRejectsPastDates(t *testing.T) {
	service := NewBookingService(nil)
	input := BookInput{
		DoctorID:        uuid.New(),
		PatientID:       uuid.New(),
		AppointmentDate: time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
		StartTime:       "09:00",
	}

	err := service.Book(context.Background(), input)
	if !errors.Is(err, ErrOutsideWorkingHours) {
		t.Fatalf("expected ErrOutsideWorkingHours, got %v", err)
	}
}

func TestBookingServiceBookRejectsInvalidStartTime(t *testing.T) {
	service := NewBookingService(nil)
	input := BookInput{
		DoctorID:        uuid.New(),
		PatientID:       uuid.New(),
		AppointmentDate: time.Now().AddDate(0, 0, 1).Format("2006-01-02"),
		StartTime:       "not-a-time",
	}

	err := service.Book(context.Background(), input)
	if err == nil {
		t.Fatal("expected invalid start time error")
	}
	if errors.Is(err, ErrOutsideWorkingHours) {
		t.Fatalf("expected time parsing error, got %v", err)
	}
}

func TestAppointmentTimesCreatesThirtyMinuteSlot(t *testing.T) {
	start, end := appointmentTimes(time.Date(2026, time.August, 20, 9, 15, 0, 0, time.UTC))

	if start != "09:15:00" {
		t.Errorf("expected start time 09:15:00, got %s", start)
	}
	if end != "09:45:00" {
		t.Errorf("expected end time 09:45:00, got %s", end)
	}
}

func TestSlotsForShiftIncludesOnlyCompleteThirtyMinuteSlots(t *testing.T) {
	tests := []struct {
		name     string
		start    string
		end      string
		expected []string
	}{
		{
			name:     "exact shift boundary",
			start:    "09:00:00",
			end:      "10:30:00",
			expected: []string{"09:00", "09:30", "10:00"},
		},
		{
			name:     "shorter than one slot",
			start:    "09:00:00",
			end:      "09:29:00",
			expected: nil,
		},
		{
			name:     "invalid time",
			start:    "invalid",
			end:      "10:00:00",
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := slotsForShift(test.start, test.end)
			if len(actual) != len(test.expected) {
				t.Fatalf("expected %v, got %v", test.expected, actual)
			}
			for index := range test.expected {
				if actual[index] != test.expected[index] {
					t.Errorf("slot %d: expected %q, got %q", index, test.expected[index], actual[index])
				}
			}
		})
	}
}

func TestSlotsForShiftRejectsReversedShift(t *testing.T) {
	if slots := slotsForShift("17:00:00", "09:00:00"); len(slots) != 0 {
		t.Fatalf("expected no slots for a reversed shift, got %v", slots)
	}
}
