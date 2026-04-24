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
}

type Repo struct{ db *sqlx.DB }

func NewRepo(db *sqlx.DB) *Repo { return &Repo{db: db} }

func (r *Repo) List(ctx context.Context, clinicID uuid.UUID) ([]Doctor, error) {
	var out []Doctor
	err := r.db.SelectContext(ctx, &out,
		`SELECT id, clinic_id, external_id, name, specialty, active
		 FROM doctors WHERE clinic_id=$1 AND active=TRUE ORDER BY name`, clinicID)
	return out, err
}

func (r *Repo) FindBySpecialty(ctx context.Context, clinicID uuid.UUID, specialty string) ([]Doctor, error) {
	var out []Doctor
	err := r.db.SelectContext(ctx, &out,
		`SELECT id, clinic_id, external_id, name, specialty, active
		 FROM doctors WHERE clinic_id=$1 AND active=TRUE AND specialty=$2`,
		clinicID, specialty)
	return out, err
}

func (r *Repo) Get(ctx context.Context, id uuid.UUID) (*Doctor, error) {
	var d Doctor
	err := r.db.GetContext(ctx, &d,
		`SELECT id, clinic_id, external_id, name, specialty, active FROM doctors WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	return &d, nil
}
