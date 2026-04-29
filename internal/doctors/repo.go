package doctors

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Doctor struct {
	ID         uuid.UUID `db:"id" json:"id"`
	ClinicID   uuid.UUID `db:"clinic_id" json:"clinic_id"`
	ExternalID *string   `db:"external_id" json:"external_id,omitempty"`
	Name       string    `db:"name" json:"name"`
	Specialty  *string   `db:"specialty" json:"specialty,omitempty"`
	Active     bool      `db:"active" json:"active"`
	SeqID      int       `db:"seq_id" json:"seq_id"`
}

type Repo struct{ db *sqlx.DB }

func NewRepo(db *sqlx.DB) *Repo { return &Repo{db: db} }

func (r *Repo) List(ctx context.Context, clinicID uuid.UUID) ([]Doctor, error) {
	var out []Doctor
	err := r.db.SelectContext(ctx, &out,
		`SELECT id, clinic_id, external_id, name, specialty, active, seq_id
		 FROM doctors WHERE clinic_id=$1 AND active=TRUE ORDER BY name`, clinicID)
	return out, err
}

func (r *Repo) FindBySpecialty(ctx context.Context, clinicID uuid.UUID, specialty string) ([]Doctor, error) {
	var out []Doctor
	err := r.db.SelectContext(ctx, &out,
		`SELECT id, clinic_id, external_id, name, specialty, active, seq_id
		 FROM doctors WHERE clinic_id=$1 AND active=TRUE AND specialty=$2`,
		clinicID, specialty)
	return out, err
}

func (r *Repo) GetByExternalID(ctx context.Context, clinicID uuid.UUID, externalID string) (*Doctor, error) {
	var d Doctor
	err := r.db.GetContext(ctx, &d,
		`SELECT id, clinic_id, external_id, name, specialty, active, seq_id
		 FROM doctors WHERE clinic_id=$1 AND external_id=$2`, clinicID, externalID)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *Repo) Get(ctx context.Context, id uuid.UUID) (*Doctor, error) {
	var d Doctor
	err := r.db.GetContext(ctx, &d,
		`SELECT id, clinic_id, external_id, name, specialty, active, seq_id FROM doctors WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *Repo) GetBySeqID(ctx context.Context, clinicID uuid.UUID, seqID int) (*Doctor, error) {
	var d Doctor
	err := r.db.GetContext(ctx, &d,
		`SELECT id, clinic_id, external_id, name, specialty, active, seq_id
		 FROM doctors WHERE clinic_id=$1 AND seq_id=$2 AND active=TRUE`, clinicID, seqID)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// Upsert inserts a doctor identified by external_id, or updates name/specialty if it already exists.
func (r *Repo) Upsert(ctx context.Context, clinicID uuid.UUID, name string, specialty *string, externalID string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO doctors (clinic_id, name, specialty, external_id)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (clinic_id, external_id) DO UPDATE
		   SET name=$2, specialty=$3`,
		clinicID, name, specialty, externalID)
	return err
}

func (r *Repo) Create(ctx context.Context, clinicID uuid.UUID, name string, specialty *string, externalID *string) (*Doctor, error) {
	var d Doctor
	err := r.db.GetContext(ctx, &d,
		`INSERT INTO doctors (clinic_id, name, specialty, external_id)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, clinic_id, external_id, name, specialty, active, seq_id`,
		clinicID, name, specialty, externalID)
	return &d, err
}

func (r *Repo) Update(ctx context.Context, id uuid.UUID, name string, specialty *string, active bool) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE doctors SET name=$1, specialty=$2, active=$3 WHERE id=$4`,
		name, specialty, active, id)
	return err
}

func (r *Repo) Deactivate(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE doctors SET active=FALSE WHERE id=$1`, id)
	return err
}
