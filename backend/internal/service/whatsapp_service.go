// Package service contains the application's business logic.
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/dentdesk/backend/internal/domain"
	"github.com/dentdesk/backend/internal/llm"
	"github.com/dentdesk/backend/internal/repository"
	"github.com/dentdesk/backend/internal/whatsapp"
)

// InboundMessageRequest carries parsed data from the webhook handler to the service.
type InboundMessageRequest struct {
	PhoneNumberID string
	From          string // sender phone (E.164 without +)
	SenderName    string
	WaMessageID   string
	MessageType   string
	Body          string
	ReceivedAt    time.Time
}

// WhatsAppService orchestrates inbound/outbound WhatsApp messaging.
type WhatsAppService struct {
	msgRepo   repository.MessageLogRepository
	patRepo   repository.PatientRepository
	clinicRepo repository.ClinicRepository
	waClient  *whatsapp.Client
	llmClient *llm.Client
}

// NewWhatsAppService creates a new WhatsAppService.
func NewWhatsAppService(
	msgRepo repository.MessageLogRepository,
	patRepo repository.PatientRepository,
	clinicRepo repository.ClinicRepository,
	waClient *whatsapp.Client,
	llmClient *llm.Client,
) *WhatsAppService {
	return &WhatsAppService{
		msgRepo:    msgRepo,
		patRepo:    patRepo,
		clinicRepo: clinicRepo,
		waClient:   waClient,
		llmClient:  llmClient,
	}
}

// HandleInbound processes an incoming WhatsApp message end-to-end:
//  1. Resolve tenant from phone_number_id
//  2. Find or create patient record
//  3. Persist the message log
//  4. Optionally generate an LLM reply
func (s *WhatsAppService) HandleInbound(ctx context.Context, req InboundMessageRequest) error {
	// 1. Resolve which clinic owns this phone number
	clinic, err := s.clinicRepo.FindByPhoneNumberID(ctx, req.PhoneNumberID)
	if err != nil {
		return fmt.Errorf("resolve clinic for phone_number_id %s: %w", req.PhoneNumberID, err)
	}

	// 2. Find or create patient by phone number
	patient, err := s.patRepo.FindByPhone(ctx, clinic.ID, req.From)
	if err != nil {
		// Auto-create a minimal patient record for the new contact
		patient = &domain.Patient{
			ClinicID:  clinic.ID,
			Phone:     req.From,
			FirstName: req.SenderName,
			LastName:  "",
		}
		if createErr := s.patRepo.Create(ctx, patient); createErr != nil {
			log.Error().Err(createErr).Str("phone", req.From).Msg("failed to auto-create patient")
			// Non-fatal: continue without patient linkage
			patient = nil
		}
	}

	// 3. Persist inbound message
	msgLog := &domain.MessageLog{
		ClinicID:          clinic.ID,
		WhatsAppMessageID: req.WaMessageID,
		FromPhone:         req.From,
		ToPhone:           req.PhoneNumberID,
		Direction:         domain.MessageDirectionInbound,
		Status:            domain.MessageStatusDelivered,
		MessageType:       req.MessageType,
		Body:              req.Body,
	}
	if patient != nil {
		patID := patient.ID
		msgLog.PatientID = &patID
	}

	if err := s.msgRepo.Create(ctx, msgLog); err != nil {
		return fmt.Errorf("persist inbound message log: %w", err)
	}

	log.Info().
		Str("clinic_id", clinic.ID.String()).
		Str("wamid", req.WaMessageID).
		Msg("inbound message stored")

	// 4. LLM auto-reply (fire-and-forget with its own context)
	// Only process text messages — skip media for now
	if req.MessageType == "text" && req.Body != "" {
		go s.generateAndSendReply(context.Background(), clinic, patient, req)
	}

	return nil
}

// generateAndSendReply asks the LLM for a reply and sends it via WhatsApp.
func (s *WhatsAppService) generateAndSendReply(
	ctx context.Context,
	clinic *domain.Clinic,
	patient *domain.Patient,
	req InboundMessageRequest,
) {
	patientName := req.SenderName
	if patient != nil {
		patientName = patient.FullName()
	}

	reply, err := s.llmClient.GenerateReply(ctx, llm.ReplyRequest{
		ClinicName:  clinic.Name,
		PatientName: patientName,
		UserMessage: req.Body,
	})
	if err != nil {
		log.Error().Err(err).Msg("LLM reply generation failed")
		return
	}

	resp, err := s.waClient.SendText(ctx, req.From, reply)
	if err != nil {
		log.Error().Err(err).Str("to", req.From).Msg("failed to send WhatsApp reply")
		return
	}

	// Persist outbound message
	wamid := ""
	if len(resp.Messages) > 0 {
		wamid = resp.Messages[0].ID
	}

	outLog := &domain.MessageLog{
		ClinicID:          clinic.ID,
		WhatsAppMessageID: wamid,
		FromPhone:         req.PhoneNumberID,
		ToPhone:           req.From,
		Direction:         domain.MessageDirectionOutbound,
		Status:            domain.MessageStatusSent,
		MessageType:       "text",
		Body:              reply,
		LLMUsed:           true,
	}
	if patient != nil {
		pid := patient.ID
		outLog.PatientID = &pid
	}

	if err := s.msgRepo.Create(ctx, outLog); err != nil {
		log.Error().Err(err).Msg("failed to persist outbound message log")
	}
}

// HandleStatusUpdate updates the delivery status of an outbound message.
func (s *WhatsAppService) HandleStatusUpdate(ctx context.Context, wamid, status string) error {
	msgStatus := domain.MessageStatus(status)
	return s.msgRepo.UpdateStatus(ctx, wamid, msgStatus)
}

// SendMessage allows staff to manually send a WhatsApp message to a patient.
func (s *WhatsAppService) SendMessage(
	ctx context.Context,
	clinicID uuid.UUID,
	patientID uuid.UUID,
	body string,
) error {
	patient, err := s.patRepo.FindByID(ctx, clinicID, patientID)
	if err != nil {
		return fmt.Errorf("find patient: %w", err)
	}

	resp, err := s.waClient.SendText(ctx, patient.Phone, body)
	if err != nil {
		return fmt.Errorf("send WhatsApp message: %w", err)
	}

	wamid := ""
	if len(resp.Messages) > 0 {
		wamid = resp.Messages[0].ID
	}

	log := &domain.MessageLog{
		ClinicID:          clinicID,
		PatientID:         &patientID,
		WhatsAppMessageID: wamid,
		FromPhone:         "", // populated from config in production
		ToPhone:           patient.Phone,
		Direction:         domain.MessageDirectionOutbound,
		Status:            domain.MessageStatusSent,
		MessageType:       "text",
		Body:              body,
	}
	return s.msgRepo.Create(ctx, log)
}
