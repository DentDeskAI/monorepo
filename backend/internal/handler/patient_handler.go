package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/dentdesk/backend/internal/middleware"
	"github.com/dentdesk/backend/internal/service"
)

// PatientHandler handles REST endpoints for patient management.
type PatientHandler struct {
	svc *service.PatientService
}

func NewPatientHandler(svc *service.PatientService) *PatientHandler {
	return &PatientHandler{svc: svc}
}

// List godoc
// @Summary List patients
// @Description Returns paginated patients for the authenticated clinic.
// @Tags patients
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} PatientListResponse
// @Failure 500 {object} ErrorResponse
// @Router /patients [get]
func (h *PatientHandler) List(c *gin.Context) {
	clinicID := middleware.ClinicIDFromCtx(c)
	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "page_size", 20)

	patients, total, err := h.svc.List(c.Request.Context(), clinicID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      patients,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// Get godoc
// @Summary Get patient
// @Description Returns a patient by ID for the authenticated clinic.
// @Tags patients
// @Produce json
// @Security BearerAuth
// @Param id path string true "Patient ID (UUID)"
// @Success 200 {object} PatientDTO
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /patients/{id} [get]
func (h *PatientHandler) Get(c *gin.Context) {
	clinicID, patientID, ok := clinicAndResourceID(c)
	if !ok {
		return
	}

	patient, err := h.svc.GetByID(c.Request.Context(), clinicID, patientID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "patient not found"})
		return
	}

	c.JSON(http.StatusOK, patient)
}

// Create godoc
// @Summary Create patient
// @Description Creates a new patient in the authenticated clinic.
// @Tags patients
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.CreatePatientRequest true "Patient payload"
// @Success 201 {object} PatientDTO
// @Failure 400 {object} ErrorResponse
// @Router /patients [post]
func (h *PatientHandler) Create(c *gin.Context) {
	clinicID := middleware.ClinicIDFromCtx(c)

	var req service.CreatePatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	patient, err := h.svc.Create(c.Request.Context(), clinicID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, patient)
}

// Update godoc
// @Summary Update patient
// @Description Updates an existing patient by ID.
// @Tags patients
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Patient ID (UUID)"
// @Param request body service.UpdatePatientRequest true "Patient update payload"
// @Success 200 {object} PatientDTO
// @Failure 400 {object} ErrorResponse
// @Router /patients/{id} [put]
func (h *PatientHandler) Update(c *gin.Context) {
	clinicID, patientID, ok := clinicAndResourceID(c)
	if !ok {
		return
	}

	var req service.UpdatePatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	patient, err := h.svc.Update(c.Request.Context(), clinicID, patientID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, patient)
}

// Delete godoc
// @Summary Delete patient
// @Description Deletes a patient by ID.
// @Tags patients
// @Produce json
// @Security BearerAuth
// @Param id path string true "Patient ID (UUID)"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /patients/{id} [delete]
func (h *PatientHandler) Delete(c *gin.Context) {
	clinicID, patientID, ok := clinicAndResourceID(c)
	if !ok {
		return
	}

	if err := h.svc.Delete(c.Request.Context(), clinicID, patientID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ─── Shared helpers ───────────────────────────────────────────────────────────

func clinicAndResourceID(c *gin.Context) (clinicID, resourceID uuid.UUID, ok bool) {
	clinicID = middleware.ClinicIDFromCtx(c)

	resourceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return uuid.Nil, uuid.Nil, false
	}

	return clinicID, resourceID, true
}

func queryInt(c *gin.Context, key string, def int) int {
	if v := c.Query(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return def
}
