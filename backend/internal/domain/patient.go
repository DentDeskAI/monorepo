package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Patient represents a dental clinic's patient. Scoped to a tenant via ClinicID.
type Patient struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Multi-tenancy key — always included in WHERE clauses
	ClinicID uuid.UUID `gorm:"type:uuid;not null;index" json:"clinic_id"`

	FirstName   string    `gorm:"size:100;not null" json:"first_name"`
	LastName    string    `gorm:"size:100;not null" json:"last_name"`
	Phone       string    `gorm:"size:30;not null" json:"phone"` // WhatsApp contact
	Email       string    `gorm:"size:255" json:"email"`
	DateOfBirth *time.Time `json:"date_of_birth,omitempty"`
	Gender      string    `gorm:"size:20" json:"gender"`
	Notes       string    `gorm:"type:text" json:"notes"`

	// Relations
	Appointments []Appointment `gorm:"foreignKey:PatientID" json:"-"`
	MessageLogs  []MessageLog  `gorm:"foreignKey:PatientID" json:"-"`
}

func (p *Patient) FullName() string {
	return p.FirstName + " " + p.LastName
}

func (p *Patient) BeforeCreate(_ *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
