package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/dentdesk/dentdesk/internal/chairs"
	"github.com/dentdesk/dentdesk/internal/doctors"
	"github.com/dentdesk/dentdesk/internal/http/middleware"
	"github.com/dentdesk/dentdesk/internal/patients"
)

type ResourceHandler struct {
	Doctors  *doctors.Repo
	Chairs   *chairs.Repo
	Patients *patients.Repo
}

// ── Doctors ───────────────────────────────────────────────────────────────────

type createDoctorReq struct {
	Name       string  `json:"name"        binding:"required" example:"Айгерим Касымова"`
	Specialty  *string `json:"specialty"   example:"therapist"`
	ExternalID *string `json:"external_id" example:"DOC-001"`
}

// CreateDoctor godoc
// @Summary     Create doctor
// @Description Adds a new doctor to the clinic. Requires owner or admin.
// @Tags        doctors
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body createDoctorReq true "Doctor data"
// @Success     201 {object} doctors.Doctor
// @Failure     400 {object} map[string]string
// @Failure     403 {object} map[string]string
// @Router      /api/doctors [post]
func (h *ResourceHandler) CreateDoctor(c *gin.Context) {
	if !isOwnerOrAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	cl := middleware.ClaimsFrom(c)
	var req createDoctorReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	d, err := h.Doctors.Create(c.Request.Context(), cl.ClinicID, req.Name, req.Specialty, req.ExternalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, d)
}

// GetDoctor godoc
// @Summary     Get doctor
// @Description Returns a single doctor by ID.
// @Tags        doctors
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "Doctor UUID"
// @Success     200 {object} doctors.Doctor
// @Failure     404 {object} map[string]string
// @Router      /api/doctors/{id} [get]
func (h *ResourceHandler) GetDoctor(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	d, err := h.Doctors.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, d)
}

type updateDoctorReq struct {
	Name      string  `json:"name"      binding:"required"`
	Specialty *string `json:"specialty"`
	Active    bool    `json:"active"`
}

// UpdateDoctor godoc
// @Summary     Update doctor
// @Description Updates a doctor's name, specialty, and active status. Requires owner or admin.
// @Tags        doctors
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id   path string          true "Doctor UUID"
// @Param       body body updateDoctorReq true "Doctor fields"
// @Success     200 {object} map[string]string
// @Failure     400 {object} map[string]string
// @Failure     403 {object} map[string]string
// @Router      /api/doctors/{id} [put]
func (h *ResourceHandler) UpdateDoctor(c *gin.Context) {
	if !isOwnerOrAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	var req updateDoctorReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	if err := h.Doctors.Update(c.Request.Context(), id, req.Name, req.Specialty, req.Active); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

// DeactivateDoctor godoc
// @Summary     Deactivate doctor
// @Description Soft-deletes a doctor (sets active=false). Requires owner or admin.
// @Tags        doctors
// @Security    BearerAuth
// @Param       id path string true "Doctor UUID"
// @Success     204
// @Failure     403 {object} map[string]string
// @Router      /api/doctors/{id} [delete]
func (h *ResourceHandler) DeactivateDoctor(c *gin.Context) {
	if !isOwnerOrAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	if err := h.Doctors.Deactivate(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ── Chairs ────────────────────────────────────────────────────────────────────

// ListChairs godoc
// @Summary     List chairs
// @Description Returns all chairs for the clinic.
// @Tags        chairs
// @Produce     json
// @Security    BearerAuth
// @Success     200 {array}  chairs.Chair
// @Router      /api/chairs [get]
func (h *ResourceHandler) ListChairs(c *gin.Context) {
	cl := middleware.ClaimsFrom(c)
	out, err := h.Chairs.List(c.Request.Context(), cl.ClinicID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

type createChairReq struct {
	Name       string  `json:"name"        binding:"required" example:"Кабинет 1"`
	ExternalID *string `json:"external_id" example:"CHAIR-01"`
}

// CreateChair godoc
// @Summary     Create chair
// @Description Adds a new chair/cabinet to the clinic. Requires owner or admin.
// @Tags        chairs
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body createChairReq true "Chair data"
// @Success     201 {object} chairs.Chair
// @Failure     400 {object} map[string]string
// @Failure     403 {object} map[string]string
// @Router      /api/chairs [post]
func (h *ResourceHandler) CreateChair(c *gin.Context) {
	if !isOwnerOrAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	cl := middleware.ClaimsFrom(c)
	var req createChairReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	ch, err := h.Chairs.Create(c.Request.Context(), cl.ClinicID, req.Name, req.ExternalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, ch)
}

type updateChairReq struct {
	Name string `json:"name" binding:"required"`
}

// UpdateChair godoc
// @Summary     Update chair
// @Description Updates a chair's name. Requires owner or admin.
// @Tags        chairs
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id   path string        true "Chair UUID"
// @Param       body body updateChairReq true "Chair name"
// @Success     200 {object} map[string]string
// @Failure     403 {object} map[string]string
// @Router      /api/chairs/{id} [put]
func (h *ResourceHandler) UpdateChair(c *gin.Context) {
	if !isOwnerOrAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	var req updateChairReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	if err := h.Chairs.Update(c.Request.Context(), id, req.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

// DeactivateChair godoc
// @Summary     Deactivate chair
// @Description Soft-deletes a chair. Requires owner or admin.
// @Tags        chairs
// @Security    BearerAuth
// @Param       id path string true "Chair UUID"
// @Success     204
// @Failure     403 {object} map[string]string
// @Router      /api/chairs/{id} [delete]
func (h *ResourceHandler) DeactivateChair(c *gin.Context) {
	if !isOwnerOrAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	if err := h.Chairs.Deactivate(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ── Patients ──────────────────────────────────────────────────────────────────

type createPatientReq struct {
	Phone      string  `json:"phone"       binding:"required" example:"+77001234567"`
	Name       *string `json:"name"        example:"Асем Нурова"`
	Language   string  `json:"language"    example:"ru"`
	ExternalID *string `json:"external_id" example:"PAT-001"`
}

// CreatePatient godoc
// @Summary     Create patient
// @Description Manually creates a patient record for the clinic.
// @Tags        patients
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body createPatientReq true "Patient data"
// @Success     201 {object} patients.Patient
// @Failure     400 {object} map[string]string
// @Failure     409 {object} map[string]string
// @Router      /api/patients [post]
func (h *ResourceHandler) CreatePatient(c *gin.Context) {
	cl := middleware.ClaimsFrom(c)
	var req createPatientReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	lang := req.Language
	if lang == "" {
		lang = "ru"
	}
	p, err := h.Patients.Create(c.Request.Context(), cl.ClinicID, req.Phone, lang, req.Name, req.ExternalID)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "phone already exists for this clinic"})
		return
	}
	c.JSON(http.StatusCreated, p)
}

// GetPatient godoc
// @Summary     Get patient
// @Description Returns a single patient by ID.
// @Tags        patients
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "Patient UUID"
// @Success     200 {object} patients.Patient
// @Failure     404 {object} map[string]string
// @Router      /api/patients/{id} [get]
func (h *ResourceHandler) GetPatient(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	p, err := h.Patients.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, p)
}

type updatePatientReq struct {
	Name       *string `json:"name"`
	Language   string  `json:"language" example:"ru"`
	ExternalID *string `json:"external_id"`
}

// UpdatePatient godoc
// @Summary     Update patient
// @Description Updates a patient's name, language, and external ID.
// @Tags        patients
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id   path string           true "Patient UUID"
// @Param       body body updatePatientReq true "Patient fields"
// @Success     200 {object} map[string]string
// @Failure     400 {object} map[string]string
// @Router      /api/patients/{id} [put]
func (h *ResourceHandler) UpdatePatient(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	var req updatePatientReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	lang := req.Language
	if lang == "" {
		lang = "ru"
	}
	if err := h.Patients.Update(c.Request.Context(), id, req.Name, lang, req.ExternalID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}
