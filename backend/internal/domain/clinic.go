// Package domain defines the core business entities of DentDesk.
package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Clinic represents a dental clinic tenant in the system.
// Every other entity is scoped to a ClinicID for multi-tenancy.
type Clinic struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name     string `gorm:"not null;size:255" json:"name"`
	Slug     string `gorm:"uniqueIndex;size:100" json:"slug"` // used for sub-domain routing
	Phone    string `gorm:"size:30" json:"phone"`
	Email    string `gorm:"uniqueIndex;size:255" json:"email"`
	Address  string `gorm:"size:500" json:"address"`
	LogoURL  string `gorm:"size:500" json:"logo_url"`
	IsActive bool   `gorm:"default:true" json:"is_active"`

	// Relations
	Doctors      []Doctor      `gorm:"foreignKey:ClinicID" json:"-"`
	Patients     []Patient     `gorm:"foreignKey:ClinicID" json:"-"`
	Appointments []Appointment `gorm:"foreignKey:ClinicID" json:"-"`
	MessageLogs  []MessageLog  `gorm:"foreignKey:ClinicID" json:"-"`
}

// User represents a staff member (admin, receptionist) of a clinic.
type User struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Multi-tenancy key
	ClinicID uuid.UUID `gorm:"type:uuid;not null;index" json:"clinic_id"`
	Clinic   Clinic    `gorm:"foreignKey:ClinicID" json:"-"`

	Email        string   `gorm:"uniqueIndex;size:255;not null" json:"email"`
	PasswordHash string   `gorm:"not null" json:"-"`
	FirstName    string   `gorm:"size:100" json:"first_name"`
	LastName     string   `gorm:"size:100" json:"last_name"`
	Role         UserRole `gorm:"size:50;default:'receptionist'" json:"role"`
	IsActive     bool     `gorm:"default:true" json:"is_active"`
}

// UserRole defines the permission level of a staff member.
type UserRole string

const (
	RoleAdmin        UserRole = "admin"
	RoleReceptionist UserRole = "receptionist"
	RoleDoctor       UserRole = "doctor"
)

// BeforeCreate sets a UUID primary key before insertion.
func (c *Clinic) BeforeCreate(_ *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
