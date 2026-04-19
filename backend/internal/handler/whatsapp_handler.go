package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/dentdesk/backend/internal/middleware"
	"github.com/dentdesk/backend/internal/service"
	"github.com/dentdesk/backend/internal/whatsapp"
)

// WhatsAppHandler handles incoming WhatsApp Cloud API webhook events.
type WhatsAppHandler struct {
	waService   *service.WhatsAppService
	verifyToken string
}

// NewWhatsAppHandler creates a new WhatsApp webhook handler.
func NewWhatsAppHandler(waService *service.WhatsAppService, verifyToken string) *WhatsAppHandler {
	return &WhatsAppHandler{
		waService:   waService,
		verifyToken: verifyToken,
	}
}

// Verify godoc
// @Summary Verify WhatsApp webhook
// @Description Meta webhook verification challenge endpoint.
// @Tags webhook
// @Produce plain
// @Param hub.mode query string true "Verification mode"
// @Param hub.verify_token query string true "Verification token"
// @Param hub.challenge query string true "Challenge string"
// @Success 200 {string} string
// @Failure 403 {object} ErrorResponse
// @Router /webhook/whatsapp [get]
func (h *WhatsAppHandler) Verify(c *gin.Context) {
	mode := c.Query("hub.mode")
	token := c.Query("hub.verify_token")
	challenge := c.Query("hub.challenge")

	if mode == "subscribe" && token == h.verifyToken {
		log.Info().Msg("WhatsApp webhook verification successful")
		c.String(http.StatusOK, challenge)
		return
	}

	log.Warn().
		Str("mode", mode).
		Str("token_provided", token).
		Msg("WhatsApp webhook verification failed")

	c.JSON(http.StatusForbidden, gin.H{"error": "verification failed"})
}

// Receive godoc
// @Summary Receive WhatsApp webhook events
// @Description Accepts inbound WhatsApp events from Meta and acknowledges immediately.
// @Tags webhook
// @Accept json
// @Produce json
// @Param payload body whatsapp.WebhookPayload true "Webhook payload"
// @Success 200
// @Router /webhook/whatsapp [post]
func (h *WhatsAppHandler) Receive(c *gin.Context) {
	var payload whatsapp.WebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		log.Error().Err(err).Msg("failed to parse WhatsApp webhook payload")
		// Always return 200 to Meta, otherwise it retries indefinitely
		c.Status(http.StatusOK)
		return
	}

	// Always acknowledge immediately — process asynchronously to avoid timeouts
	c.Status(http.StatusOK)

	// Process each entry in the background
	go h.processPayload(c.Copy(), payload)
}

type SendManualRequest struct {
	PatientID string `json:"patient_id" binding:"required"`
	Body      string `json:"body" binding:"required"`
}

// SendManual godoc
// @Summary Send manual WhatsApp message
// @Description Sends a staff-initiated WhatsApp message to a patient.
// @Tags messages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body SendManualRequest true "Manual message payload"
// @Success 200 {object} StatusResponse
// @Failure 400 {object} ErrorResponse
// @Router /messages/send [post]
func (h *WhatsAppHandler) SendManual(c *gin.Context) {
	var req SendManualRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	patientID, err := uuid.Parse(req.PatientID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid patient_id"})
		return
	}

	body := strings.TrimSpace(req.Body)
	if body == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body is required"})
		return
	}

	clinicID := middleware.ClinicIDFromCtx(c)
	if err := h.waService.SendMessage(c.Request.Context(), clinicID, patientID, body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "sent"})
}

// processPayload iterates over all entries/changes and dispatches accordingly.
func (h *WhatsAppHandler) processPayload(c *gin.Context, payload whatsapp.WebhookPayload) {
	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			if change.Field != "messages" {
				continue
			}

			val := change.Value

			// ── Inbound messages ─────────────────────────────
			for _, msg := range val.Messages {
				h.handleInboundMessage(c, val.Metadata, msg, val.Contacts)
			}

			// ── Status updates (delivered / read / failed) ───
			for _, status := range val.Statuses {
				h.handleStatusUpdate(c, status)
			}
		}
	}
}

func (h *WhatsAppHandler) handleInboundMessage(
	c *gin.Context,
	meta whatsapp.Metadata,
	msg whatsapp.InboundMessage,
	contacts []whatsapp.Contact,
) {
	senderName := resolveContactName(contacts, msg.From)
	body := resolveMessageBody(msg)
	ts := parseTimestamp(msg.Timestamp)

	logger := log.With().
		Str("from", msg.From).
		Str("type", msg.Type).
		Str("wamid", msg.ID).
		Logger()

	logger.Info().Str("body_preview", truncate(body, 80)).Msg("inbound WhatsApp message")

	req := service.InboundMessageRequest{
		PhoneNumberID: meta.PhoneNumberID,
		From:          msg.From,
		SenderName:    senderName,
		WaMessageID:   msg.ID,
		MessageType:   msg.Type,
		Body:          body,
		ReceivedAt:    ts,
	}

	if err := h.waService.HandleInbound(c.Request.Context(), req); err != nil {
		logger.Error().Err(err).Msg("error handling inbound message")
	}
}

func (h *WhatsAppHandler) handleStatusUpdate(c *gin.Context, status whatsapp.StatusUpdate) {
	log.Debug().
		Str("wamid", status.ID).
		Str("status", status.Status).
		Str("recipient", status.RecipientID).
		Msg("WhatsApp status update")

	if err := h.waService.HandleStatusUpdate(c.Request.Context(), status.ID, status.Status); err != nil {
		log.Error().Err(err).Str("wamid", status.ID).Msg("error handling status update")
	}
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// resolveMessageBody extracts text from any supported message type.
func resolveMessageBody(msg whatsapp.InboundMessage) string {
	switch msg.Type {
	case "text":
		if msg.Text != nil {
			return msg.Text.Body
		}
	case "image":
		if msg.Image != nil {
			return "[image] " + msg.Image.Caption
		}
	case "audio":
		return "[audio message]"
	case "document":
		if msg.Document != nil {
			return "[document] " + msg.Document.Caption
		}
	}
	return ""
}

// resolveContactName finds the display name from the contacts slice.
func resolveContactName(contacts []whatsapp.Contact, phone string) string {
	for _, c := range contacts {
		if c.WaID == phone {
			return c.Profile.Name
		}
	}
	return ""
}

func parseTimestamp(ts string) time.Time {
	unix, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return time.Now().UTC()
	}
	return time.Unix(unix, 0).UTC()
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
