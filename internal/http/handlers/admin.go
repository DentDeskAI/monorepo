package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/dentdesk/dentdesk/internal/http/middleware"
	"github.com/dentdesk/dentdesk/internal/services"
)

type AdminHandler struct {
	Svc *services.AdminService
}

func isOwnerOrAdmin(c *gin.Context) bool {
	cl := middleware.ClaimsFrom(c)
	return cl != nil && (cl.Role == "owner" || cl.Role == "admin")
}

func isOwner(c *gin.Context) bool {
	cl := middleware.ClaimsFrom(c)
	return cl != nil && cl.Role == "owner"
}

type registerReq struct {
	ClinicName string `json:"clinic_name" binding:"required" example:"Happy Smile Dental"`
	Timezone   string `json:"timezone"    binding:"required" example:"Asia/Almaty"`
	OwnerName  string `json:"owner_name"  binding:"required" example:"Айгерим Касымова"`
	Email      string `json:"email"       binding:"required" example:"owner@clinic.kz"`
	Password   string `json:"password"    binding:"required" example:"secret1234"`
}

func (h *AdminHandler) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	clinic, user, token, err := h.Svc.Register(c.Request.Context(), req.ClinicName, req.Timezone, req.OwnerName, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already taken"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"token":  token,
		"clinic": clinic,
		"user": gin.H{
			"id": user.ID, "email": user.Email, "name": user.Name, "role": user.Role,
		},
	})
}

func (h *AdminHandler) GetClinic(c *gin.Context) {
	cl := middleware.ClaimsFrom(c)
	clinic, err := h.Svc.GetClinic(c.Request.Context(), cl.ClinicID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "clinic not found"})
		return
	}
	c.JSON(http.StatusOK, clinic)
}

type updateClinicReq struct {
	Name            string `json:"name"              binding:"required"`
	Timezone        string `json:"timezone"          binding:"required"`
	WorkingHours    string `json:"working_hours"`
	SlotDurationMin int    `json:"slot_duration_min"`
	SchedulerType   string `json:"scheduler_type"`
}

func (h *AdminHandler) UpdateClinic(c *gin.Context) {
	if !isOwnerOrAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	cl := middleware.ClaimsFrom(c)
	var req updateClinicReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	clinic, err := h.Svc.UpdateClinic(c.Request.Context(), cl.ClinicID, req.Name, req.Timezone, req.WorkingHours, req.SchedulerType, req.SlotDurationMin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, clinic)
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	if !isOwnerOrAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	cl := middleware.ClaimsFrom(c)
	users, err := h.Svc.ListUsers(c.Request.Context(), cl.ClinicID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	type safe struct {
		ID       uuid.UUID `json:"id"`
		Email    string    `json:"email"`
		Name     string    `json:"name"`
		Role     string    `json:"role"`
		ClinicID uuid.UUID `json:"clinic_id"`
	}
	out := make([]safe, len(users))
	for i, u := range users {
		out[i] = safe{ID: u.ID, Email: u.Email, Name: u.Name, Role: u.Role, ClinicID: u.ClinicID}
	}
	c.JSON(http.StatusOK, out)
}

type createUserReq struct {
	Email    string `json:"email"    binding:"required" example:"operator@clinic.kz"`
	Password string `json:"password" binding:"required" example:"pass1234"`
	Name     string `json:"name"     binding:"required" example:"Бауыржан Ахметов"`
	Role     string `json:"role"     binding:"required" example:"operator"`
}

func (h *AdminHandler) CreateUser(c *gin.Context) {
	if !isOwnerOrAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	cl := middleware.ClaimsFrom(c)
	var req createUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	validRoles := map[string]bool{"owner": true, "admin": true, "operator": true}
	if !validRoles[req.Role] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role must be owner, admin, or operator"})
		return
	}
	u, err := h.Svc.CreateUser(c.Request.Context(), cl.ClinicID, req.Email, req.Password, req.Role, req.Name)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already taken"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": u.ID, "email": u.Email, "name": u.Name, "role": u.Role})
}

func (h *AdminHandler) GetUser(c *gin.Context) {
	if !isOwnerOrAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	u, err := h.Svc.GetUser(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": u.ID, "email": u.Email, "name": u.Name, "role": u.Role, "clinic_id": u.ClinicID})
}

type updateUserReq struct {
	Name string `json:"name" binding:"required"`
	Role string `json:"role" binding:"required"`
}

func (h *AdminHandler) UpdateUser(c *gin.Context) {
	if !isOwnerOrAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	var req updateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	if err := h.Svc.UpdateUser(c.Request.Context(), id, req.Name, req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	if !isOwner(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: owner only"})
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	if err := h.Svc.DeleteUser(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

type changePasswordReq struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func (h *AdminHandler) ChangePassword(c *gin.Context) {
	cl := middleware.ClaimsFrom(c)
	var req changePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	if err := h.Svc.ChangePassword(c.Request.Context(), cl.UserID, req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "wrong current password"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "password changed"})
}
