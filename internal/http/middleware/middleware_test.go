package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dentdesk/dentdesk/internal/auth"
)

func issueTestToken(t *testing.T, secret string, claims *auth.Claims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	require.NoError(t, err)
	return signed
}

func TestClaimsFrom(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	assert.Nil(t, ClaimsFrom(c))

	claims := &auth.Claims{UserID: uuid.New(), ClinicID: uuid.New(), Role: "admin"}
	c.Set(CtxClaims, claims)
	assert.Same(t, claims, ClaimsFrom(c))
}

func TestAuthRequired_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := auth.NewService(nil, "secret")
	r := gin.New()
	r.GET("/protected", AuthRequired(svc), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.JSONEq(t, `{"error":"missing token"}`, w.Body.String())
}

func TestAuthRequired_BearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "secret"
	claims := &auth.Claims{
		UserID:   uuid.New(),
		ClinicID: uuid.New(),
		Role:     "owner",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			Subject:   uuid.NewString(),
		},
	}
	svc := auth.NewService(nil, secret)
	r := gin.New()
	r.GET("/protected", AuthRequired(svc), func(c *gin.Context) {
		got := ClaimsFrom(c)
		c.JSON(http.StatusOK, gin.H{
			"user_id":   got.UserID.String(),
			"clinic_id": got.ClinicID.String(),
			"role":      got.Role,
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+issueTestToken(t, secret, claims))
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, claims.UserID.String(), resp["user_id"])
	assert.Equal(t, claims.ClinicID.String(), resp["clinic_id"])
	assert.Equal(t, "owner", resp["role"])
}

func TestAuthRequired_QueryToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "secret"
	claims := &auth.Claims{
		UserID:   uuid.New(),
		ClinicID: uuid.New(),
		Role:     "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	svc := auth.NewService(nil, secret)
	r := gin.New()
	r.GET("/events", AuthRequired(svc), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/events?token="+issueTestToken(t, secret, claims), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestAuthRequired_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := auth.NewService(nil, "secret")
	r := gin.New()
	r.GET("/protected", AuthRequired(svc), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected?token=bad-token", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.JSONEq(t, `{"error":"invalid token"}`, w.Body.String())
}

func TestRecoverReturnsInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Recover(zerolog.Nop()))
	r.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.JSONEq(t, `{"error":"internal"}`, w.Body.String())
}
