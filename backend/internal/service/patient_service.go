package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/dentdesk/backend/internal/domain"
	"github.com/dentdesk/backend/internal/repository"
)

type PatientService struct {
	repo repository.PatientRepository
}

func NewPatientService(repo repository.PatientRepository) *PatientService {
	return &PatientService{repo: repo}
}

// ─── DTOs ─────────────────────────────────────────────────────────────────────

type CreatePatientRequest struct {
	FirstName   string  `json:"first_name" binding:"required"`
	LastName    string  `json:"last_name" binding:"required"`
	Phone       string  `json:"phone" binding:"required"`
	Email       string  `json:"email"`
	DateOfBirth *string `json:"date_of_birth"` // "YYYY-MM-DD"
	Gender      string  `json:"gender"`
	Notes       string  `json:"notes"`
}

type UpdatePatientRequest struct {
	FirstName   *string `json:"first_name"`
	LastName    *string `json:"last_name"`
	Phone       *string `json:"phone"`
	Email       *string `json:"email"`
	DateOfBirth *string `json:"date_of_birth"`
	Gender      *string `json:"gender"`
	Notes       *string `json:"notes"`
}

// ─── Methods ──────────────────────────────────────────────────────────────────

func (s *PatientService) List(ctx context.Context, clinicID uuid.UUID, page, pageSize int) ([]domain.Patient, int64, error) {
	return s.repo.List(ctx, clinicID, page, pageSize)
}

func (s *PatientService) GetByID(ctx context.Context, clinicID, patientID uuid.UUID) (*domain.Patient, error) {
	return s.repo.FindByID(ctx, clinicID, patientID)
}

func (s *PatientService) Create(ctx context.Context, clinicID uuid.UUID, req CreatePatientRequest) (*domain.Patient, error) {
	patient := &domain.Patient{
		ClinicID:  clinicID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		Email:     req.Email,
		Gender:    req.Gender,
		Notes:     req.Notes,
	}

	if req.DateOfBirth != nil {
		dob, err := time.Parse("2006-01-02", *req.DateOfBirth)
		if err != nil {
			return nil, fmt.Errorf("invalid date_of_birth format, expected YYYY-MM-DD")
		}
		patient.DateOfBirth = &dob
	}

	if err := s.repo.Create(ctx, patient); err != nil {
		return nil, err
	}
	return patient, nil
}

func (s *PatientService) Update(ctx context.Context, clinicID, patientID uuid.UUID, req UpdatePatientRequest) (*domain.Patient, error) {
	patient, err := s.repo.FindByID(ctx, clinicID, patientID)
	if err != nil {
		return nil, err
	}

	if req.FirstName != nil {
		patient.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		patient.LastName = *req.LastName
	}
	if req.Phone != nil {
		patient.Phone = *req.Phone
	}
	if req.Email != nil {
		patient.Email = *req.Email
	}
	if req.Gender != nil {
		patient.Gender = *req.Gender
	}
	if req.Notes != nil {
		patient.Notes = *req.Notes
	}
	if req.DateOfBirth != nil {
		dob, err := time.Parse("2006-01-02", *req.DateOfBirth)
		if err != nil {
			return nil, fmt.Errorf("invalid date_of_birth format")
		}
		patient.DateOfBirth = &dob
	}

	if err := s.repo.Update(ctx, patient); err != nil {
		return nil, err
	}
	return patient, nil
}

func (s *PatientService) Delete(ctx context.Context, clinicID, patientID uuid.UUID) error {
	return s.repo.Delete(ctx, clinicID, patientID)
}
