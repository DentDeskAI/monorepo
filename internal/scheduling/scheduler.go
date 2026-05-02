// Package scheduling is the scheduling abstraction layer.
//
// Two backends implement the Scheduler interface:
//   - LocalScheduler (local.go)  — reads/writes local PostgreSQL tables.
//     Used for clinics with scheduler_type = 'local' or 'mock'.
//   - Service (macdent.go)       — reads/writes the MacDent ERP API.
//     Used for clinics with scheduler_type = 'macdent'.
//
// Registry (registry.go) wraps both and dispatches based on each clinic's
// scheduler_type column. Callers only depend on the Scheduler interface.
package scheduling

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ── domain types ──────────────────────────────────────────────────────────────

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
	Gender    *string `json:"gender"`
	ID        int     `json:"id"`
	IIN       *string `json:"iin"`
	Number    string  `json:"number"`
	Phone     *string `json:"phone"`
	Birth     *string `json:"birth"`
	IsChild   bool    `json:"isChild"`
	Comment   string  `json:"comment"`
	WhereKnow string  `json:"whereKnow"`
}

type Stomatology struct {
	ID   string `json:"id"`
	Name string `json:"name"`
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

type Appointment struct {
	ID      int    `json:"id"`
	Doctor  int    `json:"doctor"`
	Patient int    `json:"patient"`
	Date    string `json:"date"`
	Start   string `json:"start"`
	End     string `json:"end"`
	Status  int    `json:"status"`
	Zhaloba string `json:"zhaloba"`
	Comment string `json:"comment"`
	IsFirst bool   `json:"isFirst"`
	Cabinet string `json:"cabinet"`
	Rasp    string `json:"rasp"`
}

type AppointmentsResponse struct {
	Appointments []Appointment `json:"appointments"`
}

type AppointmentDetail struct {
	ID      int                   `json:"id"`
	Doctor  AppointmentDoctorRef  `json:"doctor"`
	Patient AppointmentPatientRef `json:"patient"`
	Date    string                `json:"date"`
	Start   string                `json:"start"`
	End     string                `json:"end"`
	Status  int                   `json:"status"`
	Zhaloba string                `json:"zhaloba"`
	Comment string                `json:"comment"`
	IsFirst bool                  `json:"isFirst"`
	Cabinet string                `json:"cabinet"`
	Rasp    string                `json:"rasp"`
}

type AppointmentDoctorRef struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type AppointmentPatientRef struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Phone *string `json:"phone"`
}

type UpdateAppointmentParams struct {
	DoctorID *int
	Start    *time.Time
	End      *time.Time
	Zhaloba  *string
	Comment  *string
}

type AppointmentRequestParams struct {
	PatientName  string
	PatientPhone string
	Start        time.Time
	End          time.Time
	WhereKnow    string
}

type AppointmentRequestResult struct {
	ID int `json:"id"`
}

type ScheduleAppointmentParams struct {
	DoctorID  int
	PatientID int
	Start     time.Time
	End       time.Time
	Zhaloba   string
	Cabinet   string
	IsFirst   bool
}

type ScheduleAppointmentResult struct {
	ID int `json:"id"`
}

type CreatePatientParams struct {
	Name      string
	Phone     string
	IIN       string
	Birth     string
	Gender    string
	Comment   string
	WhereKnow string
	IsChild   bool
}

type RevenueRecord struct {
	ID          int
	Date        string
	Name        string
	Amount      float64
	Type        int
	PaymentType string
	Comment     string
}

// ── interface ──────────────────────────────────────────────────────────────────

// Scheduler is the backend-agnostic scheduling interface.
// LocalScheduler (local.go) and Service (macdent.go) both implement it.
// Registry wraps both and dispatches based on clinic.scheduler_type.
type Scheduler interface {
	ListDoctors(ctx context.Context, clinicID uuid.UUID) ([]Doctor, error)
	GetDoctor(ctx context.Context, clinicID uuid.UUID, id string) (*Doctor, error)
	ListPatients(ctx context.Context, clinicID uuid.UUID) ([]Patient, error)
	GetPatient(ctx context.Context, clinicID uuid.UUID, id int) (*Patient, error)
	CreatePatient(ctx context.Context, clinicID uuid.UUID, p CreatePatientParams) (*Patient, error)
	GetClinic(ctx context.Context, clinicID uuid.UUID) (*Stomatology, error)
	ListAppointments(ctx context.Context, clinicID uuid.UUID, from, to time.Time) (*AppointmentsResponse, error)
	GetAppointmentByID(ctx context.Context, clinicID uuid.UUID, id int) (*AppointmentDetail, error)
	CreateAppointment(ctx context.Context, req BookRequest) (*BookResult, error)
	CreateScheduleAppointment(ctx context.Context, clinicID uuid.UUID, p ScheduleAppointmentParams) (*ScheduleAppointmentResult, error)
	UpdateAppointment(ctx context.Context, clinicID uuid.UUID, id int, p UpdateAppointmentParams) error
	RemoveAppointment(ctx context.Context, clinicID uuid.UUID, id int) error
	SetAppointmentStatus(ctx context.Context, clinicID uuid.UUID, id, status int) error
	SendAppointmentRequest(ctx context.Context, clinicID uuid.UUID, p AppointmentRequestParams) (*AppointmentRequestResult, error)
	GetFreeSlots(ctx context.Context, clinicID uuid.UUID, from, to time.Time, specialty string) ([]Slot, error)
	GetRevenue(ctx context.Context, clinicID uuid.UUID, from, to time.Time) ([]RevenueRecord, error)
	GetHistory(ctx context.Context, clinicID uuid.UUID, from, to time.Time) (*AppointmentsResponse, error)
}
