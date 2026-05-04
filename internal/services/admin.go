package services

import (
	"context"

	"github.com/google/uuid"

	"github.com/dentdesk/dentdesk/internal/auth"
	"github.com/dentdesk/dentdesk/internal/store"
)

type AdminService struct {
	Auth    *auth.Service
	Clinics *store.ClinicRepo
}

func NewAdminService(authSvc *auth.Service, clinicsRepo *store.ClinicRepo) *AdminService {
	return &AdminService{Auth: authSvc, Clinics: clinicsRepo}
}

func (s *AdminService) Register(ctx context.Context, clinicName, timezone, ownerName, email, password string) (*store.Clinic, *auth.User, string, error) {
	clinic, err := s.Clinics.Create(ctx, clinicName, timezone, "local")
	if err != nil {
		return nil, nil, "", err
	}
	user, err := s.Auth.CreateUser(ctx, clinic.ID, email, password, "owner", ownerName)
	if err != nil {
		return nil, nil, "", err
	}
	token, _, err := s.Auth.Login(ctx, email, password)
	if err != nil {
		return nil, nil, "", err
	}
	return clinic, user, token, nil
}

func (s *AdminService) GetClinic(ctx context.Context, clinicID uuid.UUID) (*store.Clinic, error) {
	return s.Clinics.Get(ctx, clinicID)
}

func (s *AdminService) UpdateClinic(ctx context.Context, clinicID uuid.UUID, reqName, reqTimezone, reqWorkingHours, reqSchedulerType string, reqSlotDuration int) (*store.Clinic, error) {
	dur := reqSlotDuration
	if dur == 0 {
		dur = 30
	}
	sched := reqSchedulerType
	if sched == "" {
		sched = "local"
	}
	wh := reqWorkingHours
	if wh == "" {
		wh = `{"mon":["09:00","20:00"],"tue":["09:00","20:00"],"wed":["09:00","20:00"],"thu":["09:00","20:00"],"fri":["09:00","20:00"],"sat":["10:00","18:00"],"sun":null}`
	}

	if err := s.Clinics.Update(ctx, clinicID, store.ClinicUpdateFields{
		Name: reqName, Timezone: reqTimezone,
		WorkingHours: []byte(wh), SlotDurationMin: dur, SchedulerType: sched,
	}); err != nil {
		return nil, err
	}
	return s.Clinics.Get(ctx, clinicID)
}

func (s *AdminService) ListUsers(ctx context.Context, clinicID uuid.UUID) ([]auth.User, error) {
	return s.Auth.ListUsers(ctx, clinicID)
}

func (s *AdminService) CreateUser(ctx context.Context, clinicID uuid.UUID, email, password, role, name string) (*auth.User, error) {
	return s.Auth.CreateUser(ctx, clinicID, email, password, role, name)
}

func (s *AdminService) GetUser(ctx context.Context, id uuid.UUID) (*auth.User, error) {
	return s.Auth.GetUser(ctx, id)
}

func (s *AdminService) UpdateUser(ctx context.Context, id uuid.UUID, name, role string) error {
	return s.Auth.UpdateUser(ctx, id, name, role)
}

func (s *AdminService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.Auth.DeleteUser(ctx, id)
}

func (s *AdminService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	return s.Auth.ChangePassword(ctx, userID, oldPassword, newPassword)
}
