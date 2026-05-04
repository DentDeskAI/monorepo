package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	errs "github.com/dentdesk/dentdesk/internal/platform/errors"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("demo1234")
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, "demo1234", hash)
}

func TestIssueTokenAndParse(t *testing.T) {
	svc := &Service{
		secret: []byte("test-secret"),
		ttl:    time.Hour,
	}
	user := User{
		ID:       uuid.New(),
		ClinicID: uuid.New(),
		Email:    "admin@test.kz",
		Role:     "admin",
		Name:     "Admin",
	}

	token, err := svc.issueToken(user)
	require.NoError(t, err)

	claims, err := svc.Parse(token)
	require.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.ClinicID, claims.ClinicID)
	assert.Equal(t, user.Role, claims.Role)
	assert.Equal(t, user.ID.String(), claims.Subject)
	require.NotNil(t, claims.ExpiresAt)
	assert.True(t, claims.ExpiresAt.Time.After(time.Now()))
}

func TestParseRejectsInvalidToken(t *testing.T) {
	svc := &Service{secret: []byte("test-secret")}

	claims, err := svc.Parse("not-a-jwt")
	assert.Nil(t, claims)
	assert.ErrorIs(t, err, errs.ErrUnauthorized)
}

func TestParseRejectsTokenSignedWithDifferentSecret(t *testing.T) {
	issuer := &Service{secret: []byte("issuer-secret"), ttl: time.Hour}
	parser := &Service{secret: []byte("parser-secret")}

	token, err := issuer.issueToken(User{
		ID:       uuid.New(),
		ClinicID: uuid.New(),
		Role:     "operator",
	})
	require.NoError(t, err)

	claims, parseErr := parser.Parse(token)
	assert.Nil(t, claims)
	assert.ErrorIs(t, parseErr, errs.ErrUnauthorized)
}
