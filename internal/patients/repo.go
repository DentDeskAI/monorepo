package patients

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Patient struct {
	ID         uuid.UUID `db:"id" json:"id"`
	ClinicID   uuid.UUID `db:"clinic_id" json:"clinic_id"`
	Phone      string    `db:"phone" json:"phone"`
	Name       *string   `db:"name" json:"name,omitempty"`
	ExternalID *string   `db:"external_id" json:"external_id,omitempty"`
	Language   string    `db:"language" json:"language"`
	SeqID      int       `db:"seq_id" json:"seq_id"`
}

type Repo struct{ db *sqlx.DB }

func NewRepo(db *sqlx.DB) *Repo { return &Repo{db: db} }

// GetOrCreateByPhone — используется webhook'ом при первом сообщении.
func (r *Repo) GetOrCreateByPhone(ctx context.Context, clinicID uuid.UUID, phone string) (*Patient, error) {
	var p Patient
	err := r.db.GetContext(ctx, &p,
		`SELECT id, clinic_id, phone, name, external_id, language, seq_id
		 FROM patients WHERE clinic_id=$1 AND phone=$2`, clinicID, phone)
	if err == nil {
		return &p, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	err = r.db.GetContext(ctx, &p,
		`INSERT INTO patients(clinic_id, phone) VALUES($1, $2)
		 RETURNING id, clinic_id, phone, name, external_id, language, seq_id`,
		clinicID, phone)
	return &p, err
}

func (r *Repo) List(ctx context.Context, clinicID uuid.UUID, limit int) ([]Patient, error) {
	out := make([]Patient, 0)
	err := r.db.SelectContext(ctx, &out,
		`SELECT id, clinic_id, phone, name, external_id, language, seq_id
		 FROM patients WHERE clinic_id=$1 ORDER BY created_at DESC LIMIT $2`,
		clinicID, limit)

	if out == nil {
		out = make([]Patient, 0)
	}
	return out, err
}

func (r *Repo) UpdateName(ctx context.Context, id uuid.UUID, name string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE patients SET name=$1 WHERE id=$2`, name, id)
	return err
}

func (r *Repo) Get(ctx context.Context, id uuid.UUID) (*Patient, error) {
	var p Patient
	err := r.db.GetContext(ctx, &p,
		`SELECT id, clinic_id, phone, name, external_id, language, seq_id
		 FROM patients WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repo) Create(ctx context.Context, clinicID uuid.UUID, phone, language string, name, externalID *string) (*Patient, error) {
	var p Patient
	err := r.db.GetContext(ctx, &p,
		`INSERT INTO patients (clinic_id, phone, language, name, external_id)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, clinic_id, phone, name, external_id, language, seq_id`,
		clinicID, phone, language, name, externalID)
	return &p, err
}

func (r *Repo) GetBySeqID(ctx context.Context, clinicID uuid.UUID, seqID int) (*Patient, error) {
	var p Patient
	err := r.db.GetContext(ctx, &p,
		`SELECT id, clinic_id, phone, name, external_id, language, seq_id
		 FROM patients WHERE clinic_id=$1 AND seq_id=$2`, clinicID, seqID)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repo) Update(ctx context.Context, id uuid.UUID, name *string, language string, externalID *string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE patients SET name=$1, language=$2, external_id=$3 WHERE id=$4`,
		name, language, externalID, id)
	return err
}
