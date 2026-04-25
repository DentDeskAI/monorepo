package clinics

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Clinic struct {
	ID              uuid.UUID       `db:"id" json:"id"`
	Name            string          `db:"name" json:"name"`
	Timezone        string          `db:"timezone" json:"timezone"`
	WhatsAppPhoneID *string         `db:"whatsapp_phone_id" json:"whatsapp_phone_id,omitempty"`
	SchedulerType   string          `db:"scheduler_type" json:"scheduler_type"`
	MacDentBaseURL  *string         `db:"macdent_base_url" json:"macdent_base_url,omitempty"`
	MacDentApiKey   *string         `db:"macdent_api_key" json:"macdent_api_key"`
	WorkingHours    json.RawMessage `db:"working_hours" json:"working_hours"`
	SlotDurationMin int             `db:"slot_duration_min" json:"slot_duration_min"`
	CreatedAt       time.Time       `db:"created_at" json:"created_at"`
}

type UpdateFields struct {
	Name            string
	Timezone        string
	WorkingHours    json.RawMessage
	SlotDurationMin int
	SchedulerType   string
}

type Repo struct{ db *sqlx.DB }

func NewRepo(db *sqlx.DB) *Repo { return &Repo{db: db} }

func (r *Repo) Get(ctx context.Context, id uuid.UUID) (*Clinic, error) {
	var c Clinic
	err := r.db.GetContext(ctx, &c,
		`SELECT id, name, timezone, whatsapp_phone_id, scheduler_type,
		        macdent_base_url, macdent_api_key, working_hours, slot_duration_min, created_at
		 FROM clinics WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *Repo) Create(ctx context.Context, name, timezone, schedulerType string) (*Clinic, error) {
	var c Clinic
	err := r.db.GetContext(ctx, &c,
		`INSERT INTO clinics (name, timezone, scheduler_type)
		 VALUES ($1, $2, $3)
		 RETURNING id, name, timezone, whatsapp_phone_id, scheduler_type,
		           macdent_base_url, working_hours, slot_duration_min, created_at`,
		name, timezone, schedulerType)
	return &c, err
}

func (r *Repo) Update(ctx context.Context, id uuid.UUID, f UpdateFields) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE clinics
		 SET name=$1, timezone=$2, working_hours=$3, slot_duration_min=$4, scheduler_type=$5
		 WHERE id=$6`,
		f.Name, f.Timezone, f.WorkingHours, f.SlotDurationMin, f.SchedulerType, id)
	return err
}
