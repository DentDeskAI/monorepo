package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/dentdesk/dentdesk/internal/appointments"
	"github.com/dentdesk/dentdesk/internal/conversations"
	"github.com/dentdesk/dentdesk/internal/scheduler"
)

var (
	ErrInvalidRange  = errors.New("invalid range")
	ErrInvalidStatus = errors.New("invalid status")
)

type SchedulingService struct {
	Appointments  *appointments.Repo
	Conversations *conversations.Repo
	Scheduler     scheduler.Scheduler
}

func NewSchedulingService(apptRepo *appointments.Repo, convRepo *conversations.Repo, sched scheduler.Scheduler) *SchedulingService {
	return &SchedulingService{Appointments: apptRepo, Conversations: convRepo, Scheduler: sched}
}

func (s *SchedulingService) GetSlots(ctx context.Context, clinicID uuid.UUID, from, to *time.Time, specialty string) ([]scheduler.Slot, error) {
	start := time.Now()
	end := time.Now().Add(7 * 24 * time.Hour)
	if from != nil && to != nil {
		start = *from
		end = *to
	}
	slots, err := s.Scheduler.GetFreeSlots(ctx, clinicID, start, end, specialty)
	if err != nil {
		return nil, err
	}
	if slots == nil {
		return []scheduler.Slot{}, nil
	}
	return slots, nil
}

func (s *SchedulingService) CreateAppointment(ctx context.Context, clinicID uuid.UUID, patientID uuid.UUID, doctorID, chairID *uuid.UUID, startsAt, endsAt time.Time, service *string) (*appointments.Appointment, error) {
	if !endsAt.After(startsAt) {
		return nil, ErrInvalidRange
	}
	a := &appointments.Appointment{
		ClinicID:  clinicID,
		PatientID: patientID,
		DoctorID:  doctorID,
		ChairID:   chairID,
		StartsAt:  startsAt,
		EndsAt:    endsAt,
		Service:   service,
		Status:    "scheduled",
		Source:    "operator",
	}
	return s.Appointments.Create(ctx, a)
}

func (s *SchedulingService) GetAppointment(ctx context.Context, id uuid.UUID) (*appointments.Appointment, error) {
	return s.Appointments.Get(ctx, id)
}

func (s *SchedulingService) UpdateAppointmentStatus(ctx context.Context, id uuid.UUID, status string) error {
	valid := map[string]bool{
		"scheduled": true, "confirmed": true, "cancelled": true, "completed": true, "no_show": true,
	}
	if !valid[status] {
		return ErrInvalidStatus
	}
	return s.Appointments.SetStatus(ctx, id, status)
}

func (s *SchedulingService) GetConversation(ctx context.Context, id uuid.UUID) (*conversations.Conversation, error) {
	return s.Conversations.Get(ctx, id)
}

func (s *SchedulingService) CloseConversation(ctx context.Context, id uuid.UUID) error {
	return s.Conversations.SetStatus(ctx, id, "closed")
}
