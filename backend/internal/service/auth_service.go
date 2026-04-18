package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/dentdesk/backend/internal/config"
	"github.com/dentdesk/backend/internal/domain"
	"github.com/dentdesk/backend/internal/middleware"
	"github.com/dentdesk/backend/internal/repository"
)

// AuthService handles user registration, login, and token management.
type AuthService struct {
	userRepo   userRepository
	clinicRepo repository.ClinicRepository
	cfg        config.JWTConfig
}

// userRepository is a minimal interface needed by AuthService.
type userRepository interface {
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
}

func NewAuthService(userRepo userRepository, clinicRepo repository.ClinicRepository, cfg config.JWTConfig) *AuthService {
	return &AuthService{userRepo: userRepo, clinicRepo: clinicRepo, cfg: cfg}
}

// ─── Request / Response DTOs ──────────────────────────────────────────────────

type RegisterRequest struct {
	ClinicName string `json:"clinic_name" binding:"required,min=2"`
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=8"`
	FirstName  string `json:"first_name" binding:"required"`
	LastName   string `json:"last_name"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      UserDTO   `json:"user"`
}

type UserDTO struct {
	ID        uuid.UUID        `json:"id"`
	ClinicID  uuid.UUID        `json:"clinic_id"`
	Email     string           `json:"email"`
	FirstName string           `json:"first_name"`
	LastName  string           `json:"last_name"`
	Role      domain.UserRole  `json:"role"`
}

// ─── Methods ──────────────────────────────────────────────────────────────────

// Register creates a new clinic + admin user and returns a signed JWT.
func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	// Create clinic
	clinic := &domain.Clinic{
		Name:  req.ClinicName,
		Email: req.Email,
		Slug:  slugify(req.ClinicName),
	}
	if err := s.clinicRepo.Create(ctx, clinic); err != nil {
		return nil, fmt.Errorf("create clinic: %w", err)
	}

	// Create admin user
	user := &domain.User{
		ClinicID:     clinic.ID,
		Email:        req.Email,
		PasswordHash: string(hash),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Role:         domain.RoleAdmin,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return s.buildAuthResponse(user)
}

// Login verifies credentials and returns a signed JWT.
func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	if !user.IsActive {
		return nil, errors.New("account is disabled")
	}

	return s.buildAuthResponse(user)
}

// Me returns the currently authenticated user's profile.
func (s *AuthService) Me(ctx context.Context, c *gin.Context) (*UserDTO, error) {
	userID, _ := c.Get(middleware.CtxUserID)
	clinicID := middleware.ClinicIDFromCtx(c)

	_ = userID
	_ = clinicID

	// TODO: load user from DB by userID
	return &UserDTO{}, nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func (s *AuthService) buildAuthResponse(user *domain.User) (*AuthResponse, error) {
	expiresAt := time.Now().Add(s.cfg.ExpiresIn)

	claims := middleware.DentDeskClaims{
		ClinicID: user.ClinicID,
		UserID:   user.ID,
		Role:     string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.cfg.Secret))
	if err != nil {
		return nil, fmt.Errorf("sign token: %w", err)
	}

	return &AuthResponse{
		Token:     signed,
		ExpiresAt: expiresAt,
		User: UserDTO{
			ID:        user.ID,
			ClinicID:  user.ClinicID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Role:      user.Role,
		},
	}, nil
}

func slugify(s string) string {
	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'a' && c <= 'z', c >= '0' && c <= '9':
			result = append(result, c)
		case c >= 'A' && c <= 'Z':
			result = append(result, c+32)
		case c == ' ' || c == '-' || c == '_':
			result = append(result, '-')
		}
	}
	return string(result)
}
