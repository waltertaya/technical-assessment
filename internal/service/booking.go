package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/waltertaya/doctor-appointment/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrDoctorNotFound         = errors.New("doctor not found")
	ErrOutsideWorkingHours    = errors.New("requested slot is outside working hours")
	ErrSlotBooked             = errors.New("appointment slot is already booked")
	ErrAppointmentNotFound    = errors.New("appointment not found")
	ErrAppointmentCancelled   = errors.New("appointment is already cancelled")
	ErrNewOutsideWorkingHours = errors.New("new slot is outside working hours")
	ErrNewSlotBooked          = errors.New("new appointment slot is already booked")
)

type BookingService struct {
	DB *gorm.DB
}

type BookInput struct {
	DoctorID        uuid.UUID
	PatientID       uuid.UUID
	AppointmentDate string
	StartTime       string
}

func NewBookingService(db *gorm.DB) *BookingService {
	return &BookingService{DB: db}
}

func (s *BookingService) Book(ctx context.Context, input BookInput) error {
	dateVal, err := time.Parse("2006-01-02", input.AppointmentDate)
	if err != nil || !dateVal.After(time.Now()) {
		return ErrOutsideWorkingHours
	}
	startTimeParsed, err := time.Parse("15:04", input.StartTime)
	if err != nil {
		return err
	}
	startTime, endTime := appointmentTimes(startTimeParsed)

	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var doctor models.Doctor
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&doctor, "id = ?", input.DoctorID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrDoctorNotFound
			}
			return err
		}
		if err := hasWorkingHours(tx, input.DoctorID, dateVal, startTime, endTime); err != nil {
			return err
		}
		if hasConflict, err := appointmentConflict(tx, input.DoctorID, input.AppointmentDate, startTime, endTime); err != nil {
			return err
		} else if hasConflict {
			return ErrSlotBooked
		}

		return tx.Create(&models.Appointment{
			DoctorID:        input.DoctorID,
			PatientID:       input.PatientID,
			AppointmentDate: dateVal,
			StartTime:       startTime,
			EndTime:         endTime,
			Status:          "booked",
		}).Error
	})
}

func (s *BookingService) Availability(ctx context.Context, doctorID uuid.UUID, dateStr string) ([]string, error) {
	parsedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, err
	}

	var shifts []models.WorkingHours
	if err := s.DB.WithContext(ctx).Where("doctor_id = ? AND day_of_week = ?", doctorID, parsedDate.Weekday()).Find(&shifts).Error; err != nil {
		return nil, err
	}
	var appointments []models.Appointment
	if err := s.DB.WithContext(ctx).Where("doctor_id = ? AND appointment_date = ? AND status = ?", doctorID, dateStr, "booked").Find(&appointments).Error; err != nil {
		return nil, err
	}

	bookedSlots := make(map[string]bool)
	for _, appointment := range appointments {
		if len(appointment.StartTime) >= 5 {
			bookedSlots[appointment.StartTime[:5]] = true
		}
	}

	var availableSlots []string
	for _, shift := range shifts {
		for _, slot := range slotsForShift(shift.StartTime, shift.EndTime) {
			if !bookedSlots[slot] {
				availableSlots = append(availableSlots, slot)
			}
		}
	}
	return availableSlots, nil
}

func slotsForShift(startTime, endTime string) []string {
	curr, err := time.Parse("15:04:05", startTime)
	if err != nil {
		return nil
	}
	limit, err := time.Parse("15:04:05", endTime)
	if err != nil {
		return nil
	}

	var slots []string
	for curr.Add(30*time.Minute).Before(limit) || curr.Add(30*time.Minute).Equal(limit) {
		slots = append(slots, curr.Format("15:04"))
		curr = curr.Add(30 * time.Minute)
	}
	return slots
}

func (s *BookingService) Cancel(ctx context.Context, id uuid.UUID) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var appointment models.Appointment
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&appointment, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrAppointmentNotFound
			}
			return err
		}
		if appointment.Status == "cancelled" {
			return ErrAppointmentCancelled
		}
		return tx.Model(&appointment).Update("status", "cancelled").Error
	})
}

func (s *BookingService) Reschedule(ctx context.Context, id uuid.UUID, dateStr, startStr string) error {
	dateVal, err := time.Parse("2006-01-02", dateStr)
	if err != nil || !dateVal.After(time.Now()) {
		return ErrNewOutsideWorkingHours
	}
	startTimeParsed, err := time.Parse("15:04", startStr)
	if err != nil {
		return err
	}
	startTime, endTime := appointmentTimes(startTimeParsed)

	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var appointment models.Appointment
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&appointment, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrAppointmentNotFound
			}
			return err
		}
		if appointment.Status == "cancelled" {
			return ErrAppointmentCancelled
		}
		if err := tx.Model(&appointment).Update("status", "cancelled").Error; err != nil {
			return err
		}
		if err := hasWorkingHours(tx, appointment.DoctorID, dateVal, startTime, endTime); err != nil {
			if errors.Is(err, ErrOutsideWorkingHours) {
				return ErrNewOutsideWorkingHours
			}
			return err
		}
		if hasConflict, err := appointmentConflict(tx, appointment.DoctorID, dateStr, startTime, endTime); err != nil {
			return err
		} else if hasConflict {
			return ErrNewSlotBooked
		}
		return tx.Create(&models.Appointment{
			DoctorID:        appointment.DoctorID,
			PatientID:       appointment.PatientID,
			AppointmentDate: dateVal,
			StartTime:       startTime,
			EndTime:         endTime,
			Status:          "booked",
		}).Error
	})
}

func appointmentTimes(start time.Time) (string, string) {
	return start.Format("15:04:00"), start.Add(30 * time.Minute).Format("15:04:00")
}

func hasWorkingHours(tx *gorm.DB, doctorID uuid.UUID, date time.Time, startTime, endTime string) error {
	var exists bool
	if err := tx.Model(&models.WorkingHours{}).
		Where("doctor_id = ? AND day_of_week = ? AND start_time <= ? AND end_time >= ?", doctorID, date.Weekday(), startTime, endTime).
		Select("COUNT(*) > 0").Scan(&exists).Error; err != nil {
		return err
	}
	if !exists {
		return ErrOutsideWorkingHours
	}
	return nil
}

func appointmentConflict(tx *gorm.DB, doctorID uuid.UUID, date, startTime, endTime string) (bool, error) {
	var exists bool
	err := tx.Model(&models.Appointment{}).
		Where("doctor_id = ? AND appointment_date = ? AND status = ? AND NOT (? >= end_time OR ? <= start_time)", doctorID, date, "booked", startTime, endTime).
		Select("COUNT(*) > 0").Scan(&exists).Error
	return exists, err
}
