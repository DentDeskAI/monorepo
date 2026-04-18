package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dentdesk/backend/internal/middleware"
	"github.com/dentdesk/backend/internal/repository"
	"github.com/dentdesk/backend/internal/service"
)

// AppointmentHandler handles REST endpoints for appointment management.
type AppointmentHandler struct {
	svc *service.AppointmentService
}

func NewAppointmentHandler(svc *service.AppointmentService) *AppointmentHandler {
	return &AppointmentHandler{svc: svc}
}

// List godoc
// GET /api/v1/appointments?date_from=2025-01-01&date_to=2025-01-31&doctor_id=...
func (h *AppointmentHandler) List(c *gin.Context) {
	clinicID := middleware.ClinicIDFromCtx(c)

	filter := repository.AppointmentFilter{
		Page:     queryInt(c, "page", 1),
		PageSize: queryInt(c, "page_size", 50),
	}

	if v := c.Query("date_from"); v != "" {
		filter.DateFrom = &v
	}
	if v := c.Query("date_to"); v != "" {
		filter.DateTo = &v
	}

	appts, err := h.svc.List(c.Request.Context(), clinicID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": appts})
}

// Get godoc
// GET /api/v1/appointments/:id
func (h *AppointmentHandler) Get(c *gin.Context) {
	clinicID, apptID, ok := clinicAndResourceID(c)
	if !ok {
		return
	}

	appt, err := h.svc.GetByID(c.Request.Context(), clinicID, apptID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "appointment not found"})
		return
	}

	c.JSON(http.StatusOK, appt)
}

// Create godoc
// POST /api/v1/appointments
func (h *AppointmentHandler) Create(c *gin.Context) {
	clinicID := middleware.ClinicIDFromCtx(c)

	var req service.CreateAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	appt, err := h.svc.Create(c.Request.Context(), clinicID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, appt)
}

// Update godoc
// PUT /api/v1/appointments/:id
func (h *AppointmentHandler) Update(c *gin.Context) {
	clinicID, apptID, ok := clinicAndResourceID(c)
	if !ok {
		return
	}

	var req service.UpdateAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	appt, err := h.svc.Update(c.Request.Context(), clinicID, apptID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, appt)
}

// Delete godoc
// DELETE /api/v1/appointments/:id
func (h *AppointmentHandler) Delete(c *gin.Context) {
	clinicID, apptID, ok := clinicAndResourceID(c)
	if !ok {
		return
	}

	if err := h.svc.Delete(c.Request.Context(), clinicID, apptID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// SendManual handles POST /api/v1/messages/send — staff sends a manual WhatsApp message.
func (h *AppointmentHandler) SendManual(_ *gin.Context) {} // implemented in waH
