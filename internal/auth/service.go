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

func (s *Service) ListUsers(ctx context.Context, clinicID uuid.UUID) ([]User, error) {
	var out []User
	err := s.db.SelectContext(ctx, &out,
		`SELECT id, clinic_id, email, password_hash, role, name
		 FROM users WHERE clinic_id=$1 ORDER BY name`, clinicID)
	return out, err
}

func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (*User, error) {
	var u User
	err := s.db.GetContext(ctx, &u,
		`SELECT id, clinic_id, email, password_hash, role, name FROM users WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Service) CreateUser(ctx context.Context, clinicID uuid.UUID, email, password, role, name string) (*User, error) {
	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	var u User
	err = s.db.GetContext(ctx, &u,
		`INSERT INTO users (clinic_id, email, password_hash, role, name)
		 VALUES ($1,$2,$3,$4,$5)
		 RETURNING id, clinic_id, email, password_hash, role, name`,
		clinicID, email, hash, role, name)
	return &u, err
}

func (s *Service) UpdateUser(ctx context.Context, id uuid.UUID, name, role string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE users SET name=$1, role=$2 WHERE id=$3`, name, role, id)
	return err
}

func (s *Service) DeleteUser(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM users WHERE id=$1`, id)
	return err
}

func (s *Service) ChangePassword(ctx context.Context, userID uuid.UUID, oldPw, newPw string) error {
	var u User
	if err := s.db.GetContext(ctx, &u,
		`SELECT id, clinic_id, email, password_hash, role, name FROM users WHERE id=$1`, userID); err != nil {
		return errs.ErrUnauthorized
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(oldPw)); err != nil {
		return errs.ErrUnauthorized
	}
	hash, err := HashPassword(newPw)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `UPDATE users SET password_hash=$1 WHERE id=$2`, hash, userID)
	return err
}
