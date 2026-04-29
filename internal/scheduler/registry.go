package scheduler

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/dentdesk/dentdesk/internal/appointments"
	"github.com/dentdesk/dentdesk/internal/clinics"
	"github.com/dentdesk/dentdesk/internal/doctors"
	"github.com/dentdesk/dentdesk/internal/patients"
)

// Registry is the top-level Scheduler implementation.
// It dispatches each call to the right backend (MacDent or Local) based on
// the clinic's scheduler_type, caching the per-clinic Scheduler instance.
type Registry struct {
	db        *sqlx.DB
	httpCl    *http.Client
	clinicsR  *clinics.Repo
	doctorsR  *doctors.Repo
	patientsR *patients.Repo
	apptR     *appointments.Repo

	mu    sync.RWMutex
	cache map[uuid.UUID]Scheduler
}

// NewRegistry creates a Registry. Pass the same repos already wired in main.
func NewRegistry(
	db *sqlx.DB,
	httpCl *http.Client,
	clinicsR *clinics.Repo,
	doctorsR *doctors.Repo,
	patientsR *patients.Repo,
	apptR *appointments.Repo,
) *Registry {
	return &Registry{
		db:        db,
		httpCl:    httpCl,
		clinicsR:  clinicsR,
		doctorsR:  doctorsR,
		patientsR: patientsR,
		apptR:     apptR,
		cache:     map[uuid.UUID]Scheduler{},
	}
}

// Invalidate removes a cached Scheduler so it is rebuilt on next use.
// Call after updating a clinic's scheduler_type.
func (r *Registry) Invalidate(clinicID uuid.UUID) {
	r.mu.Lock()
	delete(r.cache, clinicID)
	r.mu.Unlock()
}

// forClinic returns the Scheduler for the clinic, building it if not cached.
func (r *Registry) forClinic(ctx context.Context, clinicID uuid.UUID) (Scheduler, error) {
	r.mu.RLock()
	s, ok := r.cache[clinicID]
	r.mu.RUnlock()
	if ok {
		return s, nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	// Double-check after acquiring write lock.
	if s, ok = r.cache[clinicID]; ok {
		return s, nil
	}

	clinic, err := r.clinicsR.Get(ctx, clinicID)
	if err != nil {
		return nil, fmt.Errorf("clinic %s not found: %w", clinicID, err)
	}

	switch clinic.SchedulerType {
	case "local", "mock":
		s = &LocalScheduler{
			doctorsR:  r.doctorsR,
			patientsR: r.patientsR,
			apptR:     r.apptR,
			clinicsR:  r.clinicsR,
		}
	case "macdent":
		s = &Service{db: r.db, http: r.httpCl}
	default:
		return nil, fmt.Errorf("unknown scheduler_type %q for clinic %s", clinic.SchedulerType, clinicID)
	}

	r.cache[clinicID] = s
	return s, nil
}

// ── Scheduler interface delegation ───────────────────────────────────────────

func (r *Registry) ListDoctors(ctx context.Context, clinicID uuid.UUID) ([]Doctor, error) {
	s, err := r.forClinic(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	return s.ListDoctors(ctx, clinicID)
}

func (r *Registry) GetDoctor(ctx context.Context, clinicID uuid.UUID, id string) (*Doctor, error) {
	s, err := r.forClinic(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	return s.GetDoctor(ctx, clinicID, id)
}

func (r *Registry) ListPatients(ctx context.Context, clinicID uuid.UUID) ([]Patient, error) {
	s, err := r.forClinic(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	return s.ListPatients(ctx, clinicID)
}

func (r *Registry) GetPatient(ctx context.Context, clinicID uuid.UUID, id int) (*Patient, error) {
	s, err := r.forClinic(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	return s.GetPatient(ctx, clinicID, id)
}

func (r *Registry) CreatePatient(ctx context.Context, clinicID uuid.UUID, p CreatePatientParams) (*Patient, error) {
	s, err := r.forClinic(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	return s.CreatePatient(ctx, clinicID, p)
}

func (r *Registry) GetClinic(ctx context.Context, clinicID uuid.UUID) (*Stomatology, error) {
	s, err := r.forClinic(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	return s.GetClinic(ctx, clinicID)
}

func (r *Registry) ListAppointments(ctx context.Context, clinicID uuid.UUID, from, to time.Time) (*AppointmentsResponse, error) {
	s, err := r.forClinic(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	return s.ListAppointments(ctx, clinicID, from, to)
}

func (r *Registry) GetAppointmentByID(ctx context.Context, clinicID uuid.UUID, id int) (*AppointmentDetail, error) {
	s, err := r.forClinic(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	return s.GetAppointmentByID(ctx, clinicID, id)
}

func (r *Registry) CreateAppointment(ctx context.Context, req BookRequest) (*BookResult, error) {
	s, err := r.forClinic(ctx, req.ClinicID)
	if err != nil {
		return nil, err
	}
	return s.CreateAppointment(ctx, req)
}

func (r *Registry) CreateScheduleAppointment(ctx context.Context, clinicID uuid.UUID, p ScheduleAppointmentParams) (*ScheduleAppointmentResult, error) {
	s, err := r.forClinic(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	return s.CreateScheduleAppointment(ctx, clinicID, p)
}

func (r *Registry) UpdateAppointment(ctx context.Context, clinicID uuid.UUID, id int, p UpdateAppointmentParams) error {
	s, err := r.forClinic(ctx, clinicID)
	if err != nil {
		return err
	}
	return s.UpdateAppointment(ctx, clinicID, id, p)
}

func (r *Registry) RemoveAppointment(ctx context.Context, clinicID uuid.UUID, id int) error {
	s, err := r.forClinic(ctx, clinicID)
	if err != nil {
		return err
	}
	return s.RemoveAppointment(ctx, clinicID, id)
}

func (r *Registry) SetAppointmentStatus(ctx context.Context, clinicID uuid.UUID, id, status int) error {
	s, err := r.forClinic(ctx, clinicID)
	if err != nil {
		return err
	}
	return s.SetAppointmentStatus(ctx, clinicID, id, status)
}

func (r *Registry) SendAppointmentRequest(ctx context.Context, clinicID uuid.UUID, p AppointmentRequestParams) (*AppointmentRequestResult, error) {
	s, err := r.forClinic(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	return s.SendAppointmentRequest(ctx, clinicID, p)
}

func (r *Registry) GetFreeSlots(ctx context.Context, clinicID uuid.UUID, from, to time.Time, specialty string) ([]Slot, error) {
	s, err := r.forClinic(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	return s.GetFreeSlots(ctx, clinicID, from, to, specialty)
}

func (r *Registry) GetRevenue(ctx context.Context, clinicID uuid.UUID, from, to time.Time) ([]RevenueRecord, error) {
	s, err := r.forClinic(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	return s.GetRevenue(ctx, clinicID, from, to)
}

func (r *Registry) GetHistory(ctx context.Context, clinicID uuid.UUID, from, to time.Time) (*AppointmentsResponse, error) {
	s, err := r.forClinic(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	return s.GetHistory(ctx, clinicID, from, to)
}
