// Package scheduler is the scheduling layer of DentDesk.
//
// Phase 1 note: today this layer reads through the MacDent integration directly
// (read-through). Long-term DentDesk will be local-first — handlers will read
// from internal repositories and integrations will sync data in the background.
// When that time comes, replace the calls inside Service methods with repo
// reads and move the MacDent calls to a sync worker. The public method
// signatures should not need to change.
package scheduler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/dentdesk/dentdesk/internal/integrations/macdent"
)

// ── domain types ─────────────────────────────────────────────────────────────

type Slot struct {
	StartsAt time.Time  `json:"starts_at"`
	EndsAt   time.Time  `json:"ends_at"`
	DoctorID uuid.UUID  `json:"doctor_id"`
	Doctor   string     `json:"doctor"`
	ChairID  *uuid.UUID `json:"chair_id,omitempty"`
}

type Doctor struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Specialties []string `json:"specialties,omitempty"`
}

type Patient struct {
	Name      string  `json:"name"`
	Gender    *string `json:"gender"`
	ID        int     `json:"id"`
	IIN       *string `json:"iin"`
	Number    string  `json:"number"`
	Phone     *string `json:"phone"`
	Birth     *string `json:"birth"`
	IsChild   bool    `json:"isChild"`
	Comment   string  `json:"comment"`
	WhereKnow string  `json:"whereKnow"`
}

type Stomatology struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type BookRequest struct {
	ClinicID  uuid.UUID
	PatientID uuid.UUID
	DoctorID  uuid.UUID
	ChairID   *uuid.UUID
	StartsAt  time.Time
	EndsAt    time.Time
	Service   string
}

type BookResult struct {
	AppointmentID uuid.UUID
	ExternalID    *string
}

// ── service ──────────────────────────────────────────────────────────────────

// Service is DentDesk's scheduling service. It's a single concrete type — no
// adapter interface — because there is exactly one source of truth today
// (MacDent). When a second integration (e.g. IDENT) appears, this is the place
// to introduce an interface designed against two real implementations.
type Service struct {
	db   *sqlx.DB
	http *http.Client
}

func NewService(db *sqlx.DB) *Service {
	return &Service{
		db:   db,
		http: &http.Client{Timeout: 15 * time.Second},
	}
}

func (s *Service) macdent(ctx context.Context, clinicID uuid.UUID) (*macdent.Client, error) {
	return macdent.ClientFor(ctx, s.db, s.http, clinicID)
}

// ── doctors ──────────────────────────────────────────────────────────────────

func (s *Service) ListDoctors(ctx context.Context, clinicID uuid.UUID) ([]Doctor, error) {
	c, err := s.macdent(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	mds, err := c.ListDoctors(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Doctor, 0, len(mds))
	for _, d := range mds {
		specs := make([]string, 0, len(d.Specialnosti))
		for _, sp := range d.Specialnosti {
			specs = append(specs, sp.Name)
		}
		out = append(out, Doctor{
			ID:          fmt.Sprint(d.ID),
			Name:        d.Name,
			Specialties: specs,
		})
	}
	return out, nil
}

func (s *Service) GetDoctor(ctx context.Context, clinicID uuid.UUID, id string) (*Doctor, error) {
	list, err := s.ListDoctors(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	for _, d := range list {
		if d.ID == id {
			return &d, nil
		}
	}
	return nil, fmt.Errorf("doctor %s not found", id)
}

// ── patients ─────────────────────────────────────────────────────────────────

func (s *Service) ListPatients(ctx context.Context, clinicID uuid.UUID) ([]Patient, error) {
	c, err := s.macdent(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	resp, err := c.ListPatients(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Patient, 0, len(resp.Patients))
	for _, p := range resp.Patients {
		out = append(out, Patient{
			Name:      p.Name,
			Gender:    p.Gender,
			ID:        p.ID,
			IIN:       p.IIN,
			Number:    p.Number,
			Phone:     p.Phone,
			Birth:     p.Birth,
			IsChild:   p.IsChild,
			Comment:   p.Comment,
			WhereKnow: p.WhereKnow,
		})
	}
	return out, nil
}

func (s *Service) GetPatient(ctx context.Context, clinicID uuid.UUID, id int) (*Patient, error) {
	list, err := s.ListPatients(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	for _, p := range list {
		if p.ID == id {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("patient %d not found", id)
}

// ── clinic ───────────────────────────────────────────────────────────────────

func (s *Service) GetClinic(ctx context.Context, clinicID uuid.UUID) (*Stomatology, error) {
	c, err := s.macdent(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	stom, err := c.GetClinic(ctx)
	if err != nil {
		return nil, err
	}
	return &Stomatology{ID: stom.ID, Name: stom.Name}, nil
}

// ── slots ────────────────────────────────────────────────────────────────────

func (s *Service) GetFreeSlots(
	ctx context.Context, clinicID uuid.UUID, from, to time.Time, specialty string,
) ([]Slot, error) {
	c, err := s.macdent(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	doctors, err := c.ListDoctors(ctx)
	if err != nil {
		return nil, err
	}
	var slots []Slot
	for _, doc := range doctors {
		schedules, err := c.GetFreeTime(ctx, doc.ID, from, to)
		if err != nil {
			continue
		}
		for _, iv := range schedules {
			start, err1 := time.Parse(macdent.DateLayout, iv.From)
			end, err2 := time.Parse(macdent.DateLayout, iv.To)
			if err1 != nil || err2 != nil {
				continue
			}
			for cur := start; !cur.Add(30 * time.Minute).After(end); cur = cur.Add(30 * time.Minute) {
				slots = append(slots, Slot{
					StartsAt: cur,
					EndsAt:   cur.Add(30 * time.Minute),
					DoctorID: uuid.Nil,
					Doctor:   doc.Name,
				})
			}
		}
	}
	return slots, nil
}

// ── appointments (zapis) ─────────────────────────────────────────────────────

func (s *Service) CreateAppointment(ctx context.Context, req BookRequest) (*BookResult, error) {
	c, err := s.macdent(ctx, req.ClinicID)
	if err != nil {
		return nil, err
	}
	// TODO: resolve MacDent integer doctor ID from req.DoctorID (UUID → external_id lookup)
	z, err := c.AddZapis(ctx, 0, req.StartsAt, req.EndsAt, req.Service)
	if err != nil {
		return nil, err
	}
	ext := fmt.Sprint(z.ID)
	return &BookResult{AppointmentID: uuid.New(), ExternalID: &ext}, nil
}

func (s *Service) CancelAppointment(ctx context.Context, appointmentID uuid.UUID) error {
	// TODO: resolve MacDent integer zapis ID from appointmentID (external_id lookup) then call SetStatus(id, 2)
	_ = appointmentID
	return nil
}
