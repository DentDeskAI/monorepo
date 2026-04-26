package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/dentdesk/dentdesk/internal/http/middleware"
	"github.com/dentdesk/dentdesk/internal/scheduler"
	"github.com/dentdesk/dentdesk/internal/services"
)

// SchedulingHandler exposes scheduling endpoints. Pure read endpoints call the
// scheduler.Service directly (no business logic to add). Endpoints with
// validation or DB writes go through the SchedulingService.
type SchedulingHandler struct {
	Sched *scheduler.Service
	Svc   *services.SchedulingService
}

func (h *SchedulingHandler) SyncDoctors(c *gin.Context) {
	cl := middleware.ClaimsFrom(c)
	n, err := h.Svc.SyncDoctors(c.Request.Context(), cl.ClinicID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"synced": n})
}

func (h *SchedulingHandler) GetDoctor(c *gin.Context) {
	cl := middleware.ClaimsFrom(c)
	d, err := h.Sched.GetDoctor(c.Request.Context(), cl.ClinicID, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, d)
}

func (h *SchedulingHandler) GetDoctors(c *gin.Context) {
	cl := middleware.ClaimsFrom(c)
	docs, err := h.Sched.ListDoctors(c.Request.Context(), cl.ClinicID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, docs)
}

func (h *SchedulingHandler) GetPatient(c *gin.Context) {
	cl := middleware.ClaimsFrom(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	p, err := h.Sched.GetPatient(c.Request.Context(), cl.ClinicID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *SchedulingHandler) GetPatients(c *gin.Context) {
	cl := middleware.ClaimsFrom(c)
	docs, err := h.Sched.ListPatients(c.Request.Context(), cl.ClinicID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, docs)
}

func (h *SchedulingHandler) GetClinic(c *gin.Context) {
	cl := middleware.ClaimsFrom(c)
	stom, err := h.Sched.GetClinic(c.Request.Context(), cl.ClinicID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stom)
}

func (h *SchedulingHandler) GetSlots(c *gin.Context) {
	cl := middleware.ClaimsFrom(c)
	var from *time.Time
	var to *time.Time
	parsedFrom, err1 := time.Parse(time.RFC3339, c.Query("from"))
	parsedTo, err2 := time.Parse(time.RFC3339, c.Query("to"))
	if err1 == nil && err2 == nil {
		from = &parsedFrom
		to = &parsedTo
	}
	slots, err := h.Svc.GetSlots(c.Request.Context(), cl.ClinicID, from, to, c.Query("specialty"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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

func (h *SchedulingHandler) CreateAppointment(c *gin.Context) {
	cl := middleware.ClaimsFrom(c)
	var req createAppointmentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	out, err := h.Svc.CreateAppointment(c.Request.Context(), cl.ClinicID, req.PatientID, req.DoctorID, req.ChairID, req.StartsAt, req.EndsAt, req.Service)
	if err != nil {
		if errors.Is(err, services.ErrInvalidRange) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ends_at must be after starts_at"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, out)
}

func (h *SchedulingHandler) GetAppointment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	a, err := h.Svc.GetAppointment(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, a)
}

type updateStatusReq struct {
	Status string `json:"status" binding:"required" example:"confirmed"`
}

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
	if err := h.Svc.UpdateAppointmentStatus(c.Request.Context(), id, req.Status); err != nil {
		if errors.Is(err, services.ErrInvalidStatus) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": req.Status})
}

func (h *SchedulingHandler) GetConversation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	conv, err := h.Svc.GetConversation(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, conv)
}

func (h *SchedulingHandler) CloseConversation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	if err := h.Svc.CloseConversation(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
