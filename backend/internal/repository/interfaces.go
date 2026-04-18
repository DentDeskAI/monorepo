// Package repository defines repository interfaces and GORM implementations.
package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/dentdesk/backend/internal/domain"
)

// ─── Interfaces ──────────────────────────────────────────────────────────────
// Defining interfaces here enables easy mocking in unit tests.

type ClinicRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Clinic, error)
	FindByPhoneNumberID(ctx context.Context, phoneNumberID string) (*domain.Clinic, error)
	Create(ctx context.Context, clinic *domain.Clinic) error
}

type PatientRepository interface {
	FindByID(ctx context.Context, clinicID, patientID uuid.UUID) (*domain.Patient, error)
	FindByPhone(ctx context.Context, clinicID uuid.UUID, phone string) (*domain.Patient, error)
	List(ctx context.Context, clinicID uuid.UUID, page, pageSize int) ([]domain.Patient, int64, error)
	Create(ctx context.Context, patient *domain.Patient) error
	Update(ctx context.Context, patient *domain.Patient) error
	Delete(ctx context.Context, clinicID, patientID uuid.UUID) error
}

type AppointmentRepository interface {
	FindByID(ctx context.Context, clinicID, appointmentID uuid.UUID) (*domain.Appointment, error)
	ListByClinic(ctx context.Context, clinicID uuid.UUID, filter AppointmentFilter) ([]domain.Appointment, error)
	Create(ctx context.Context, appt *domain.Appointment) error
	Update(ctx context.Context, appt *domain.Appointment) error
	Delete(ctx context.Context, clinicID, appointmentID uuid.UUID) error
}

type MessageLogRepository interface {
	Create(ctx context.Context, msg *domain.MessageLog) error
	ListByPatient(ctx context.Context, clinicID, patientID uuid.UUID, page, pageSize int) ([]domain.MessageLog, int64, error)
	ListByClinic(ctx context.Context, clinicID uuid.UUID, page, pageSize int) ([]domain.MessageLog, int64, error)
	UpdateStatus(ctx context.Context, wamid string, status domain.MessageStatus) error
}

// AppointmentFilter carries query parameters for appointment listing.
type AppointmentFilter struct {
	DoctorID  *uuid.UUID
	PatientID *uuid.UUID
	DateFrom  *string // ISO-8601 date
	DateTo    *string
	Status    *domain.AppointmentStatus
	Page      int
	PageSize  int
}
