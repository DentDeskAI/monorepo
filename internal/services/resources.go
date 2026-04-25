package services

import (
	"context"

	"github.com/google/uuid"

	"github.com/dentdesk/dentdesk/internal/chairs"
	"github.com/dentdesk/dentdesk/internal/doctors"
	"github.com/dentdesk/dentdesk/internal/patients"
)

type ResourceService struct {
	Doctors  *doctors.Repo
	Chairs   *chairs.Repo
	Patients *patients.Repo
}

func NewResourceService(doctorsRepo *doctors.Repo, chairsRepo *chairs.Repo, patientsRepo *patients.Repo) *ResourceService {
	return &ResourceService{Doctors: doctorsRepo, Chairs: chairsRepo, Patients: patientsRepo}
}

func (s *ResourceService) CreateDoctor(ctx context.Context, clinicID uuid.UUID, name string, specialty, externalID *string) (*doctors.Doctor, error) {
	return s.Doctors.Create(ctx, clinicID, name, specialty, externalID)
}

func (s *ResourceService) GetDoctor(ctx context.Context, id uuid.UUID) (*doctors.Doctor, error) {
	return s.Doctors.Get(ctx, id)
}

func (s *ResourceService) UpdateDoctor(ctx context.Context, id uuid.UUID, name string, specialty *string, active bool) error {
	return s.Doctors.Update(ctx, id, name, specialty, active)
}

func (s *ResourceService) DeactivateDoctor(ctx context.Context, id uuid.UUID) error {
	return s.Doctors.Deactivate(ctx, id)
}

func (s *ResourceService) ListChairs(ctx context.Context, clinicID uuid.UUID) ([]chairs.Chair, error) {
	return s.Chairs.List(ctx, clinicID)
}

func (s *ResourceService) CreateChair(ctx context.Context, clinicID uuid.UUID, name string, externalID *string) (*chairs.Chair, error) {
	return s.Chairs.Create(ctx, clinicID, name, externalID)
}

func (s *ResourceService) UpdateChair(ctx context.Context, id uuid.UUID, name string) error {
	return s.Chairs.Update(ctx, id, name)
}

func (s *ResourceService) DeactivateChair(ctx context.Context, id uuid.UUID) error {
	return s.Chairs.Deactivate(ctx, id)
}

func (s *ResourceService) CreatePatient(ctx context.Context, clinicID uuid.UUID, phone, language string, name, externalID *string) (*patients.Patient, error) {
	lang := language
	if lang == "" {
		lang = "ru"
	}
	return s.Patients.Create(ctx, clinicID, phone, lang, name, externalID)
}

func (s *ResourceService) GetPatient(ctx context.Context, id uuid.UUID) (*patients.Patient, error) {
	return s.Patients.Get(ctx, id)
}

func (s *ResourceService) UpdatePatient(ctx context.Context, id uuid.UUID, name *string, language string, externalID *string) error {
	lang := language
	if lang == "" {
		lang = "ru"
	}
	return s.Patients.Update(ctx, id, name, lang, externalID)
}
