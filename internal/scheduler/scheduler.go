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
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Specialties []string `json:"specialties,omitempty"`
}

type Patient struct {
	Name      string  `json:"name"`
	Gender    *string `json:"gender"` // nullable
	ID        int     `json:"id"`
	IIN       *string `json:"iin"` // nullable
	Number    string  `json:"number"`
	Phone     *string `json:"phone"` // nullable
	Birth     *string `json:"birth"` // nullable (could be time.Time if formatted)
	IsChild   bool    `json:"isChild"`
	Comment   string  `json:"comment"`
	WhereKnow string  `json:"whereKnow"`
}

func toSchedulerPatient(p Patient) Patient {
	return Patient{
		Name:      p.Name,
		Gender:    p.Gender,
		ID:        p.ID,
		IIN:       p.IIN,
		Number:    p.Number,
		Phone:     p.Phone,
		Birth:     p.Birth,
		IsChild:   p.IsChild,
		Comment:   p.Comment,
		WhereKnow: p.WhereKnow,
	}
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
	GetDoctor(ctx context.Context, clinicID uuid.UUID, id string) (*Doctor, error)
	ListPatients(ctx context.Context, clinicID uuid.UUID) ([]Patient, error)
	GetPatient(ctx context.Context, clinicID uuid.UUID, id int) (*Patient, error)
	GetFreeSlots(ctx context.Context, clinicID uuid.UUID, from, to time.Time, specialty string) ([]Slot, error)
	CreateAppointment(ctx context.Context, req BookRequest) (*BookResult, error)
	CancelAppointment(ctx context.Context, appointmentID uuid.UUID) error
}
