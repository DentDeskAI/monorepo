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
// GET /api/v1/patients?page=1&page_size=20
func (h *PatientHandler) List(c *gin.Context) {
	clinicID := middleware.ClinicIDFromCtx(c)
	page     := queryInt(c, "page", 1)
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
// GET /api/v1/patients/:id
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
// POST /api/v1/patients
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
// PUT /api/v1/patients/:id
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
// DELETE /api/v1/patients/:id
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
