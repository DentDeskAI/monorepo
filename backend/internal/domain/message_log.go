package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MessageDirection indicates if a message was sent by the clinic or patient.
type MessageDirection string

const (
	MessageDirectionInbound  MessageDirection = "inbound"  // patient → clinic
	MessageDirectionOutbound MessageDirection = "outbound" // clinic → patient
)

// MessageStatus tracks delivery state for outbound messages.
type MessageStatus string

const (
	MessageStatusPending   MessageStatus = "pending"
	MessageStatusSent      MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead      MessageStatus = "read"
	MessageStatusFailed    MessageStatus = "failed"
)

// MessageLog stores every WhatsApp message exchanged for a tenant.
// Acts as the source of truth for the Dialogs (inbox) feature.
type MessageLog struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Multi-tenancy key
	ClinicID uuid.UUID `gorm:"type:uuid;not null;index" json:"clinic_id"`

	// May be nil for messages from unknown contacts (to be matched later)
	PatientID *uuid.UUID `gorm:"type:uuid;index" json:"patient_id,omitempty"`
	Patient   *Patient   `gorm:"foreignKey:PatientID" json:"patient,omitempty"`

	// WhatsApp identifiers
	WhatsAppMessageID string `gorm:"size:255;index" json:"wa_message_id"` // wamid from Meta
	FromPhone         string `gorm:"size:30;not null" json:"from_phone"`
	ToPhone           string `gorm:"size:30;not null" json:"to_phone"`

	Direction MessageDirection `gorm:"size:20;not null" json:"direction"`
	Status    MessageStatus    `gorm:"size:20;default:'pending'" json:"status"`

	// Content
	MessageType string `gorm:"size:50;default:'text'" json:"message_type"` // text | image | audio | template
	Body        string `gorm:"type:text" json:"body"`
	MediaURL    string `gorm:"size:500" json:"media_url,omitempty"`

	// LLM metadata — populated when AI assisted the reply
	LLMUsed   bool   `gorm:"default:false" json:"llm_used"`
	LLMPrompt string `gorm:"type:text" json:"-"` // stored for audit, hidden from API
}

func (m *MessageLog) BeforeCreate(_ *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
