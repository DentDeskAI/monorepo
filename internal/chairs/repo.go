package chairs

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Chair struct {
	ID         uuid.UUID `db:"id" json:"id"`
	ClinicID   uuid.UUID `db:"clinic_id" json:"clinic_id"`
	ExternalID *string   `db:"external_id" json:"external_id,omitempty"`
	Name       string    `db:"name" json:"name"`
	Active     bool      `db:"active" json:"active"`
}

type Repo struct{ db *sqlx.DB }

func NewRepo(db *sqlx.DB) *Repo { return &Repo{db: db} }

func (r *Repo) List(ctx context.Context, clinicID uuid.UUID) ([]Chair, error) {
	var out []Chair
	err := r.db.SelectContext(ctx, &out,
		`SELECT id, clinic_id, external_id, name, active
		 FROM chairs WHERE clinic_id=$1 ORDER BY name`, clinicID)
	return out, err
}

func (r *Repo) Get(ctx context.Context, id uuid.UUID) (*Chair, error) {
	var c Chair
	err := r.db.GetContext(ctx, &c,
		`SELECT id, clinic_id, external_id, name, active FROM chairs WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *Repo) Create(ctx context.Context, clinicID uuid.UUID, name string, externalID *string) (*Chair, error) {
	var c Chair
	err := r.db.GetContext(ctx, &c,
		`INSERT INTO chairs (clinic_id, name, external_id)
		 VALUES ($1, $2, $3)
		 RETURNING id, clinic_id, external_id, name, active`,
		clinicID, name, externalID)
	return &c, err
}

func (r *Repo) Update(ctx context.Context, id uuid.UUID, name string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE chairs SET name=$1 WHERE id=$2`, name, id)
	return err
}

func (r *Repo) Deactivate(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE chairs SET active=FALSE WHERE id=$1`, id)
	return err
}
