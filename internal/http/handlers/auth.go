package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dentdesk/dentdesk/internal/auth"
	"github.com/dentdesk/dentdesk/internal/http/middleware"
	errs "github.com/dentdesk/dentdesk/internal/platform/errors"
)

type AuthHandler struct {
	Svc *auth.Service
}

type loginReq struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

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
