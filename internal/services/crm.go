package services

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/dentdesk/dentdesk/internal/realtime"
	"github.com/dentdesk/dentdesk/internal/store"
	"github.com/dentdesk/dentdesk/internal/whatsapp"
)

var ErrConversationNotFound = errors.New("conversation not found")

type ChatListRow struct {
	Conversation store.Conversation `json:"conversation"`
	Patient      *store.Patient     `json:"patient"`
	LastMessage  *store.Message     `json:"last_message,omitempty"`
}

type Stats struct {
	ActiveChats   int `json:"active_chats"`
	TodayAppts    int `json:"today_appts"`
	TotalPatients int `json:"total_patients"`
}

type CRMService struct {
	DB            *sqlx.DB
	Patients      *store.PatientRepo
	Conversations *store.ConversationRepo
	Appointments  *store.AppointmentRepo
	Doctors       *store.DoctorRepo
	Hub           *realtime.Hub
	WhatsApp      *whatsapp.Client
}

func NewCRMService(db *sqlx.DB, patientsRepo *store.PatientRepo, convRepo *store.ConversationRepo, apptRepo *store.AppointmentRepo, doctorsRepo *store.DoctorRepo, hub *realtime.Hub, wa *whatsapp.Client) *CRMService {
	return &CRMService{
		DB:            db,
		Patients:      patientsRepo,
		Conversations: convRepo,
		Appointments:  apptRepo,
		Doctors:       doctorsRepo,
		Hub:           hub,
		WhatsApp:      wa,
	}
}

func (s *CRMService) ListChats(ctx context.Context, clinicID uuid.UUID) ([]ChatListRow, error) {
	convs, err := s.Conversations.ListForClinic(ctx, clinicID, 100)
	if err != nil {
		return nil, err
	}
	out := make([]ChatListRow, 0, len(convs))
	for _, conv := range convs {
		p, _ := s.Patients.Get(ctx, conv.PatientID)
		msgs, _ := s.Conversations.ListMessages(ctx, conv.ID, 1)
		var last *store.Message
		if len(msgs) > 0 {
			last = &msgs[0]
		}
		out = append(out, ChatListRow{Conversation: conv, Patient: p, LastMessage: last})
	}
	return out, nil
}

func (s *CRMService) ListMessages(ctx context.Context, conversationID uuid.UUID) ([]store.Message, error) {
	return s.Conversations.RecentHistory(ctx, conversationID, 200)
}

func (s *CRMService) OperatorSend(ctx context.Context, clinicID, operatorID, conversationID uuid.UUID, body string) (*store.Message, error) {
	convs, _ := s.Conversations.ListForClinic(ctx, clinicID, 500)
	var conv *store.Conversation
	for i := range convs {
		if convs[i].ID == conversationID {
			conv = &convs[i]
			break
		}
	}
	if conv == nil {
		return nil, ErrConversationNotFound
	}
	p, err := s.Patients.Get(ctx, conv.PatientID)
	if err != nil {
		return nil, err
	}

	meta, _ := json.Marshal(map[string]any{"operator_id": operatorID})
	msg, _, err := s.Conversations.InsertMessage(ctx, &store.Message{
		ConversationID: conversationID,
		Direction:      "outbound",
		Sender:         "operator",
		Body:           body,
		Meta:           meta,
	})
	if err != nil {
		return nil, err
	}
	_ = s.Conversations.SetStatus(ctx, conversationID, "handoff")

	go func(phone, text string) {
		msgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = s.WhatsApp.SendText(msgCtx, phone, text)
	}(p.Phone, body)

	s.Hub.Publish(clinicID, "message", msg)
	return msg, nil
}

func (s *CRMService) ReleaseHandoff(ctx context.Context, conversationID uuid.UUID) error {
	return s.Conversations.SetStatus(ctx, conversationID, "active")
}

func (s *CRMService) ListPatients(ctx context.Context, clinicID uuid.UUID) ([]store.Patient, error) {
	return s.Patients.List(ctx, clinicID, 200)
}

func (s *CRMService) PatientAppointments(ctx context.Context, patientID uuid.UUID) ([]store.Appointment, error) {
	return s.Appointments.ListForPatient(ctx, patientID)
}

func (s *CRMService) Calendar(ctx context.Context, clinicID uuid.UUID, from, to *time.Time) ([]store.Appointment, error) {
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now().Add(7 * 24 * time.Hour)
	if from != nil && to != nil {
		start = *from
		end = *to
	}
	return s.Appointments.ListForPeriod(ctx, clinicID, start, end)
}

func (s *CRMService) ListDoctors(ctx context.Context, clinicID uuid.UUID) ([]store.Doctor, error) {
	return s.Doctors.List(ctx, clinicID)
}

func (s *CRMService) Stats(ctx context.Context, clinicID uuid.UUID) (Stats, error) {
	var out Stats
	_ = s.DB.GetContext(ctx, &out.ActiveChats,
		`SELECT COUNT(*) FROM conversations WHERE clinic_id=$1 AND status='active'
		   AND last_message_at > NOW() - INTERVAL '24 hours'`, clinicID)
	_ = s.DB.GetContext(ctx, &out.TodayAppts,
		`SELECT COUNT(*) FROM appointments WHERE clinic_id=$1
		   AND starts_at::date = CURRENT_DATE AND status IN ('scheduled','confirmed')`, clinicID)
	_ = s.DB.GetContext(ctx, &out.TotalPatients,
		`SELECT COUNT(*) FROM patients WHERE clinic_id=$1`, clinicID)
	return out, nil
}
