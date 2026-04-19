package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dentdesk/backend/internal/service"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register godoc
// @Summary Register clinic admin
// @Description Creates a clinic and its first admin user, then returns a JWT.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body service.RegisterRequest true "Registration payload"
// @Success 201 {object} service.AuthResponse
// @Failure 400 {object} ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// Login godoc
// @Summary Login user
// @Description Authenticates user credentials and returns a JWT.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body service.LoginRequest true "Login payload"
// @Success 200 {object} service.AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Me godoc
// @Summary Current user profile
// @Description Returns profile data of the authenticated user.
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} service.UserDTO
// @Failure 500 {object} ErrorResponse
// @Router /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	resp, err := h.authService.Me(c.Request.Context(), c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
