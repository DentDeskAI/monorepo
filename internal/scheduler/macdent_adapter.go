package scheduler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/dentdesk/dentdesk/internal/macdent"
)

// MacDentAdapter implements Scheduler using the MacDent external API.
// It is intentionally thin: DB key lookup + type mapping only.
// All HTTP logic lives in the macdent package.
type MacDentAdapter struct {
	db         *sqlx.DB
	httpClient *http.Client
}

func NewMacDentAdapter(db *sqlx.DB) *MacDentAdapter {
	return &MacDentAdapter{
		db:         db,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// clientFor fetches the clinic's MacDent API key and returns a ready-to-use client.
func (a *MacDentAdapter) clientFor(ctx context.Context, clinicID uuid.UUID) (*macdent.Client, error) {
	var apiKey string
	if err := a.db.GetContext(ctx, &apiKey,
		`SELECT macdent_api_key FROM clinics WHERE id = $1`, clinicID); err != nil {
		return nil, fmt.Errorf("macdent: get api key: %w", err)
	}
	return macdent.NewWithHTTP(apiKey, a.httpClient), nil
}

// ── doctor ────────────────────────────────────────────────────────────────────

func (a *MacDentAdapter) ListDoctors(ctx context.Context, clinicID uuid.UUID) ([]Doctor, error) {
	client, err := a.clientFor(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	mds, err := client.ListDoctors(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Doctor, 0, len(mds))
	for _, d := range mds {
		specs := make([]string, 0, len(d.Specialnosti))
		for _, s := range d.Specialnosti {
			specs = append(specs, s.Name)
		}
		out = append(out, Doctor{
			ID:          fmt.Sprint(d.ID),
			Name:        d.Name,
			Specialties: specs,
		})
	}
	return out, nil
}

func (a *MacDentAdapter) GetDoctor(ctx context.Context, clinicID uuid.UUID, id string) (*Doctor, error) {
	list, err := a.ListDoctors(ctx, clinicID)
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

// ── patient ───────────────────────────────────────────────────────────────────

func (a *MacDentAdapter) ListPatients(ctx context.Context, clinicID uuid.UUID) ([]Patient, error) {
	client, err := a.clientFor(ctx, clinicID)
	if err != nil {
		return nil, err
	}

	resp, err := client.ListPatients(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]Patient, 0, len(resp.Patients))
	for _, p := range resp.Patients {
		out = append(out, toSchedulerPatient(Patient(p)))
	}

	return out, nil
}

func (a *MacDentAdapter) GetPatient(ctx context.Context, clinicID uuid.UUID, id int) (*Patient, error) {
	list, err := a.ListPatients(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	for _, p := range list {
		if p.ID == id {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("doctor %d not found", id)
}

// ── slots ─────────────────────────────────────────────────────────────────────

func (a *MacDentAdapter) GetFreeSlots(
	ctx context.Context, clinicID uuid.UUID, from, to time.Time, specialty string,
) ([]Slot, error) {
	client, err := a.clientFor(ctx, clinicID)
	if err != nil {
		return nil, err
	}

	doctors, err := client.ListDoctors(ctx)
	if err != nil {
		return nil, err
	}

	var slots []Slot
	for _, doc := range doctors {
		schedules, err := client.GetFreeTime(ctx, doc.ID, from, to)
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

// ── zapis ─────────────────────────────────────────────────────────────────────

func (a *MacDentAdapter) CreateAppointment(ctx context.Context, req BookRequest) (*BookResult, error) {
	client, err := a.clientFor(ctx, req.ClinicID)
	if err != nil {
		return nil, err
	}
	// TODO: resolve MacDent integer doctor ID from req.DoctorID (UUID → external_id lookup)
	z, err := client.AddZapis(ctx, 0, req.StartsAt, req.EndsAt, req.Service)
	if err != nil {
		return nil, err
	}
	ext := fmt.Sprint(z.ID)
	return &BookResult{AppointmentID: uuid.New(), ExternalID: &ext}, nil
}

func (a *MacDentAdapter) CancelAppointment(ctx context.Context, appointmentID uuid.UUID) error {
	// TODO: resolve MacDent integer zapis ID from appointmentID (external_id lookup) then call SetStatus(id, 2)
	_ = appointmentID
	return nil
}
