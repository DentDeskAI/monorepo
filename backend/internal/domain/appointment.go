package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AppointmentStatus represents the lifecycle state of an appointment.
type AppointmentStatus string

const (
	AppointmentStatusScheduled  AppointmentStatus = "scheduled"
	AppointmentStatusConfirmed  AppointmentStatus = "confirmed"
	AppointmentStatusInProgress AppointmentStatus = "in_progress"
	AppointmentStatusCompleted  AppointmentStatus = "completed"
	AppointmentStatusCancelled  AppointmentStatus = "cancelled"
	AppointmentStatusNoShow     AppointmentStatus = "no_show"
)

// Appointment represents a scheduled visit between a patient and a doctor.
type Appointment struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Multi-tenancy key
	ClinicID uuid.UUID `gorm:"type:uuid;not null;index" json:"clinic_id"`

	PatientID uuid.UUID `gorm:"type:uuid;not null;index" json:"patient_id"`
	Patient   Patient   `gorm:"foreignKey:PatientID" json:"patient,omitempty"`

	DoctorID uuid.UUID `gorm:"type:uuid;not null;index" json:"doctor_id"`
	Doctor   Doctor    `gorm:"foreignKey:DoctorID" json:"doctor,omitempty"`

	StartsAt time.Time `gorm:"not null;index" json:"starts_at"`
	EndsAt   time.Time `gorm:"not null" json:"ends_at"`

	Status AppointmentStatus `gorm:"size:50;default:'scheduled'" json:"status"`
	Title  string            `gorm:"size:255" json:"title"` // procedure description
	Notes  string            `gorm:"type:text" json:"notes"`

	// Reminders
	ReminderSentAt *time.Time `json:"reminder_sent_at,omitempty"`
}

func (a *Appointment) BeforeCreate(_ *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

func (a *Appointment) Duration() time.Duration {
	return a.EndsAt.Sub(a.StartsAt)
}
