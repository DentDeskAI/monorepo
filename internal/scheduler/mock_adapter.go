package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// MockAdapter — для демо-режима. Генерирует красивые слоты + ведёт учёт в памяти.
// Записи всё равно сохраняет в appointments (чтобы CRM их видел).
type MockAdapter struct {
	db   *sqlx.DB
	mu   sync.Mutex
	held map[string]bool // "doctorID|starts" => true
}

func NewMockAdapter(db *sqlx.DB) *MockAdapter {
	return &MockAdapter{db: db, held: map[string]bool{}}
}

func (a *MockAdapter) ListDoctors(ctx context.Context, clinicID uuid.UUID) ([]Doctor, error) {
	type row struct {
		ID          uuid.UUID `db:"id"`
		Name        string    `db:"name"`
		Specialties string    `db:"specialty"`
	}
	var rows []row
	if err := a.db.SelectContext(ctx, &rows,
		`SELECT id, name, specialty FROM doctors WHERE clinic_id=$1 AND active=TRUE ORDER BY name LIMIT 3`,
		clinicID); err != nil {
		return nil, err
	}
	out := make([]Doctor, len(rows))
	for i, r := range rows {
		out[i] = Doctor{ID: r.ID.String(), Name: r.Name, Specialties: []string{r.Specialties}}
	}
	return out, nil
}

func (a *MockAdapter) GetDoctor(ctx context.Context, clinicID uuid.UUID, id string) (*Doctor, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("bad id: %w", err)
	}
	var row struct {
		ID        uuid.UUID `db:"id"`
		Name      string    `db:"name"`
		Specialty string    `db:"specialty"`
	}
	if err := a.db.GetContext(ctx, &row,
		`SELECT id, name, specialty FROM doctors WHERE id=$1 AND clinic_id=$2 AND active=TRUE`,
		uid, clinicID); err != nil {
		return nil, err
	}
	return &Doctor{ID: row.ID.String(), Name: row.Name, Specialties: []string{row.Specialty}}, nil
}

func (a *MockAdapter) GetFreeSlots(ctx context.Context, clinicID uuid.UUID, from, to time.Time, specialty string) ([]Slot, error) {
	type doc struct {
		ID   uuid.UUID `db:"id"`
		Name string    `db:"name"`
	}
	var docs []doc
	q := `SELECT id, name FROM doctors WHERE clinic_id=$1 AND active=TRUE`
	args := []any{clinicID}
	if specialty != "" {
		q += ` AND specialty=$2`
		args = append(args, specialty)
	}
	q += ` LIMIT 3`
	if err := a.db.SelectContext(ctx, &docs, q, args...); err != nil {
		return nil, err
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	out := []Slot{}
	dur := 30 * time.Minute
	now := time.Now()
	// три следующих дня × 4 слота × N врачей — но срезаем 8 штук.
	for day := 0; day < 3 && len(out) < 8; day++ {
		base := time.Date(now.Year(), now.Month(), now.Day()+day, 10, 0, 0, 0, now.Location())
		hours := []int{10, 12, 15, 17}
		for _, h := range hours {
			start := time.Date(base.Year(), base.Month(), base.Day(), h, 0, 0, 0, base.Location())
			if start.Before(now.Add(30 * time.Minute)) {
				continue
			}
			for _, d := range docs {
				key := d.ID.String() + "|" + start.Format(time.RFC3339)
				if a.held[key] {
					continue
				}
				out = append(out, Slot{
					StartsAt: start,
					EndsAt:   start.Add(dur),
					DoctorID: d.ID,
					Doctor:   d.Name,
				})
				if len(out) >= 8 {
					break
				}
			}
		}
	}
	return out, nil
}

func (a *MockAdapter) CreateAppointment(ctx context.Context, req BookRequest) (*BookResult, error) {
	a.mu.Lock()
	key := req.DoctorID.String() + "|" + req.StartsAt.Format(time.RFC3339)
	a.held[key] = true
	a.mu.Unlock()

	var id uuid.UUID
	err := a.db.GetContext(ctx, &id,
		`INSERT INTO appointments(clinic_id, patient_id, doctor_id, chair_id, starts_at, ends_at, service, source)
		 VALUES($1, $2, $3, $4, $5, $6, $7, 'bot')
		 RETURNING id`,
		req.ClinicID, req.PatientID, req.DoctorID, req.ChairID, req.StartsAt, req.EndsAt, req.Service)
	if err != nil {
		return nil, err
	}
	return &BookResult{AppointmentID: id}, nil
}

func (a *MockAdapter) CancelAppointment(ctx context.Context, appointmentID uuid.UUID) error {
	_, err := a.db.ExecContext(ctx,
		`UPDATE appointments SET status='cancelled' WHERE id=$1`, appointmentID)
	return err
}
