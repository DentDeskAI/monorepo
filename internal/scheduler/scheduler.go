// Package scheduler — абстракция над источниками расписания.
// Позволяет обслуживать клиники как с MacDent, так и без него (local/mock).
package scheduler

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Slot struct {
	StartsAt time.Time  `json:"starts_at"`
	EndsAt   time.Time  `json:"ends_at"`
	DoctorID uuid.UUID  `json:"doctor_id"`
	Doctor   string     `json:"doctor"`
	ChairID  *uuid.UUID `json:"chair_id,omitempty"`
}

type Doctor struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Specialty string `json:"specialty,omitempty"`
}

type BookRequest struct {
	ClinicID  uuid.UUID
	PatientID uuid.UUID
	DoctorID  uuid.UUID
	ChairID   *uuid.UUID
	StartsAt  time.Time
	EndsAt    time.Time
	Service   string
}

type BookResult struct {
	AppointmentID uuid.UUID
	ExternalID    *string
}

// Scheduler — то, что умеет наш поставщик расписания.
// LocalAdapter хранит в нашей PG. MacDentAdapter ходит в чужой API.
type Scheduler interface {
	ListDoctors(ctx context.Context, clinicID uuid.UUID) ([]Doctor, error)
	GetFreeSlots(ctx context.Context, clinicID uuid.UUID, from, to time.Time, specialty string) ([]Slot, error)
	CreateAppointment(ctx context.Context, req BookRequest) (*BookResult, error)
	CancelAppointment(ctx context.Context, appointmentID uuid.UUID) error
}
