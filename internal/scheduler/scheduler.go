// Package scheduler is the scheduling layer of DentDesk.
//
// Phase 1.5 note: handlers and services depend on the Scheduler contract and
// Registry dispatches to the configured backend for each clinic. MacDent-backed
// clinics still read through MacDent live; local/mock clinics read from
// PostgreSQL repos. Long-term, MacDent clinics should become local-first via a
// sync worker without changing this package's public domain types.
package scheduler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/dentdesk/dentdesk/internal/integrations/macdent"
)

// ── domain types ─────────────────────────────────────────────────────────────

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

type RevenueRecord struct {
	ID          int
	Date        string
	Name        string
	Amount      float64
	Type        int
	PaymentType string
	Comment     string
}

// ── interface ─────────────────────────────────────────────────────────────────

// Scheduler is the backend-agnostic scheduling interface.
// Service (MacDent) and LocalScheduler both implement it.
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

// ── service ──────────────────────────────────────────────────────────────────

// Service is the MacDent-backed scheduler implementation. The public Scheduler
// contract above stays in DentDesk domain types; MacDent DTOs are translated at
// this boundary.
type Service struct {
	db   *sqlx.DB
	http *http.Client
}

func NewService(db *sqlx.DB) *Service {
	return &Service{
		db:   db,
		http: &http.Client{Timeout: 15 * time.Second},
	}
}

func (s *Service) macdent(ctx context.Context, clinicID uuid.UUID) (*macdent.Client, error) {
	return macdent.ClientFor(ctx, s.db, s.http, clinicID)
}

// ── doctors ──────────────────────────────────────────────────────────────────

func (s *Service) ListDoctors(ctx context.Context, clinicID uuid.UUID) ([]Doctor, error) {
	c, err := s.macdent(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	mds, err := c.ListDoctors(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Doctor, 0, len(mds))
	for _, d := range mds {
		specs := make([]string, 0, len(d.Specialnosti))
		for _, sp := range d.Specialnosti {
			specs = append(specs, sp.Name)
		}
		out = append(out, Doctor{
			ID:          fmt.Sprint(d.ID),
			Name:        d.Name,
			Specialties: specs,
		})
	}
	return out, nil
}

func (s *Service) GetDoctor(ctx context.Context, clinicID uuid.UUID, id string) (*Doctor, error) {
	list, err := s.ListDoctors(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	for _, d := range list {
		if d.ID == id {
			return &d, nil
		}
	}
	return nil, fmt.Errorf("doctor %s not found", id)
}

// ── patients ─────────────────────────────────────────────────────────────────

func (s *Service) ListPatients(ctx context.Context, clinicID uuid.UUID) ([]Patient, error) {
	c, err := s.macdent(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	resp, err := c.ListPatients(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Patient, 0, len(resp.Patients))
	for _, p := range resp.Patients {
		out = append(out, toPatient(p))
	}
	return out, nil
}

func (s *Service) GetPatient(ctx context.Context, clinicID uuid.UUID, id int) (*Patient, error) {
	c, err := s.macdent(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	p, err := c.GetPatientByID(ctx, id)
	if err != nil {
		return nil, err
	}
	out := toPatient(*p)
	return &out, nil
}

func toPatient(p macdent.Patient) Patient {
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

// ── clinic ───────────────────────────────────────────────────────────────────

func (s *Service) GetClinic(ctx context.Context, clinicID uuid.UUID) (*Stomatology, error) {
	c, err := s.macdent(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	stom, err := c.GetClinic(ctx)
	if err != nil {
		return nil, err
	}
	return &Stomatology{ID: stom.ID, Name: stom.Name}, nil
}

// ── appointments (zapis) ─────────────────────────────────────────────────────

func (s *Service) ListAppointments(ctx context.Context, clinicID uuid.UUID, from, to time.Time) (*AppointmentsResponse, error) {
	c, err := s.macdent(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	apps, err := c.GetAppointments(ctx, from, to)
	if err != nil {
		return nil, err
	}
	out := make([]Appointment, 0, len(apps.Appointments))
	for _, a := range apps.Appointments {
		out = append(out, Appointment{
			ID:      a.ID,
			Doctor:  a.Doctor,
			Patient: a.Patient,
			Date:    a.Date,
			Start:   a.Start,
			End:     a.End,
			Status:  a.Status,
			Zhaloba: a.Zhaloba,
			Comment: a.Comment,
			IsFirst: a.IsFirst,
			Cabinet: a.Cabinet,
			Rasp:    a.Rasp,
		})
	}
	return &AppointmentsResponse{Appointments: out}, nil
}

func (s *Service) GetAppointmentByID(ctx context.Context, clinicID uuid.UUID, id int) (*AppointmentDetail, error) {
	c, err := s.macdent(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	detail, err := c.GetAppointmentByID(ctx, id)
	if err != nil {
		return nil, err
	}
	out := toAppointmentDetail(*detail)
	return &out, nil
}

func toAppointmentDetail(a macdent.ZapisDetail) AppointmentDetail {
	return AppointmentDetail{
		ID: a.ID,
		Doctor: AppointmentDoctorRef{
			ID:   a.Doctor.ID,
			Name: a.Doctor.Name,
		},
		Patient: AppointmentPatientRef{
			ID:    a.Patient.ID,
			Name:  a.Patient.Name,
			Phone: a.Patient.Phone,
		},
		Date:    a.Date,
		Start:   a.Start,
		End:     a.End,
		Status:  a.Status,
		Zhaloba: a.Zhaloba,
		Comment: a.Comment,
		IsFirst: a.IsFirst,
		Cabinet: a.Cabinet,
		Rasp:    a.Rasp,
	}
}

func (s *Service) UpdateAppointment(ctx context.Context, clinicID uuid.UUID, id int, p UpdateAppointmentParams) error {
	c, err := s.macdent(ctx, clinicID)
	if err != nil {
		return err
	}
	return c.UpdateAppointment(ctx, id, macdent.UpdateZapisParams{
		DoctorID: p.DoctorID,
		Start:    p.Start,
		End:      p.End,
		Zhaloba:  p.Zhaloba,
		Comment:  p.Comment,
	})
}

func (s *Service) RemoveAppointment(ctx context.Context, clinicID uuid.UUID, id int) error {
	c, err := s.macdent(ctx, clinicID)
	if err != nil {
		return err
	}
	return c.RemoveAppointment(ctx, id)
}

func (s *Service) CreateAppointment(ctx context.Context, req BookRequest) (*BookResult, error) {
	c, err := s.macdent(ctx, req.ClinicID)
	if err != nil {
		return nil, err
	}
	// TODO: resolve MacDent integer doctor/patient IDs from req.DoctorID/req.PatientID (UUID → external_id lookup)
	z, err := c.AddZapis(ctx, macdent.AddZapisParams{
		Start:   req.StartsAt,
		End:     req.EndsAt,
		Zhaloba: req.Service,
	})
	if err != nil {
		return nil, err
	}
	ext := fmt.Sprint(z.ID)
	return &BookResult{AppointmentID: uuid.New(), ExternalID: &ext}, nil
}

// ── direct MacDent creation (integer IDs) ────────────────────────────────────

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

func (s *Service) CreatePatient(ctx context.Context, clinicID uuid.UUID, p CreatePatientParams) (*Patient, error) {
	c, err := s.macdent(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	mp, err := c.AddPatient(ctx, macdent.AddPatientParams{
		Name:      p.Name,
		Phone:     p.Phone,
		IIN:       p.IIN,
		Birth:     p.Birth,
		Gender:    p.Gender,
		Comment:   p.Comment,
		WhereKnow: p.WhereKnow,
		IsChild:   p.IsChild,
	})
	if err != nil {
		return nil, err
	}
	out := toPatient(*mp)
	return &out, nil
}

func (s *Service) CreateScheduleAppointment(ctx context.Context, clinicID uuid.UUID, p ScheduleAppointmentParams) (*ScheduleAppointmentResult, error) {
	c, err := s.macdent(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	z, err := c.AddZapis(ctx, macdent.AddZapisParams{
		DoctorID:  p.DoctorID,
		PatientID: p.PatientID,
		Start:     p.Start,
		End:       p.End,
		Zhaloba:   p.Zhaloba,
		Cabinet:   p.Cabinet,
		IsFirst:   p.IsFirst,
	})
	if err != nil {
		return nil, err
	}
	return &ScheduleAppointmentResult{ID: z.ID}, nil
}

func (s *Service) SetAppointmentStatus(ctx context.Context, clinicID uuid.UUID, id, status int) error {
	c, err := s.macdent(ctx, clinicID)
	if err != nil {
		return err
	}
	return c.SetStatus(ctx, id, status)
}

func (s *Service) SendAppointmentRequest(ctx context.Context, clinicID uuid.UUID, p AppointmentRequestParams) (*AppointmentRequestResult, error) {
	c, err := s.macdent(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	res, err := c.SendAppointmentRequest(ctx, macdent.AppointmentRequest{
		PatientName:  p.PatientName,
		PatientPhone: p.PatientPhone,
		Start:        p.Start,
		End:          p.End,
		WhereKnow:    p.WhereKnow,
	})
	if err != nil {
		return nil, err
	}
	return &AppointmentRequestResult{ID: res.ID}, nil
}

// ── rashodi ──────────────────────────────────────────────────────────────────

func (s *Service) GetRevenue(ctx context.Context, clinicID uuid.UUID, from, to time.Time) ([]RevenueRecord, error) {
	c, err := s.macdent(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	resp, err := c.GetRashodi(ctx, from, to)
	if err != nil {
		return nil, err
	}
	out := make([]RevenueRecord, 0, len(resp.Rashodi))
	for _, r := range resp.Rashodi {
		out = append(out, RevenueRecord{
			ID:          r.ID,
			Date:        r.Date,
			Name:        r.Name,
			Amount:      r.SummFloat(),
			Type:        r.Type,
			PaymentType: r.TypeOplata,
			Comment:     r.Comment,
		})
	}
	return out, nil
}

// ── slots ────────────────────────────────────────────────────────────────────

func (s *Service) GetFreeSlots(
	ctx context.Context, clinicID uuid.UUID, from, to time.Time, specialty string,
) ([]Slot, error) {
	c, err := s.macdent(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	doctors, err := c.ListDoctors(ctx)
	if err != nil {
		return nil, err
	}
	var slots []Slot
	for _, doc := range doctors {
		schedules, err := c.GetFreeTime(ctx, doc.ID, from, to)
		if err != nil {
			continue
		}
		for _, iv := range schedules {
			start, err1 := time.Parse(macdent.DateLayout, iv.From)
			end, err2 := time.Parse(macdent.DateLayout, iv.To)
			if err1 != nil || err2 != nil {
				continue
			}
			for cur := start; !cur.Add(30 * time.Minute).After(end); cur = cur.Add(30 * time.Minute) {
				slots = append(slots, Slot{
					StartsAt: cur,
					EndsAt:   cur.Add(30 * time.Minute),
					DoctorID: uuid.Nil,
					Doctor:   doc.Name,
				})
			}
		}
	}
	return slots, nil
}

type HistoryRecord struct {
	Date          time.Time
	Birthday      *time.Time
	Amount        float64
	PatientName   string
	PhoneNumber   string
	ServiceType   string
	PaymentMethod string
	DoctorName    string
}

type HistoryResponse struct {
	Records []HistoryRecord
}

// GetHistory returns the historical appointments for a clinic over a date range.
// For now this is the same dataset as ListAppointments (live MacDent zapis/find
// already includes completed and cancelled records). When payment data is
// integrated, enrich the records here.
func (s *Service) GetHistory(ctx context.Context, clinicID uuid.UUID, from, to time.Time) (*AppointmentsResponse, error) {
	return s.ListAppointments(ctx, clinicID, from, to)
}
