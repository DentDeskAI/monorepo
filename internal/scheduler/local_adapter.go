package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// LocalAdapter — источник расписания на нашей БД. Используется, когда у клиники нет MacDent.
type LocalAdapter struct {
	db              *sqlx.DB
	slotDurationMin int
	openHour        int
	closeHour       int
}

func (a *LocalAdapter) ListPatients(ctx context.Context, clinicID uuid.UUID) ([]Patient, error) {
	return nil, nil
}

func (a *LocalAdapter) GetPatient(ctx context.Context, clinicID uuid.UUID, id int) (*Patient, error) {
	return nil, fmt.Errorf("not supported in local adapter")
}

func (a *LocalAdapter) GetClinic(ctx context.Context, clinicID uuid.UUID) (*Stomatology, error) {
	var row struct {
		ID   string `db:"id"`
		Name string `db:"name"`
	}
	if err := a.db.GetContext(ctx, &row, `SELECT id, name FROM clinics WHERE id=$1`, clinicID); err != nil {
		return nil, err
	}
	return &Stomatology{ID: row.ID, Name: row.Name}, nil
}

func NewLocalAdapter(db *sqlx.DB) *LocalAdapter {
	return &LocalAdapter{db: db, slotDurationMin: 30, openHour: 9, closeHour: 20}
}

type busyRow struct {
	DoctorID uuid.UUID `db:"doctor_id"`
	StartsAt time.Time `db:"starts_at"`
	EndsAt   time.Time `db:"ends_at"`
}

func (a *LocalAdapter) ListDoctors(ctx context.Context, clinicID uuid.UUID) ([]Doctor, error) {
	type row struct {
		ID        uuid.UUID `db:"id"`
		Name      string    `db:"name"`
		Specialty string    `db:"specialty"`
	}
	var rows []row
	if err := a.db.SelectContext(ctx, &rows,
		`SELECT id, name, specialty FROM doctors WHERE clinic_id=$1 AND active=TRUE ORDER BY name`,
		clinicID); err != nil {
		return nil, err
	}
	out := make([]Doctor, len(rows))
	for i, r := range rows {
		out[i] = Doctor{ID: r.ID.String(), Name: r.Name, Specialties: []string{r.Specialty}}
	}
	return out, nil
}

func (a *LocalAdapter) GetDoctor(ctx context.Context, clinicID uuid.UUID, id string) (*Doctor, error) {
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

func (a *LocalAdapter) GetFreeSlots(ctx context.Context, clinicID uuid.UUID, from, to time.Time, specialty string) ([]Slot, error) {
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
	if err := a.db.SelectContext(ctx, &docs, q, args...); err != nil {
		return nil, err
	}
	if len(docs) == 0 {
		return nil, nil
	}

	var busies []busyRow
	if err := a.db.SelectContext(ctx, &busies,
		`SELECT doctor_id, starts_at, ends_at FROM appointments
		 WHERE clinic_id=$1 AND status IN ('scheduled','confirmed')
		   AND starts_at < $3 AND ends_at > $2`,
		clinicID, from, to); err != nil {
		return nil, err
	}

	busyByDoctor := map[uuid.UUID][]busyRow{}
	for _, b := range busies {
		busyByDoctor[b.DoctorID] = append(busyByDoctor[b.DoctorID], b)
	}

	dur := time.Duration(a.slotDurationMin) * time.Minute
	now := time.Now()
	var out []Slot
	for _, d := range docs {
		day := truncateDay(from)
		for !day.After(to) {
			openAt := time.Date(day.Year(), day.Month(), day.Day(), a.openHour, 0, 0, 0, day.Location())
			closeAt := time.Date(day.Year(), day.Month(), day.Day(), a.closeHour, 0, 0, 0, day.Location())
			for t := openAt; !t.Add(dur).After(closeAt); t = t.Add(dur) {
				end := t.Add(dur)
				if end.Before(from) || t.After(to) {
					continue
				}
				if t.Before(now.Add(30 * time.Minute)) {
					continue
				}
				if intersectsAny(t, end, busyByDoctor[d.ID]) {
					continue
				}
				out = append(out, Slot{
					StartsAt: t, EndsAt: end,
					DoctorID: d.ID, Doctor: d.Name,
				})
				if len(out) >= 48 {
					return out, nil
				}
			}
			day = day.Add(24 * time.Hour)
		}
	}
	return out, nil
}

func (a *LocalAdapter) CreateAppointment(ctx context.Context, req BookRequest) (*BookResult, error) {
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

func (a *LocalAdapter) CancelAppointment(ctx context.Context, appointmentID uuid.UUID) error {
	_, err := a.db.ExecContext(ctx,
		`UPDATE appointments SET status='cancelled' WHERE id=$1`, appointmentID)
	return err
}

func truncateDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func intersectsAny(start, end time.Time, busies []busyRow) bool {
	for _, b := range busies {
		if start.Before(b.EndsAt) && end.After(b.StartsAt) {
			return true
		}
	}
	return false
}
