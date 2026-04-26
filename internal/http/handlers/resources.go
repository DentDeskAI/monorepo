package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/dentdesk/dentdesk/internal/http/middleware"
	"github.com/dentdesk/dentdesk/internal/services"
)

type ResourceHandler struct {
	Svc *services.ResourceService
}

type createDoctorReq struct {
	Name       string  `json:"name"        binding:"required" example:"Айгерим Касымова"`
	Specialty  *string `json:"specialty"   example:"therapist"`
	ExternalID *string `json:"external_id" example:"DOC-001"`
}

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
	d, err := h.Svc.CreateDoctor(c.Request.Context(), cl.ClinicID, req.Name, req.Specialty, req.ExternalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, d)
}

func (h *ResourceHandler) GetDoctor(c *gin.Context) {
	param := c.Param("id")
	if id, err := uuid.Parse(param); err == nil {
		d, err := h.Svc.GetDoctor(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, d)
		return
	}
	// param is a MacDent integer ID stored in external_id
	cl := middleware.ClaimsFrom(c)
	d, err := h.Svc.GetDoctorByExternalID(c.Request.Context(), cl.ClinicID, param)
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
	if err := h.Svc.UpdateDoctor(c.Request.Context(), id, req.Name, req.Specialty, req.Active); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

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
	if err := h.Svc.DeactivateDoctor(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ResourceHandler) ListChairs(c *gin.Context) {
	cl := middleware.ClaimsFrom(c)
	out, err := h.Svc.ListChairs(c.Request.Context(), cl.ClinicID)
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
	ch, err := h.Svc.CreateChair(c.Request.Context(), cl.ClinicID, req.Name, req.ExternalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, ch)
}

type updateChairReq struct {
	Name string `json:"name" binding:"required"`
}

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
	if err := h.Svc.UpdateChair(c.Request.Context(), id, req.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

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
	if err := h.Svc.DeactivateChair(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

type createPatientReq struct {
	Phone      string  `json:"phone"       binding:"required" example:"+77001234567"`
	Name       *string `json:"name"        example:"Асем Нурова"`
	Language   string  `json:"language"    example:"ru"`
	ExternalID *string `json:"external_id" example:"PAT-001"`
}

func (h *ResourceHandler) CreatePatient(c *gin.Context) {
	cl := middleware.ClaimsFrom(c)
	var req createPatientReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	p, err := h.Svc.CreatePatient(c.Request.Context(), cl.ClinicID, req.Phone, req.Language, req.Name, req.ExternalID)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "phone already exists for this clinic"})
		return
	}
	c.JSON(http.StatusCreated, p)
}

func (h *ResourceHandler) GetPatient(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	p, err := h.Svc.GetPatient(c.Request.Context(), id)
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
	if err := h.Svc.UpdatePatient(c.Request.Context(), id, req.Name, req.Language, req.ExternalID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}
