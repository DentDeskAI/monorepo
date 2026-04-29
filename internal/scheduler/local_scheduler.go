package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/dentdesk/dentdesk/internal/appointments"
	"github.com/dentdesk/dentdesk/internal/clinics"
	"github.com/dentdesk/dentdesk/internal/doctors"
	"github.com/dentdesk/dentdesk/internal/patients"
)

// localStatusToInt maps local DB appointment status strings to Macdent integer codes.
// The frontend/dashboard expect these integer codes.
var localStatusToInt = map[string]int{
	"scheduled": 0,
	"confirmed": 1,
	"cancelled": 2,
	"completed": 4,
	"no_show":   2,
}

// intToLocalStatus is the reverse mapping used by SetAppointmentStatus.
var intToLocalStatus = map[int]string{
	0: "scheduled",
	1: "confirmed",
	2: "cancelled",
	3: "completed", // "came" → completed
	4: "completed", // "left" → completed
	5: "confirmed", // "in_process" → confirmed
	6: "scheduled", // "late" → keep scheduled
}

const macdentDateLayout = "02.01.2006 15:04:05"

// LocalScheduler implements Scheduler using local PostgreSQL tables.
// Used for clinics with scheduler_type = 'local' or 'mock'.
type LocalScheduler struct {
	doctorsR  *doctors.Repo
	patientsR *patients.Repo
	apptR     *appointments.Repo
	clinicsR  *clinics.Repo
}

// ── doctors ──────────────────────────────────────────────────────────────────

func (s *LocalScheduler) ListDoctors(ctx context.Context, clinicID uuid.UUID) ([]Doctor, error) {
	list, err := s.doctorsR.List(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	out := make([]Doctor, 0, len(list))
	for _, d := range list {
		out = append(out, localDoctorToScheduler(d))
	}
	return out, nil
}

func (s *LocalScheduler) GetDoctor(ctx context.Context, clinicID uuid.UUID, id string) (*Doctor, error) {
	seqID := 0
	if _, err := fmt.Sscanf(id, "%d", &seqID); err != nil {
		return nil, fmt.Errorf("doctor id must be an integer, got %q", id)
	}
	d, err := s.doctorsR.GetBySeqID(ctx, clinicID, seqID)
	if err != nil {
		return nil, fmt.Errorf("doctor %s not found", id)
	}
	out := localDoctorToScheduler(*d)
	return &out, nil
}

func localDoctorToScheduler(d doctors.Doctor) Doctor {
	specs := []string{}
	if d.Specialty != nil && *d.Specialty != "" {
		specs = append(specs, *d.Specialty)
	}
	return Doctor{
		ID:          fmt.Sprint(d.SeqID),
		Name:        d.Name,
		Specialties: specs,
	}
}

// ── patients ─────────────────────────────────────────────────────────────────

func (s *LocalScheduler) ListPatients(ctx context.Context, clinicID uuid.UUID) ([]Patient, error) {
	list, err := s.patientsR.List(ctx, clinicID, 500)
	if err != nil {
		return nil, err
	}
	out := make([]Patient, 0, len(list))
	for _, p := range list {
		out = append(out, localPatientToScheduler(p))
	}
	return out, nil
}

func (s *LocalScheduler) GetPatient(ctx context.Context, clinicID uuid.UUID, id int) (*Patient, error) {
	p, err := s.patientsR.GetBySeqID(ctx, clinicID, id)
	if err != nil {
		return nil, fmt.Errorf("patient %d not found", id)
	}
	out := localPatientToScheduler(*p)
	return &out, nil
}

func (s *LocalScheduler) CreatePatient(ctx context.Context, clinicID uuid.UUID, p CreatePatientParams) (*Patient, error) {
	created, err := s.patientsR.GetOrCreateByPhone(ctx, clinicID, p.Phone)
	if err != nil {
		return nil, err
	}
	if p.Name != "" {
		if err := s.patientsR.UpdateName(ctx, created.ID, p.Name); err != nil {
			return nil, err
		}
		created.Name = &p.Name
	}
	out := localPatientToScheduler(*created)
	return &out, nil
}

func localPatientToScheduler(p patients.Patient) Patient {
	name := ""
	if p.Name != nil {
		name = *p.Name
	}
	return Patient{
		ID:        p.SeqID,
		Name:      name,
		Phone:     &p.Phone,
		Number:    fmt.Sprint(p.SeqID),
		Gender:    nil,
		IIN:       nil,
		Birth:     nil,
		IsChild:   false,
		Comment:   "",
		WhereKnow: "",
	}
}

// ── clinic ───────────────────────────────────────────────────────────────────

func (s *LocalScheduler) GetClinic(ctx context.Context, clinicID uuid.UUID) (*Stomatology, error) {
	c, err := s.clinicsR.Get(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	return &Stomatology{ID: c.ID.String(), Name: c.Name}, nil
}

// ── appointments ─────────────────────────────────────────────────────────────

func (s *LocalScheduler) ListAppointments(ctx context.Context, clinicID uuid.UUID, from, to time.Time) (*AppointmentsResponse, error) {
	list, err := s.apptR.ListForPeriod(ctx, clinicID, from, to)
	if err != nil {
		return nil, err
	}
	out := make([]Appointment, 0, len(list))
	for _, a := range list {
		out = append(out, localApptToScheduler(a, s))
	}
	return &AppointmentsResponse{Appointments: out}, nil
}

func (s *LocalScheduler) GetAppointmentByID(ctx context.Context, clinicID uuid.UUID, id int) (*AppointmentDetail, error) {
	a, err := s.apptR.GetBySeqID(ctx, clinicID, id)
	if err != nil {
		return nil, fmt.Errorf("appointment %d not found", id)
	}

	doctorSeqID := 0
	doctorName := ""
	if a.DoctorID != nil {
		d, err2 := s.doctorsR.Get(ctx, *a.DoctorID)
		if err2 == nil {
			doctorSeqID = d.SeqID
			doctorName = d.Name
		}
	}

	patientSeqID := 0
	patientName := ""
	var patientPhone *string
	p, err2 := s.patientsR.Get(ctx, a.PatientID)
	if err2 == nil {
		patientSeqID = p.SeqID
		patientName = ""
		if p.Name != nil {
			patientName = *p.Name
		}
		patientPhone = &p.Phone
	}

	service := ""
	if a.Service != nil {
		service = *a.Service
	}

	return &AppointmentDetail{
		ID:      a.SeqID,
		Date:    a.StartsAt.Format("02.01.2006"),
		Start:   a.StartsAt.Format(macdentDateLayout),
		End:     a.EndsAt.Format(macdentDateLayout),
		Status:  localStatusToInt[a.Status],
		Zhaloba: service,
		Doctor: AppointmentDoctorRef{
			ID:   doctorSeqID,
			Name: doctorName,
		},
		Patient: AppointmentPatientRef{
			ID:    patientSeqID,
			Name:  patientName,
			Phone: patientPhone,
		},
	}, nil
}

func (s *LocalScheduler) CreateAppointment(ctx context.Context, req BookRequest) (*BookResult, error) {
	a, err := s.apptR.Create(ctx, &appointments.Appointment{
		ClinicID:  req.ClinicID,
		PatientID: req.PatientID,
		DoctorID:  &req.DoctorID,
		ChairID:   req.ChairID,
		StartsAt:  req.StartsAt,
		EndsAt:    req.EndsAt,
		Service:   &req.Service,
		Status:    "scheduled",
		Source:    "operator",
	})
	if err != nil {
		return nil, err
	}
	extID := fmt.Sprint(a.SeqID)
	return &BookResult{AppointmentID: a.ID, ExternalID: &extID}, nil
}

func (s *LocalScheduler) CreateScheduleAppointment(ctx context.Context, clinicID uuid.UUID, p ScheduleAppointmentParams) (*ScheduleAppointmentResult, error) {
	doc, err := s.doctorsR.GetBySeqID(ctx, clinicID, p.DoctorID)
	if err != nil {
		return nil, fmt.Errorf("doctor %d not found", p.DoctorID)
	}
	pat, err := s.patientsR.GetBySeqID(ctx, clinicID, p.PatientID)
	if err != nil {
		return nil, fmt.Errorf("patient %d not found", p.PatientID)
	}
	service := p.Zhaloba
	a, err := s.apptR.Create(ctx, &appointments.Appointment{
		ClinicID:  clinicID,
		PatientID: pat.ID,
		DoctorID:  &doc.ID,
		StartsAt:  p.Start,
		EndsAt:    p.End,
		Service:   &service,
		Status:    "scheduled",
		Source:    "operator",
	})
	if err != nil {
		return nil, err
	}
	return &ScheduleAppointmentResult{ID: a.SeqID}, nil
}

func (s *LocalScheduler) UpdateAppointment(ctx context.Context, clinicID uuid.UUID, id int, p UpdateAppointmentParams) error {
	a, err := s.apptR.GetBySeqID(ctx, clinicID, id)
	if err != nil {
		return fmt.Errorf("appointment %d not found", id)
	}
	f := appointments.UpdateFields{
		StartsAt: p.Start,
		EndsAt:   p.End,
	}
	if p.DoctorID != nil {
		doc, err2 := s.doctorsR.GetBySeqID(ctx, clinicID, *p.DoctorID)
		if err2 != nil {
			return fmt.Errorf("doctor %d not found", *p.DoctorID)
		}
		f.DoctorID = &doc.ID
	}
	if p.Zhaloba != nil {
		f.Service = p.Zhaloba
	} else if p.Comment != nil {
		f.Service = p.Comment
	}
	return s.apptR.UpdateFields(ctx, a.ID, f)
}

func (s *LocalScheduler) RemoveAppointment(ctx context.Context, clinicID uuid.UUID, id int) error {
	a, err := s.apptR.GetBySeqID(ctx, clinicID, id)
	if err != nil {
		return fmt.Errorf("appointment %d not found", id)
	}
	return s.apptR.Delete(ctx, a.ID)
}

func (s *LocalScheduler) SetAppointmentStatus(ctx context.Context, clinicID uuid.UUID, id, status int) error {
	a, err := s.apptR.GetBySeqID(ctx, clinicID, id)
	if err != nil {
		return fmt.Errorf("appointment %d not found", id)
	}
	localStatus, ok := intToLocalStatus[status]
	if !ok {
		localStatus = "scheduled"
	}
	return s.apptR.SetStatus(ctx, a.ID, localStatus)
}

func (s *LocalScheduler) SendAppointmentRequest(ctx context.Context, clinicID uuid.UUID, p AppointmentRequestParams) (*AppointmentRequestResult, error) {
	pat, err := s.patientsR.GetOrCreateByPhone(ctx, clinicID, p.PatientPhone)
	if err != nil {
		return nil, err
	}
	if p.PatientName != "" && (pat.Name == nil || *pat.Name == "") {
		_ = s.patientsR.UpdateName(ctx, pat.ID, p.PatientName)
	}
	a, err := s.apptR.Create(ctx, &appointments.Appointment{
		ClinicID:  clinicID,
		PatientID: pat.ID,
		StartsAt:  p.Start,
		EndsAt:    p.End,
		Status:    "scheduled",
		Source:    "bot",
	})
	if err != nil {
		return nil, err
	}
	return &AppointmentRequestResult{ID: a.SeqID}, nil
}

// ── free slots ────────────────────────────────────────────────────────────────

func (s *LocalScheduler) GetFreeSlots(ctx context.Context, clinicID uuid.UUID, from, to time.Time, specialty string) ([]Slot, error) {
	clinic, err := s.clinicsR.Get(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	slotDur := time.Duration(clinic.SlotDurationMin) * time.Minute
	if slotDur == 0 {
		slotDur = 30 * time.Minute
	}
	wh := parseWorkingHours(clinic.WorkingHours)

	var docList []doctors.Doctor
	if specialty != "" {
		docList, err = s.doctorsR.FindBySpecialty(ctx, clinicID, specialty)
	} else {
		docList, err = s.doctorsR.List(ctx, clinicID)
	}
	if err != nil {
		return nil, err
	}

	booked, err := s.apptR.ListForPeriod(ctx, clinicID, from, to)
	if err != nil {
		return nil, err
	}

	// Index booked intervals by doctor UUID.
	bookedByDoctor := map[uuid.UUID][][2]time.Time{}
	for _, a := range booked {
		if a.DoctorID == nil || (a.Status != "scheduled" && a.Status != "confirmed") {
			continue
		}
		bookedByDoctor[*a.DoctorID] = append(bookedByDoctor[*a.DoctorID], [2]time.Time{a.StartsAt, a.EndsAt})
	}

	var slots []Slot
	for _, doc := range docList {
		for day := floorToDay(from); !day.After(floorToDay(to)); day = day.AddDate(0, 0, 1) {
			hours, ok := wh[weekdayKey(day)]
			if !ok || hours == nil {
				continue
			}
			dayOpen, err1 := dayAt(day, hours.Open)
			dayClose, err2 := dayAt(day, hours.Close)
			if err1 != nil || err2 != nil {
				continue
			}
			for cur := dayOpen; !cur.Add(slotDur).After(dayClose); cur = cur.Add(slotDur) {
				slotEnd := cur.Add(slotDur)
				if slotEnd.Before(from) || cur.After(to) {
					continue
				}
				if slotOverlapsBooked(doc.ID, cur, slotEnd, bookedByDoctor) {
					continue
				}
				slots = append(slots, Slot{
					StartsAt: cur,
					EndsAt:   slotEnd,
					DoctorID: doc.ID,
					Doctor:   doc.Name,
				})
			}
		}
	}
	return slots, nil
}

func slotOverlapsBooked(docID uuid.UUID, start, end time.Time, bookedByDoctor map[uuid.UUID][][2]time.Time) bool {
	for _, iv := range bookedByDoctor[docID] {
		if start.Before(iv[1]) && end.After(iv[0]) {
			return true
		}
	}
	return false
}

// ── rashodi / history ─────────────────────────────────────────────────────────

// GetRevenue returns empty — local clinics have no financial transaction data.
func (s *LocalScheduler) GetRevenue(_ context.Context, _ uuid.UUID, _, _ time.Time) ([]RevenueRecord, error) {
	return nil, nil
}

func (s *LocalScheduler) GetHistory(ctx context.Context, clinicID uuid.UUID, from, to time.Time) (*AppointmentsResponse, error) {
	return s.ListAppointments(ctx, clinicID, from, to)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func localApptToScheduler(a appointments.Appointment, _ *LocalScheduler) Appointment {
	doctorSeqID := 0
	if a.DoctorSeqID != nil {
		doctorSeqID = *a.DoctorSeqID
	}
	patientSeqID := 0
	if a.PatientSeqID != nil {
		patientSeqID = *a.PatientSeqID
	}

	service := ""
	if a.Service != nil {
		service = *a.Service
	}
	status, ok := localStatusToInt[a.Status]
	if !ok {
		status = 0
	}
	return Appointment{
		ID:      a.SeqID,
		Doctor:  doctorSeqID,
		Patient: patientSeqID,
		Date:    a.StartsAt.Format("02.01.2006"),
		Start:   a.StartsAt.Format(macdentDateLayout),
		End:     a.EndsAt.Format(macdentDateLayout),
		Status:  status,
		Zhaloba: service,
		Comment: "",
		IsFirst: false,
		Cabinet: "",
		Rasp:    "",
	}
}
