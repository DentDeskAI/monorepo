package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dentdesk/dentdesk/internal/auth"
	"github.com/dentdesk/dentdesk/internal/http/middleware"
	errs "github.com/dentdesk/dentdesk/internal/platform/errors"
)

// AuthService defines the interface for auth operations
type AuthService interface {
	Login(ctx context.Context, email, password string) (string, *auth.User, error)
	Parse(token string) (*auth.Claims, error)
}

type AuthHandler struct {
	Svc AuthService
}

type loginReq struct {
	Email    string `json:"email" binding:"required" example:"test@demo.kz"`
	Password string `json:"password" binding:"required" example:"demo1234"`
}

// Login godoc
// @Summary     Login
// @Description Authenticate with email + password, returns a signed JWT.
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body loginReq true "Credentials"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]string
// @Failure     401 {object} map[string]string
// @Router      /api/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	token, user, err := h.Svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, errs.ErrUnauthorized) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":        user.ID,
			"email":     user.Email,
			"name":      user.Name,
			"role":      user.Role,
			"clinic_id": user.ClinicID,
		},
	})
}

// Me godoc
// @Summary     Current user
// @Description Returns the authenticated user's ID, clinic, and role from the JWT.
// @Tags        auth
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} map[string]interface{}
// @Failure     401 {object} map[string]string
// @Router      /api/auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	if claims == nil {
		c.Status(http.StatusUnauthorized)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user_id":   claims.UserID,
		"clinic_id": claims.ClinicID,
		"role":      claims.Role,
	})
}
