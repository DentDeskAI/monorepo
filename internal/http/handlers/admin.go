package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/dentdesk/dentdesk/internal/auth"
	"github.com/dentdesk/dentdesk/internal/clinics"
	"github.com/dentdesk/dentdesk/internal/http/middleware"
)

type AdminHandler struct {
	Auth    *auth.Service
	Clinics *clinics.Repo
}

// ── helpers ──────────────────────────────────────────────────────────────────

func isOwnerOrAdmin(c *gin.Context) bool {
	cl := middleware.ClaimsFrom(c)
	return cl != nil && (cl.Role == "owner" || cl.Role == "admin")
}

func isOwner(c *gin.Context) bool {
	cl := middleware.ClaimsFrom(c)
	return cl != nil && cl.Role == "owner"
}

// ── Register (public) ────────────────────────────────────────────────────────

type registerReq struct {
	ClinicName string `json:"clinic_name" binding:"required" example:"Happy Smile Dental"`
	Timezone   string `json:"timezone"    binding:"required" example:"Asia/Almaty"`
	OwnerName  string `json:"owner_name"  binding:"required" example:"Айгерим Касымова"`
	Email      string `json:"email"       binding:"required" example:"owner@clinic.kz"`
	Password   string `json:"password"    binding:"required" example:"secret1234"`
}

// Register godoc
// @Summary     Register clinic
// @Description Creates a new clinic and its first owner user. Returns a JWT token.
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body registerReq true "Registration data"
// @Success     201 {object} map[string]interface{}
// @Failure     400 {object} map[string]string
// @Failure     409 {object} map[string]string
// @Router      /api/register [post]
func (h *AdminHandler) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	clinic, err := h.Clinics.Create(c.Request.Context(), req.ClinicName, req.Timezone, "local")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	user, err := h.Auth.CreateUser(c.Request.Context(), clinic.ID, req.Email, req.Password, "owner", req.OwnerName)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already taken"})
		return
	}
	token, _, err := h.Auth.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

// ── Clinic ───────────────────────────────────────────────────────────────────

// GetClinic godoc
// @Summary     Get clinic
// @Description Returns the current clinic's settings.
// @Tags        clinic
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} clinics.Clinic
// @Failure     401 {object} map[string]string
// @Router      /api/clinic [get]
func (h *AdminHandler) GetClinic(c *gin.Context) {
	cl := middleware.ClaimsFrom(c)
	clinic, err := h.Clinics.Get(c.Request.Context(), cl.ClinicID)
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

// UpdateClinic godoc
// @Summary     Update clinic
// @Description Updates clinic settings. Requires owner or admin role.
// @Tags        clinic
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body updateClinicReq true "Clinic fields"
// @Success     200 {object} clinics.Clinic
// @Failure     400 {object} map[string]string
// @Failure     403 {object} map[string]string
// @Router      /api/clinic [put]
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
	dur := req.SlotDurationMin
	if dur == 0 {
		dur = 30
	}
	sched := req.SchedulerType
	if sched == "" {
		sched = "local"
	}
	wh := req.WorkingHours
	if wh == "" {
		wh = `{"mon":["09:00","20:00"],"tue":["09:00","20:00"],"wed":["09:00","20:00"],"thu":["09:00","20:00"],"fri":["09:00","20:00"],"sat":["10:00","18:00"],"sun":null}`
	}
	if err := h.Clinics.Update(c.Request.Context(), cl.ClinicID, clinics.UpdateFields{
		Name: req.Name, Timezone: req.Timezone,
		WorkingHours: []byte(wh), SlotDurationMin: dur, SchedulerType: sched,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	clinic, _ := h.Clinics.Get(c.Request.Context(), cl.ClinicID)
	c.JSON(http.StatusOK, clinic)
}

// ── Users ────────────────────────────────────────────────────────────────────

// ListUsers godoc
// @Summary     List users
// @Description Lists all users (operators/admins/owners) in the clinic. Requires owner or admin.
// @Tags        users
// @Produce     json
// @Security    BearerAuth
// @Success     200 {array}  object
// @Failure     403 {object} map[string]string
// @Router      /api/users [get]
func (h *AdminHandler) ListUsers(c *gin.Context) {
	if !isOwnerOrAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	cl := middleware.ClaimsFrom(c)
	users, err := h.Auth.ListUsers(c.Request.Context(), cl.ClinicID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// strip password hashes
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

// CreateUser godoc
// @Summary     Create user
// @Description Adds a new user to the clinic. Requires owner or admin.
// @Tags        users
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body createUserReq true "User data"
// @Success     201 {object} object
// @Failure     400 {object} map[string]string
// @Failure     403 {object} map[string]string
// @Failure     409 {object} map[string]string
// @Router      /api/users [post]
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
	u, err := h.Auth.CreateUser(c.Request.Context(), cl.ClinicID, req.Email, req.Password, req.Role, req.Name)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already taken"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": u.ID, "email": u.Email, "name": u.Name, "role": u.Role})
}

// GetUser godoc
// @Summary     Get user
// @Description Returns a single user by ID. Requires owner or admin.
// @Tags        users
// @Produce     json
// @Security    BearerAuth
// @Param       id path string true "User UUID"
// @Success     200 {object} object
// @Failure     403 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Router      /api/users/{id} [get]
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
	u, err := h.Auth.GetUser(c.Request.Context(), id)
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

// UpdateUser godoc
// @Summary     Update user
// @Description Updates a user's name and role. Requires owner or admin.
// @Tags        users
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id   path string        true "User UUID"
// @Param       body body updateUserReq true "Updated fields"
// @Success     200 {object} map[string]string
// @Failure     400 {object} map[string]string
// @Failure     403 {object} map[string]string
// @Router      /api/users/{id} [put]
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
	if err := h.Auth.UpdateUser(c.Request.Context(), id, req.Name, req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

// DeleteUser godoc
// @Summary     Delete user
// @Description Removes a user from the clinic. Requires owner role only.
// @Tags        users
// @Security    BearerAuth
// @Param       id path string true "User UUID"
// @Success     204
// @Failure     403 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Router      /api/users/{id} [delete]
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
	if err := h.Auth.DeleteUser(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

type changePasswordReq struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// ChangePassword godoc
// @Summary     Change password
// @Description Changes the authenticated user's own password.
// @Tags        auth
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body changePasswordReq true "Old and new password"
// @Success     200 {object} map[string]string
// @Failure     400 {object} map[string]string
// @Failure     401 {object} map[string]string
// @Router      /api/auth/change-password [post]
func (h *AdminHandler) ChangePassword(c *gin.Context) {
	cl := middleware.ClaimsFrom(c)
	var req changePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	if err := h.Auth.ChangePassword(c.Request.Context(), cl.UserID, req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "wrong current password"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "password changed"})
}
