package appointments

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Appointment struct {
	ID         uuid.UUID  `db:"id" json:"id"`
	ClinicID   uuid.UUID  `db:"clinic_id" json:"clinic_id"`
	PatientID  uuid.UUID  `db:"patient_id" json:"patient_id"`
	DoctorID   *uuid.UUID `db:"doctor_id" json:"doctor_id,omitempty"`
	ChairID    *uuid.UUID `db:"chair_id" json:"chair_id,omitempty"`
	ExternalID *string    `db:"external_id" json:"external_id,omitempty"`
	StartsAt   time.Time  `db:"starts_at" json:"starts_at"`
	EndsAt     time.Time  `db:"ends_at" json:"ends_at"`
	Service    *string    `db:"service" json:"service,omitempty"`
	Status     string     `db:"status" json:"status"`
	Source     string     `db:"source" json:"source"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
	// Опциональные join-поля для UI
	PatientName  *string `db:"patient_name" json:"patient_name,omitempty"`
	PatientPhone *string `db:"patient_phone" json:"patient_phone,omitempty"`
	DoctorName   *string `db:"doctor_name" json:"doctor_name,omitempty"`
}

type Repo struct{ db *sqlx.DB }

func NewRepo(db *sqlx.DB) *Repo { return &Repo{db: db} }

// ListForPeriod — для календаря CRM.
func (r *Repo) ListForPeriod(ctx context.Context, clinicID uuid.UUID, from, to time.Time) ([]Appointment, error) {
	var out []Appointment
	err := r.db.SelectContext(ctx, &out,
		`SELECT a.id, a.clinic_id, a.patient_id, a.doctor_id, a.chair_id, a.external_id,
		        a.starts_at, a.ends_at, a.service, a.status, a.source, a.created_at,
		        p.name AS patient_name, p.phone AS patient_phone,
		        d.name AS doctor_name
		 FROM appointments a
		 LEFT JOIN patients p ON p.id = a.patient_id
		 LEFT JOIN doctors  d ON d.id = a.doctor_id
		 WHERE a.clinic_id=$1 AND a.starts_at >= $2 AND a.starts_at < $3
		 ORDER BY a.starts_at`, clinicID, from, to)
	return out, err
}

func (r *Repo) ListForPatient(ctx context.Context, patientID uuid.UUID) ([]Appointment, error) {
	var out []Appointment
	err := r.db.SelectContext(ctx, &out,
		`SELECT a.id, a.clinic_id, a.patient_id, a.doctor_id, a.chair_id, a.external_id,
		        a.starts_at, a.ends_at, a.service, a.status, a.source, a.created_at,
		        d.name AS doctor_name
		 FROM appointments a
		 LEFT JOIN doctors d ON d.id = a.doctor_id
		 WHERE a.patient_id=$1 ORDER BY a.starts_at DESC LIMIT 50`, patientID)
	return out, err
}

func (r *Repo) Get(ctx context.Context, id uuid.UUID) (*Appointment, error) {
	var a Appointment
	err := r.db.GetContext(ctx, &a,
		`SELECT a.id, a.clinic_id, a.patient_id, a.doctor_id, a.chair_id, a.external_id,
		        a.starts_at, a.ends_at, a.service, a.status, a.source, a.created_at,
		        p.name AS patient_name, p.phone AS patient_phone,
		        d.name AS doctor_name
		 FROM appointments a
		 LEFT JOIN patients p ON p.id = a.patient_id
		 LEFT JOIN doctors  d ON d.id = a.doctor_id
		 WHERE a.id=$1`, id)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// DueForReminder24h — для worker'а: записи через ~24ч, без напоминания.
func (r *Repo) DueForReminder24h(ctx context.Context, now time.Time) ([]Appointment, error) {
	var out []Appointment
	// окно 23ч30м..24ч30м от now
	err := r.db.SelectContext(ctx, &out,
		`SELECT a.id, a.clinic_id, a.patient_id, a.doctor_id, a.chair_id, a.external_id,
		        a.starts_at, a.ends_at, a.service, a.status, a.source, a.created_at,
		        p.name AS patient_name, p.phone AS patient_phone,
		        d.name AS doctor_name
		 FROM appointments a
		 JOIN patients p ON p.id = a.patient_id
		 LEFT JOIN doctors d ON d.id = a.doctor_id
		 WHERE a.status IN ('scheduled','confirmed')
		   AND a.reminder_24h_sent_at IS NULL
		   AND a.starts_at BETWEEN $1 AND $2`,
		now.Add(23*time.Hour+30*time.Minute),
		now.Add(24*time.Hour+30*time.Minute),
	)
	return out, err
}

// DueForReminder2h — записи через ~2ч.
func (r *Repo) DueForReminder2h(ctx context.Context, now time.Time) ([]Appointment, error) {
	var out []Appointment
	err := r.db.SelectContext(ctx, &out,
		`SELECT a.id, a.clinic_id, a.patient_id, a.doctor_id, a.chair_id, a.external_id,
		        a.starts_at, a.ends_at, a.service, a.status, a.source, a.created_at,
		        p.name AS patient_name, p.phone AS patient_phone,
		        d.name AS doctor_name
		 FROM appointments a
		 JOIN patients p ON p.id = a.patient_id
		 LEFT JOIN doctors d ON d.id = a.doctor_id
		 WHERE a.status IN ('scheduled','confirmed')
		   AND a.reminder_2h_sent_at IS NULL
		   AND a.starts_at BETWEEN $1 AND $2`,
		now.Add(1*time.Hour+30*time.Minute),
		now.Add(2*time.Hour+30*time.Minute),
	)
	return out, err
}

// DueForFollowup — visited today/yesterday без follow-up.
func (r *Repo) DueForFollowup(ctx context.Context, now time.Time) ([]Appointment, error) {
	var out []Appointment
	err := r.db.SelectContext(ctx, &out,
		`SELECT a.id, a.clinic_id, a.patient_id, a.doctor_id, a.chair_id, a.external_id,
		        a.starts_at, a.ends_at, a.service, a.status, a.source, a.created_at,
		        p.name AS patient_name, p.phone AS patient_phone
		 FROM appointments a
		 JOIN patients p ON p.id = a.patient_id
		 WHERE a.status='completed'
		   AND a.followup_sent_at IS NULL
		   AND a.ends_at < $1
		   AND a.ends_at > $2`,
		now.Add(-2*time.Hour),
		now.Add(-72*time.Hour),
	)
	return out, err
}

func (r *Repo) MarkReminder24hSent(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE appointments SET reminder_24h_sent_at=NOW() WHERE id=$1`, id)
	return err
}
func (r *Repo) MarkReminder2hSent(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE appointments SET reminder_2h_sent_at=NOW() WHERE id=$1`, id)
	return err
}
func (r *Repo) MarkFollowupSent(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE appointments SET followup_sent_at=NOW() WHERE id=$1`, id)
	return err
}

func (r *Repo) SetStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE appointments SET status=$1 WHERE id=$2`, status, id)
	return err
}

func (r *Repo) Create(ctx context.Context, a *Appointment) (*Appointment, error) {
	var out Appointment
	err := r.db.GetContext(ctx, &out,
		`INSERT INTO appointments
		   (clinic_id, patient_id, doctor_id, chair_id, starts_at, ends_at, service, status, source)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		 RETURNING id, clinic_id, patient_id, doctor_id, chair_id, external_id,
		           starts_at, ends_at, service, status, source, created_at`,
		a.ClinicID, a.PatientID, a.DoctorID, a.ChairID,
		a.StartsAt, a.EndsAt, a.Service, a.Status, a.Source)
	return &out, err
}
