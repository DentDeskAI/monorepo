package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"

	errs "github.com/dentdesk/dentdesk/internal/platform/errors"
)

type User struct {
	ID           uuid.UUID `db:"id"`
	ClinicID     uuid.UUID `db:"clinic_id"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	Role         string    `db:"role"`
	Name         string    `db:"name"`
}

type Claims struct {
	UserID   uuid.UUID `json:"uid"`
	ClinicID uuid.UUID `json:"cid"`
	Role     string    `json:"role"`
	jwt.RegisteredClaims
}

type Service struct {
	db     *sqlx.DB
	secret []byte
	ttl    time.Duration
}

func NewService(db *sqlx.DB, secret string) *Service {
	return &Service{
		db:     db,
		secret: []byte(secret),
		ttl:    12 * time.Hour,
	}
}

func (s *Service) Login(ctx context.Context, email, password string) (string, *User, error) {
	var u User
	err := s.db.GetContext(ctx, &u,
		`SELECT id, clinic_id, email, password_hash, role, name FROM users WHERE email=$1`, email)
	if err != nil {
		return "", nil, errs.ErrUnauthorized
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", nil, errs.ErrUnauthorized
	}
	token, err := s.issueToken(u)
	if err != nil {
		return "", nil, err
	}
	return token, &u, nil
}

func (s *Service) issueToken(u User) (string, error) {
	claims := Claims{
		UserID:   u.ID,
		ClinicID: u.ClinicID,
		Role:     u.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   u.ID.String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *Service) Parse(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("bad signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, errs.ErrUnauthorized
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errs.ErrUnauthorized
	}
	return claims, nil
}

// HashPassword — утилита для seed / регистрации.
func HashPassword(pw string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(h), err
}
