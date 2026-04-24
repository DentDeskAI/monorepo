package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/dentdesk/dentdesk/internal/appointments"
	"github.com/dentdesk/dentdesk/internal/conversations"
	"github.com/dentdesk/dentdesk/internal/doctors"
	"github.com/dentdesk/dentdesk/internal/http/middleware"
	"github.com/dentdesk/dentdesk/internal/patients"
	"github.com/dentdesk/dentdesk/internal/realtime"
	"github.com/dentdesk/dentdesk/internal/whatsapp"
)

type CRMHandler struct {
	DB            *sqlx.DB
	Patients      *patients.Repo
	Conversations *conversations.Repo
	Appointments  *appointments.Repo
	Doctors       *doctors.Repo
	Hub           *realtime.Hub
	WhatsApp      *whatsapp.Client
}

// --- Chats ---

func (h *CRMHandler) ListChats(c *gin.Context) {
	cl := middleware.ClaimsFrom(c)
	convs, err := h.Conversations.ListForClinic(c.Request.Context(), cl.ClinicID, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Обогащаем пациентом + последним сообщением (N+1 вариант, норм для MVP).
	type row struct {
		Conversation conversations.Conversation `json:"conversation"`
		Patient      *patients.Patient          `json:"patient"`
		LastMessage  *conversations.Message     `json:"last_message,omitempty"`
	}
	out := make([]row, 0, len(convs))
	for _, conv := range convs {
		p, _ := h.Patients.Get(c.Request.Context(), conv.PatientID)
		msgs, _ := h.Conversations.ListMessages(c.Request.Context(), conv.ID, 1)
		var last *conversations.Message
		if len(msgs) > 0 {
			last = &msgs[0]
		}
		out = append(out, row{Conversation: conv, Patient: p, LastMessage: last})
	}
	c.JSON(http.StatusOK, out)
}

func (h *CRMHandler) ListMessages(c *gin.Context) {
	convID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	msgs, err := h.Conversations.RecentHistory(c.Request.Context(), convID, 200)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, msgs)
}

type sendMsgReq struct {
	Body string `json:"body" binding:"required"`
}

func (h *CRMHandler) OperatorSend(c *gin.Context) {
	cl := middleware.ClaimsFrom(c)
	convID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	var req sendMsgReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad body"})
		return
	}
	// Берём conversation + patient, чтобы знать номер.
	convs, _ := h.Conversations.ListForClinic(c.Request.Context(), cl.ClinicID, 500)
	var conv *conversations.Conversation
	for i := range convs {
		if convs[i].ID == convID {
			conv = &convs[i]
			break
		}
	}
	if conv == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	p, err := h.Patients.Get(c.Request.Context(), conv.PatientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	meta, _ := json.Marshal(map[string]any{"operator_id": cl.UserID})
	msg, _, err := h.Conversations.InsertMessage(c.Request.Context(), &conversations.Message{
		ConversationID: convID,
		Direction:      "outbound",
		Sender:         "operator",
		Body:           req.Body,
		Meta:           meta,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Переводим диалог в handoff (бот замолкает).
	_ = h.Conversations.SetStatus(c.Request.Context(), convID, "handoff")

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = h.WhatsApp.SendText(ctx, p.Phone, req.Body)
	}()

	h.Hub.Publish(cl.ClinicID, "message", msg)
	c.JSON(http.StatusOK, msg)
}

func (h *CRMHandler) ReleaseHandoff(c *gin.Context) {
	convID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	if err := h.Conversations.SetStatus(c.Request.Context(), convID, "active"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Patients ---

func (h *CRMHandler) ListPatients(c *gin.Context) {
	cl := middleware.ClaimsFrom(c)
	out, err := h.Patients.List(c.Request.Context(), cl.ClinicID, 200)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *CRMHandler) PatientAppointments(c *gin.Context) {
	pid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	out, err := h.Appointments.ListForPatient(c.Request.Context(), pid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

// --- Calendar ---

func (h *CRMHandler) Calendar(c *gin.Context) {
	cl := middleware.ClaimsFrom(c)
	fromStr := c.Query("from")
	toStr := c.Query("to")
	from, err1 := time.Parse(time.RFC3339, fromStr)
	to, err2 := time.Parse(time.RFC3339, toStr)
	if err1 != nil || err2 != nil {
		// дефолт: ближайшие 7 дней
		from = time.Now().Add(-24 * time.Hour)
		to = time.Now().Add(7 * 24 * time.Hour)
	}
	out, err := h.Appointments.ListForPeriod(c.Request.Context(), cl.ClinicID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

// --- Doctors ---

func (h *CRMHandler) ListDoctors(c *gin.Context) {
	cl := middleware.ClaimsFrom(c)
	out, err := h.Doctors.List(c.Request.Context(), cl.ClinicID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

// --- SSE ---

func (h *CRMHandler) SSE(c *gin.Context) {
	// Gin SSE не поддерживает "текущий EventSource" из браузера с JWT в хедере,
	// поэтому токен приходит в query ?token=...
	tokenStr := c.Query("token")
	_ = tokenStr // Токен валидируется в wrapper-мидлваре

	cl := middleware.ClaimsFrom(c)
	ch, unsub := h.Hub.Subscribe(cl.ClinicID)
	defer unsub()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	c.Writer.Flush()

	ctx := c.Request.Context()
	ping := time.NewTicker(20 * time.Second)
	defer ping.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ping.C:
			_, _ = c.Writer.Write([]byte(": ping\n\n"))
			c.Writer.Flush()
		case ev, ok := <-ch:
			if !ok {
				return
			}
			b, _ := json.Marshal(ev)
			_, _ = c.Writer.Write([]byte("event: " + ev.Type + "\ndata: "))
			_, _ = c.Writer.Write(b)
			_, _ = c.Writer.Write([]byte("\n\n"))
			c.Writer.Flush()
		}
	}
}

// --- Stats (tiny dashboard widget) ---

func (h *CRMHandler) Stats(c *gin.Context) {
	cl := middleware.ClaimsFrom(c)
	ctx := c.Request.Context()
	type stats struct {
		ActiveChats    int `json:"active_chats"`
		TodayAppts     int `json:"today_appts"`
		TotalPatients  int `json:"total_patients"`
	}
	var s stats
	db := h.DB
	_ = db.GetContext(ctx, &s.ActiveChats,
		`SELECT COUNT(*) FROM conversations WHERE clinic_id=$1 AND status='active'
		   AND last_message_at > NOW() - INTERVAL '24 hours'`, cl.ClinicID)
	_ = db.GetContext(ctx, &s.TodayAppts,
		`SELECT COUNT(*) FROM appointments WHERE clinic_id=$1
		   AND starts_at::date = CURRENT_DATE AND status IN ('scheduled','confirmed')`, cl.ClinicID)
	_ = db.GetContext(ctx, &s.TotalPatients,
		`SELECT COUNT(*) FROM patients WHERE clinic_id=$1`, cl.ClinicID)
	c.JSON(http.StatusOK, s)
}
