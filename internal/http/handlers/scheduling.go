package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/dentdesk/dentdesk/internal/appointments"
	"github.com/dentdesk/dentdesk/internal/conversations"
	"github.com/dentdesk/dentdesk/internal/http/middleware"
	"github.com/dentdesk/dentdesk/internal/scheduler"
)

type SchedulingHandler struct {
	Appointments  *appointments.Repo
	Conversations *conversations.Repo
	Scheduler     scheduler.Scheduler
}

// GetSlots godoc
// @Summary     Get free slots
// @Description Returns available appointment slots for the clinic in a time range.
// @Tags        scheduling
// @Produce     json
// @Security    BearerAuth
// @Param       from      query string false "Start time RFC3339" example:"2026-04-25T09:00:00Z"
// @Param       to        query string false "End time RFC3339"   example:"2026-04-26T20:00:00Z"
// @Param       specialty query string false "Filter by specialty" example:"therapist"
// @Success     200 {array}  object
// @Failure     401 {object} map[string]string
// @Router      /api/slots [get]
func (h *SchedulingHandler) GetSlots(c *gin.Context) {
	cl := middleware.ClaimsFrom(c)
	from, err1 := time.Parse(time.RFC3339, c.Query("from"))
	to, err2 := time.Parse(time.RFC3339, c.Query("to"))
	if err1 != nil || err2 != nil {
		from = time.Now()
		to = time.Now().Add(7 * 24 * time.Hour)
	}
	slots, err := h.Scheduler.GetFreeSlots(c.Request.Context(), cl.ClinicID, from, to, c.Query("specialty"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if slots == nil {
		slots = []scheduler.Slot{}
	}
	c.JSON(http.StatusOK, slots)
}

type createAppointmentReq struct {
	PatientID uuid.UUID  `json:"patient_id" binding:"required"`
	DoctorID  *uuid.UUID `json:"doctor_id"`
	ChairID   *uuid.UUID `json:"chair_id"`
	StartsAt  time.Time  `json:"starts_at"  binding:"required"`
	EndsAt    time.Time  `json:"ends_at"    binding:"required"`
	Service   *string    `json:"service"    example:"Чистка зубов"`
}

// CreateAppointment godoc
// @Summary     Create appointment
// @Description Manually books an appointment (source=operator).
// @Tags        scheduling
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body createAppointmentReq true "Appointment data"
// @Success     201 {object} object
// @Failure     400 {object} map[string]string
// @Failure     401 {object} map[string]string
// @Router      /api/appointments [post]
func (h *SchedulingHandler) CreateAppointment(c *gin.Context) {
	cl := middleware.ClaimsFrom(c)
	var req createAppointmentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	if !req.EndsAt.After(req.StartsAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ends_at must be after starts_at"})
		return
	}
	a := &appointments.Appointment{
		ClinicID:  cl.ClinicID,
		PatientID: req.PatientID,
		DoctorID:  req.DoctorID,
		ChairID:   req.ChairID,
		StartsAt:  req.StartsAt,
		EndsAt:    req.EndsAt,
		Service:   req.Service,
		Status:    "scheduled",
		Source:    "operator",
	}
	out, err := h.Appointments.Create(c.Request.Context(), a)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, out)
}

// GetAppointment godoc
// @Summary     Get appointment
// @Description Returns a single appointment with patient and doctor info.
// @Tags        scheduling
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "Appointment UUID"
// @Success     200 {object} object
// @Failure     404 {object} map[string]string
// @Router      /api/appointments/{id} [get]
func (h *SchedulingHandler) GetAppointment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	a, err := h.Appointments.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, a)
}

type updateStatusReq struct {
	Status string `json:"status" binding:"required" example:"confirmed"`
}

// UpdateAppointmentStatus godoc
// @Summary     Update appointment status
// @Description Updates appointment status. Allowed values: scheduled, confirmed, cancelled, completed, no_show.
// @Tags        scheduling
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id   path string          true "Appointment UUID"
// @Param       body body updateStatusReq true "New status"
// @Success     200 {object} map[string]string
// @Failure     400 {object} map[string]string
// @Router      /api/appointments/{id}/status [put]
func (h *SchedulingHandler) UpdateAppointmentStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	var req updateStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	valid := map[string]bool{
		"scheduled": true, "confirmed": true, "cancelled": true, "completed": true, "no_show": true,
	}
	if !valid[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
		return
	}
	if err := h.Appointments.SetStatus(c.Request.Context(), id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": req.Status})
}

// GetConversation godoc
// @Summary     Get conversation
// @Description Returns a single conversation by ID.
// @Tags        chats
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "Conversation UUID"
// @Success     200 {object} object
// @Failure     404 {object} map[string]string
// @Router      /api/chats/{id} [get]
func (h *SchedulingHandler) GetConversation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	conv, err := h.Conversations.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, conv)
}

// CloseConversation godoc
// @Summary     Close conversation
// @Description Sets conversation status to closed. Bot will no longer respond.
// @Tags        chats
// @Security    BearerAuth
// @Param       id path string true "Conversation UUID"
// @Success     204
// @Failure     400 {object} map[string]string
// @Router      /api/chats/{id}/close [post]
func (h *SchedulingHandler) CloseConversation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	if err := h.Conversations.SetStatus(c.Request.Context(), id, "closed"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
