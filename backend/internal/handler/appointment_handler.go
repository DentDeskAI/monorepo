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
// @Summary List appointments
// @Description Returns appointments for the authenticated clinic with optional date range filter.
// @Tags appointments
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(50)
// @Param date_from query string false "Start date (YYYY-MM-DD)"
// @Param date_to query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} AppointmentListResponse
// @Failure 500 {object} ErrorResponse
// @Router /appointments [get]
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
// @Summary Get appointment
// @Description Returns an appointment by ID for the authenticated clinic.
// @Tags appointments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Appointment ID (UUID)"
// @Success 200 {object} AppointmentDTO
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /appointments/{id} [get]
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
// @Summary Create appointment
// @Description Creates a new appointment in the authenticated clinic.
// @Tags appointments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.CreateAppointmentRequest true "Appointment payload"
// @Success 201 {object} AppointmentDTO
// @Failure 400 {object} ErrorResponse
// @Router /appointments [post]
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
// @Summary Update appointment
// @Description Updates an existing appointment by ID.
// @Tags appointments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Appointment ID (UUID)"
// @Param request body service.UpdateAppointmentRequest true "Appointment update payload"
// @Success 200 {object} AppointmentDTO
// @Failure 400 {object} ErrorResponse
// @Router /appointments/{id} [put]
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
// @Summary Delete appointment
// @Description Deletes an appointment by ID.
// @Tags appointments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Appointment ID (UUID)"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /appointments/{id} [delete]
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
