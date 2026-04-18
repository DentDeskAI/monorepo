package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/dentdesk/backend/internal/domain"
	"github.com/dentdesk/backend/internal/repository"
)

type AppointmentService struct {
	repo repository.AppointmentRepository
}

func NewAppointmentService(repo repository.AppointmentRepository) *AppointmentService {
	return &AppointmentService{repo: repo}
}

// ─── DTOs ─────────────────────────────────────────────────────────────────────

type CreateAppointmentRequest struct {
	PatientID string `json:"patient_id" binding:"required,uuid"`
	DoctorID  string `json:"doctor_id" binding:"required,uuid"`
	StartsAt  string `json:"starts_at" binding:"required"` // RFC3339
	EndsAt    string `json:"ends_at" binding:"required"`
	Title     string `json:"title"`
	Notes     string `json:"notes"`
}

type UpdateAppointmentRequest struct {
	StartsAt *string                    `json:"starts_at"`
	EndsAt   *string                    `json:"ends_at"`
	Status   *domain.AppointmentStatus  `json:"status"`
	Title    *string                    `json:"title"`
	Notes    *string                    `json:"notes"`
}

// ─── Methods ──────────────────────────────────────────────────────────────────

func (s *AppointmentService) List(ctx context.Context, clinicID uuid.UUID, filter repository.AppointmentFilter) ([]domain.Appointment, error) {
	return s.repo.ListByClinic(ctx, clinicID, filter)
}

func (s *AppointmentService) GetByID(ctx context.Context, clinicID, id uuid.UUID) (*domain.Appointment, error) {
	return s.repo.FindByID(ctx, clinicID, id)
}

func (s *AppointmentService) Create(ctx context.Context, clinicID uuid.UUID, req CreateAppointmentRequest) (*domain.Appointment, error) {
	startsAt, err := time.Parse(time.RFC3339, req.StartsAt)
	if err != nil {
		return nil, fmt.Errorf("invalid starts_at format, expected RFC3339")
	}
	endsAt, err := time.Parse(time.RFC3339, req.EndsAt)
	if err != nil {
		return nil, fmt.Errorf("invalid ends_at format, expected RFC3339")
	}
	if !endsAt.After(startsAt) {
		return nil, fmt.Errorf("ends_at must be after starts_at")
	}

	patientID := uuid.MustParse(req.PatientID)
	doctorID  := uuid.MustParse(req.DoctorID)

	appt := &domain.Appointment{
		ClinicID:  clinicID,
		PatientID: patientID,
		DoctorID:  doctorID,
		StartsAt:  startsAt,
		EndsAt:    endsAt,
		Status:    domain.AppointmentStatusScheduled,
		Title:     req.Title,
		Notes:     req.Notes,
	}

	if err := s.repo.Create(ctx, appt); err != nil {
		return nil, err
	}
	return appt, nil
}

func (s *AppointmentService) Update(ctx context.Context, clinicID, id uuid.UUID, req UpdateAppointmentRequest) (*domain.Appointment, error) {
	appt, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, err
	}

	if req.StartsAt != nil {
		t, err := time.Parse(time.RFC3339, *req.StartsAt)
		if err != nil {
			return nil, fmt.Errorf("invalid starts_at format")
		}
		appt.StartsAt = t
	}
	if req.EndsAt != nil {
		t, err := time.Parse(time.RFC3339, *req.EndsAt)
		if err != nil {
			return nil, fmt.Errorf("invalid ends_at format")
		}
		appt.EndsAt = t
	}
	if req.Status != nil {
		appt.Status = *req.Status
	}
	if req.Title != nil {
		appt.Title = *req.Title
	}
	if req.Notes != nil {
		appt.Notes = *req.Notes
	}

	if err := s.repo.Update(ctx, appt); err != nil {
		return nil, err
	}
	return appt, nil
}

func (s *AppointmentService) Delete(ctx context.Context, clinicID, id uuid.UUID) error {
	return s.repo.Delete(ctx, clinicID, id)
}
