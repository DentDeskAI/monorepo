package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/dentdesk/dentdesk/internal/scheduling"
	"github.com/dentdesk/dentdesk/internal/store"
)

var (
	ErrInvalidRange  = errors.New("invalid range")
	ErrInvalidStatus = errors.New("invalid status")
)

// SchedulingService holds business logic that lives on top of the scheduling
// layer: appointment validation, conversation lifecycle, etc. Pure passthrough
// reads (list/get) are the responsibility of Scheduler and are called
// directly from handlers.
type SchedulingService struct {
	Appointments  *store.AppointmentRepo
	Conversations *store.ConversationRepo
	Sched         scheduling.Scheduler
	Doctors       *store.DoctorRepo
}

func NewSchedulingService(
	apptRepo *store.AppointmentRepo,
	convRepo *store.ConversationRepo,
	sched scheduling.Scheduler,
	doctorsRepo *store.DoctorRepo,
) *SchedulingService {
	return &SchedulingService{
		Appointments:  apptRepo,
		Conversations: convRepo,
		Sched:         sched,
		Doctors:       doctorsRepo,
	}
}

// SyncDoctors pulls the doctor list from the scheduler and upserts each entry
// into the local doctors table (keyed by external_id). For MacDent clinics
// this syncs remote doctors to local. For local/mock clinics it is a no-op
// because the scheduler returns doctors already in the local DB.
func (s *SchedulingService) SyncDoctors(ctx context.Context, clinicID uuid.UUID) (int, error) {
	list, err := s.Sched.ListDoctors(ctx, clinicID)
	if err != nil {
		return 0, err
	}
	for _, d := range list {
		var spec *string
		if len(d.Specialties) > 0 {
			spec = &d.Specialties[0]
		}
		if err := s.Doctors.Upsert(ctx, clinicID, d.Name, spec, d.ID); err != nil {
			return 0, err
		}
	}
	return len(list), nil
}

// GetSlots normalises the time range and forwards to the scheduling.
func (s *SchedulingService) GetSlots(ctx context.Context, clinicID uuid.UUID, from, to *time.Time, specialty string) ([]scheduling.Slot, error) {
	start := time.Now()
	end := time.Now().Add(7 * 24 * time.Hour)
	if from != nil && to != nil {
		start = *from
		end = *to
	}
	slots, err := s.Sched.GetFreeSlots(ctx, clinicID, start, end, specialty)
	if err != nil {
		return nil, err
	}
	if slots == nil {
		return []scheduling.Slot{}, nil
	}
	return slots, nil
}

// CreateAppointment validates the time range and persists the appointment in
// our local DB (the canonical record). External-system push is a separate
// concern handled by the integration layer.
func (s *SchedulingService) CreateAppointment(ctx context.Context, clinicID uuid.UUID, patientID uuid.UUID, doctorID, chairID *uuid.UUID, startsAt, endsAt time.Time, service *string) (*store.Appointment, error) {
	if !endsAt.After(startsAt) {
		return nil, ErrInvalidRange
	}
	a := &store.Appointment{
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

func (s *SchedulingService) GetAppointment(ctx context.Context, id uuid.UUID) (*store.Appointment, error) {
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

func (s *SchedulingService) GetConversation(ctx context.Context, id uuid.UUID) (*store.Conversation, error) {
	return s.Conversations.Get(ctx, id)
}

func (s *SchedulingService) CloseConversation(ctx context.Context, id uuid.UUID) error {
	return s.Conversations.SetStatus(ctx, id, "closed")
}
