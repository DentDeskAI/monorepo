package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/dentdesk/dentdesk/internal/http/middleware"
	"github.com/dentdesk/dentdesk/internal/patients"
	"github.com/dentdesk/dentdesk/internal/services"
	"github.com/dentdesk/dentdesk/internal/store"
)

type CRMHandler struct {
	Svc      *services.CRMService
	Patients *store.PatientRepo
}

func (h *CRMHandler) ListChats(c *gin.Context) {
	cl := middleware.ClaimsFrom(c)
	out, err := h.Svc.ListChats(c.Request.Context(), cl.ClinicID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *CRMHandler) ListMessages(c *gin.Context) {
	convID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	msgs, err := h.Svc.ListMessages(c.Request.Context(), convID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, msgs)
}

type sendMsgReq struct {
	Body string `json:"body" binding:"required" example:"Завтра в 10:00 вам подходит?"`
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
	msg, err := h.Svc.OperatorSend(c.Request.Context(), cl.ClinicID, cl.UserID, convID, req.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, msg)
}

func (h *CRMHandler) ReleaseHandoff(c *gin.Context) {
	convID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	if err := h.Svc.ReleaseHandoff(c.Request.Context(), convID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *CRMHandler) ListPatients(c *gin.Context) {
	cl := middleware.ClaimsFrom(c)
	out, err := h.Svc.ListPatients(c.Request.Context(), cl.ClinicID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *CRMHandler) PatientAppointments(c *gin.Context) {
	cl := middleware.ClaimsFrom(c)
	idStr := c.Param("id")

	// Try UUID first (internal callers), then integer seq_id (local/mock clinics).
	pid, err := uuid.Parse(idStr)
	if err != nil {
		seqID, atoiErr := strconv.Atoi(idStr)
		if atoiErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
			return
		}
		p, lookupErr := h.Patients.GetBySeqID(c.Request.Context(), cl.ClinicID, seqID)
		if lookupErr != nil {
			c.JSON(http.StatusOK, []interface{}{})
			return
		}
		pid = p.ID
	}

	out, err := h.Svc.PatientAppointments(c.Request.Context(), pid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *CRMHandler) Calendar(c *gin.Context) {
	cl := middleware.ClaimsFrom(c)
	var from *time.Time
	var to *time.Time
	parsedFrom, err1 := time.Parse(time.RFC3339, c.Query("from"))
	parsedTo, err2 := time.Parse(time.RFC3339, c.Query("to"))
	if err1 == nil && err2 == nil {
		from = &parsedFrom
		to = &parsedTo
	}
	out, err := h.Svc.Calendar(c.Request.Context(), cl.ClinicID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *CRMHandler) ListDoctors(c *gin.Context) {
	cl := middleware.ClaimsFrom(c)
	out, err := h.Svc.ListDoctors(c.Request.Context(), cl.ClinicID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *CRMHandler) SSE(c *gin.Context) {
	cl := middleware.ClaimsFrom(c)
	ch, unsub := h.Svc.Hub.Subscribe(cl.ClinicID)
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

func (h *CRMHandler) Stats(c *gin.Context) {
	cl := middleware.ClaimsFrom(c)
	s, _ := h.Svc.Stats(c.Request.Context(), cl.ClinicID)
	c.JSON(http.StatusOK, s)
}
