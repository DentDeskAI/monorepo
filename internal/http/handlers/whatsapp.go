package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"

	"github.com/dentdesk/dentdesk/internal/conversations"
	"github.com/dentdesk/dentdesk/internal/llm"
	"github.com/dentdesk/dentdesk/internal/patients"
	"github.com/dentdesk/dentdesk/internal/realtime"
	"github.com/dentdesk/dentdesk/internal/scheduler"
	"github.com/dentdesk/dentdesk/internal/whatsapp"
)

type WhatsAppHandler struct {
	DB            *sqlx.DB
	Redis         *redis.Client
	Log           zerolog.Logger
	VerifyToken   string
	WhatsApp      *whatsapp.Client
	Patients      *patients.Repo
	Conversations *conversations.Repo
	Orchestrator  *llm.Orchestrator
	Scheduler     scheduler.Scheduler
	Hub           *realtime.Hub
}

// Verify godoc
// @Summary     WhatsApp webhook verification
// @Description Meta calls this to verify the webhook endpoint. Returns hub.challenge on success.
// @Tags        webhook
// @Param       hub.mode         query string true  "Must be 'subscribe'"
// @Param       hub.verify_token query string true  "Secret token configured in Meta dashboard"
// @Param       hub.challenge    query string true  "Challenge string to echo back"
// @Success     200 {string} string "challenge"
// @Failure     403
// @Router      /webhook/whatsapp [get]
func (h *WhatsAppHandler) Verify(c *gin.Context) {
	mode := c.Query("hub.mode")
	token := c.Query("hub.verify_token")
	challenge := c.Query("hub.challenge")
	if mode == "subscribe" && token == h.VerifyToken {
		c.String(http.StatusOK, challenge)
		return
	}
	c.Status(http.StatusForbidden)
}

// Receive godoc
// @Summary     WhatsApp incoming message
// @Description Receives incoming messages from Meta and triggers bot orchestration asynchronously.
// @Tags        webhook
// @Accept      json
// @Success     200
// @Router      /webhook/whatsapp [post]
func (h *WhatsAppHandler) Receive(c *gin.Context) {
	var payload whatsapp.WebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		h.Log.Warn().Err(err).Msg("bad webhook payload")
		c.Status(http.StatusOK) // Meta будет ретраить на 4xx, не хотим
		return
	}
	// Отвечаем Meta сразу — обработку делаем асинхронно (с отдельным контекстом).
	c.Status(http.StatusOK)

	items := payload.Extract()
	for _, m := range items {
		msg := m // capture
		go h.process(msg)
	}
}

func (h *WhatsAppHandler) process(m whatsapp.Extracted) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1) Дедуп: SETNX по message_id в Redis на 10 минут.
	dedupKey := "wa:msg:" + m.MessageID
	ok, err := h.Redis.SetNX(ctx, dedupKey, "1", 10*time.Minute).Result()
	if err == nil && !ok {
		h.Log.Debug().Str("msg_id", m.MessageID).Msg("duplicate webhook, skipping")
		return
	}

	// 2) Находим клинику по phone_number_id.
	type clinicRow struct {
		ID string `db:"id"`
	}
	var row clinicRow
	if err := h.DB.GetContext(ctx, &row,
		`SELECT id FROM clinics WHERE whatsapp_phone_id=$1`, m.PhoneNumberID); err != nil {
		// fallback: единственная клиника (MVP)
		if err := h.DB.GetContext(ctx, &row,
			`SELECT id FROM clinics ORDER BY created_at LIMIT 1`); err != nil {
			h.Log.Error().Err(err).Msg("no clinic for message")
			return
		}
	}
	clinicID, _ := parseUUID(row.ID)

	// 3) Пациент + диалог.
	patient, err := h.Patients.GetOrCreateByPhone(ctx, clinicID, m.From)
	if err != nil {
		h.Log.Error().Err(err).Msg("patient upsert")
		return
	}
	if m.ProfileName != "" && (patient.Name == nil || *patient.Name == "") {
		_ = h.Patients.UpdateName(ctx, patient.ID, m.ProfileName)
	}

	conv, err := h.Conversations.GetOrCreate(ctx, clinicID, patient.ID)
	if err != nil {
		h.Log.Error().Err(err).Msg("conversation")
		return
	}

	// 4) Сохраняем входящее сообщение (идемпотентно).
	wamsgID := m.MessageID
	stored, isNew, err := h.Conversations.InsertMessage(ctx, &conversations.Message{
		ConversationID: conv.ID,
		WAMessageID:    &wamsgID,
		Direction:      "inbound",
		Sender:         "patient",
		Body:           m.Text,
	})
	if err != nil {
		h.Log.Error().Err(err).Msg("insert message")
		return
	}
	if !isNew {
		h.Log.Debug().Msg("already processed")
		return
	}
	h.Hub.Publish(clinicID, "message", stored)

	// 5) Если диалог в handoff — бот молчит, оператор разберётся.
	if conv.Status == "handoff" {
		return
	}

	// 6) Оркестрация LLM.
	history, _ := h.Conversations.RecentHistory(ctx, conv.ID, 12)
	// Убираем последнее inbound (оно уже передаётся отдельно в orchestrator как incoming).
	if n := len(history); n > 0 && history[n-1].ID == stored.ID {
		history = history[:n-1]
	}
	var state llm.ConvState
	_ = json.Unmarshal(conv.Context, &state)

	reply, err := h.Orchestrator.Handle(ctx, clinicID, patient.ID, m.Text, history, state)
	if err != nil {
		h.Log.Error().Err(err).Msg("orchestrate")
		return
	}

	// 7) Сохраняем новый state.
	stateBytes, _ := json.Marshal(reply.NewState)
	if err := h.Conversations.UpdateContext(ctx, conv.ID, stateBytes); err != nil {
		h.Log.Warn().Err(err).Msg("update context")
	}

	// 8) Сохраняем outbound сообщение.
	outMeta := map[string]any{}
	if reply.Meta != nil {
		outMeta = reply.Meta
	}
	outMeta["action"] = reply.ActionTaken
	metaB, _ := json.Marshal(outMeta)

	outMsg, _, err := h.Conversations.InsertMessage(ctx, &conversations.Message{
		ConversationID: conv.ID,
		Direction:      "outbound",
		Sender:         "bot",
		Body:           reply.Text,
		Meta:           metaB,
	})
	if err != nil {
		h.Log.Error().Err(err).Msg("save bot msg")
	} else {
		h.Hub.Publish(clinicID, "message", outMsg)
	}

	// 9) Отправляем в WhatsApp.
	if err := h.WhatsApp.SendText(ctx, m.From, reply.Text); err != nil {
		h.Log.Error().Err(err).Msg("whatsapp send")
	}

	// 10) Если создали запись — паблишим событие.
	if reply.Appointment != nil {
		h.Hub.Publish(clinicID, "appointment", map[string]any{
			"id":         reply.Appointment.AppointmentID,
			"patient_id": patient.ID,
		})
	}
}
