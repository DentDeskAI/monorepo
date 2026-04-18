package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Doctor represents a dental professional working at a clinic.
type Doctor struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Multi-tenancy key
	ClinicID uuid.UUID `gorm:"type:uuid;not null;index" json:"clinic_id"`

	// Optional link to a User account for portal access
	UserID *uuid.UUID `gorm:"type:uuid;index" json:"user_id,omitempty"`

	FirstName   string `gorm:"size:100;not null" json:"first_name"`
	LastName    string `gorm:"size:100;not null" json:"last_name"`
	Speciality  string `gorm:"size:200" json:"speciality"` // e.g. "Orthodontics"
	AvatarURL   string `gorm:"size:500" json:"avatar_url"`
	Color       string `gorm:"size:7;default:'#4F46E5'" json:"color"` // hex, used in calendar

	// Relations
	Appointments []Appointment `gorm:"foreignKey:DoctorID" json:"-"`
}

func (d *Doctor) FullName() string {
	return d.FirstName + " " + d.LastName
}

func (d *Doctor) BeforeCreate(_ *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}
