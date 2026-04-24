package conversations

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Conversation struct {
	ID            uuid.UUID       `db:"id" json:"id"`
	ClinicID      uuid.UUID       `db:"clinic_id" json:"clinic_id"`
	PatientID     uuid.UUID       `db:"patient_id" json:"patient_id"`
	Status        string          `db:"status" json:"status"`
	Context       json.RawMessage `db:"context" json:"context"`
	LastMessageAt time.Time       `db:"last_message_at" json:"last_message_at"`
	CreatedAt     time.Time       `db:"created_at" json:"created_at"`
}

type Message struct {
	ID             uuid.UUID       `db:"id" json:"id"`
	ConversationID uuid.UUID       `db:"conversation_id" json:"conversation_id"`
	WAMessageID    *string         `db:"wa_message_id" json:"wa_message_id,omitempty"`
	Direction      string          `db:"direction" json:"direction"`
	Sender         string          `db:"sender" json:"sender"`
	Body           string          `db:"body" json:"body"`
	Meta           json.RawMessage `db:"meta" json:"meta"`
	CreatedAt      time.Time       `db:"created_at" json:"created_at"`
}

type Repo struct{ db *sqlx.DB }

func NewRepo(db *sqlx.DB) *Repo { return &Repo{db: db} }

// GetOrCreate — идемпотентное получение/создание диалога пациент-клиника.
func (r *Repo) GetOrCreate(ctx context.Context, clinicID, patientID uuid.UUID) (*Conversation, error) {
	var c Conversation
	err := r.db.GetContext(ctx, &c,
		`SELECT id, clinic_id, patient_id, status, context, last_message_at, created_at
		 FROM conversations WHERE clinic_id=$1 AND patient_id=$2`, clinicID, patientID)
	if err == nil {
		return &c, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	err = r.db.GetContext(ctx, &c,
		`INSERT INTO conversations(clinic_id, patient_id)
		 VALUES($1, $2)
		 RETURNING id, clinic_id, patient_id, status, context, last_message_at, created_at`,
		clinicID, patientID)
	return &c, err
}

func (r *Repo) UpdateContext(ctx context.Context, id uuid.UUID, contextJSON json.RawMessage) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE conversations SET context=$1, last_message_at=NOW() WHERE id=$2`,
		contextJSON, id)
	return err
}

func (r *Repo) SetStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE conversations SET status=$1 WHERE id=$2`, status, id)
	return err
}

func (r *Repo) ListForClinic(ctx context.Context, clinicID uuid.UUID, limit int) ([]Conversation, error) {
	var out []Conversation
	err := r.db.SelectContext(ctx, &out,
		`SELECT id, clinic_id, patient_id, status, context, last_message_at, created_at
		 FROM conversations WHERE clinic_id=$1 ORDER BY last_message_at DESC LIMIT $2`,
		clinicID, limit)
	return out, err
}

// InsertMessage — возвращает существующее сообщение, если wa_message_id уже есть (идемпотентность).
// Возвращает (msg, isNew).
func (r *Repo) InsertMessage(ctx context.Context, m *Message) (*Message, bool, error) {
	if len(m.Meta) == 0 {
		m.Meta = json.RawMessage("{}")
	}
	row := r.db.QueryRowxContext(ctx,
		`INSERT INTO messages(conversation_id, wa_message_id, direction, sender, body, meta)
		 VALUES($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (wa_message_id) DO NOTHING
		 RETURNING id, conversation_id, wa_message_id, direction, sender, body, meta, created_at`,
		m.ConversationID, m.WAMessageID, m.Direction, m.Sender, m.Body, m.Meta)

	var inserted Message
	if err := row.StructScan(&inserted); err != nil {
		if errors.Is(err, sql.ErrNoRows) && m.WAMessageID != nil {
			// дубликат — достаём существующее
			var existing Message
			if err := r.db.GetContext(ctx, &existing,
				`SELECT id, conversation_id, wa_message_id, direction, sender, body, meta, created_at
				 FROM messages WHERE wa_message_id=$1`, *m.WAMessageID); err != nil {
				return nil, false, err
			}
			return &existing, false, nil
		}
		return nil, false, err
	}
	// обновим last_message_at в диалоге
	if _, err := r.db.ExecContext(ctx,
		`UPDATE conversations SET last_message_at=NOW() WHERE id=$1`, inserted.ConversationID); err != nil {
		return &inserted, true, err
	}
	return &inserted, true, nil
}

func (r *Repo) ListMessages(ctx context.Context, conversationID uuid.UUID, limit int) ([]Message, error) {
	var out []Message
	err := r.db.SelectContext(ctx, &out,
		`SELECT id, conversation_id, wa_message_id, direction, sender, body, meta, created_at
		 FROM messages WHERE conversation_id=$1 ORDER BY created_at DESC LIMIT $2`,
		conversationID, limit)
	return out, err
}

// RecentHistory возвращает последние N сообщений в хронологическом порядке (для LLM контекста).
func (r *Repo) RecentHistory(ctx context.Context, conversationID uuid.UUID, limit int) ([]Message, error) {
	msgs, err := r.ListMessages(ctx, conversationID, limit)
	if err != nil {
		return nil, err
	}
	// разворачиваем
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, nil
}
