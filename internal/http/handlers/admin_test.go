package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/dentdesk/dentdesk/internal/auth"
	"github.com/dentdesk/dentdesk/internal/http/middleware"
)

func createAdminContext(method, path string, body []byte, claims *auth.Claims) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	if claims != nil {
		c.Set(middleware.CtxClaims, claims)
	}
	return c, w
}

func TestAdminRoleHelpers(t *testing.T) {
	clinicID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name             string
		claims           *auth.Claims
		wantOwnerOrAdmin bool
		wantOwner        bool
	}{
		{
			name: "owner",
			claims: &auth.Claims{
				UserID:   userID,
				ClinicID: clinicID,
				Role:     "owner",
			},
			wantOwnerOrAdmin: true,
			wantOwner:        true,
		},
		{
			name: "admin",
			claims: &auth.Claims{
				UserID:   userID,
				ClinicID: clinicID,
				Role:     "admin",
			},
			wantOwnerOrAdmin: true,
			wantOwner:        false,
		},
		{
			name: "operator",
			claims: &auth.Claims{
				UserID:   userID,
				ClinicID: clinicID,
				Role:     "operator",
			},
			wantOwnerOrAdmin: false,
			wantOwner:        false,
		},
		{
			name:             "no claims",
			claims:           nil,
			wantOwnerOrAdmin: false,
			wantOwner:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := createAdminContext(http.MethodGet, "/test", nil, tt.claims)
			assert.Equal(t, tt.wantOwnerOrAdmin, isOwnerOrAdmin(c))
			assert.Equal(t, tt.wantOwner, isOwner(c))
		})
	}
}

func TestAdminHandler_UpdateClinic_ForbiddenWithoutAdminRole(t *testing.T) {
	handler := &AdminHandler{}
	c, w := createAdminContext(
		http.MethodPut,
		"/api/admin/clinic",
		[]byte(`{"name":"Clinic","timezone":"UTC"}`),
		&auth.Claims{UserID: uuid.New(), ClinicID: uuid.New(), Role: "operator"},
	)

	handler.UpdateClinic(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.JSONEq(t, `{"error":"forbidden"}`, w.Body.String())
}

func TestAdminHandler_UpdateClinic_BadRequestOnInvalidBody(t *testing.T) {
	handler := &AdminHandler{}
	c, w := createAdminContext(
		http.MethodPut,
		"/api/admin/clinic",
		[]byte(`{"name":"Clinic"}`),
		&auth.Claims{UserID: uuid.New(), ClinicID: uuid.New(), Role: "owner"},
	)

	handler.UpdateClinic(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"invalid body"}`, w.Body.String())
}

func TestAdminHandler_CreateUser_RejectsInvalidRoleBeforeServiceCall(t *testing.T) {
	handler := &AdminHandler{}
	c, w := createAdminContext(
		http.MethodPost,
		"/api/admin/users",
		[]byte(`{"email":"test@test.com","password":"pass","name":"John","role":"guest"}`),
		&auth.Claims{UserID: uuid.New(), ClinicID: uuid.New(), Role: "owner"},
	)

	handler.CreateUser(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"role must be owner, admin, or operator"}`, w.Body.String())
}

func TestAdminHandler_GetUser_BadID(t *testing.T) {
	handler := &AdminHandler{}
	c, w := createAdminContext(http.MethodGet, "/api/admin/users/not-a-uuid", nil, &auth.Claims{
		UserID:   uuid.New(),
		ClinicID: uuid.New(),
		Role:     "admin",
	})
	c.Params = []gin.Param{{Key: "id", Value: "not-a-uuid"}}

	handler.GetUser(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"bad id"}`, w.Body.String())
}

func TestAdminHandler_DeleteUser_ForbiddenForNonOwner(t *testing.T) {
	handler := &AdminHandler{}
	c, w := createAdminContext(http.MethodDelete, "/api/admin/users/"+uuid.NewString(), nil, &auth.Claims{
		UserID:   uuid.New(),
		ClinicID: uuid.New(),
		Role:     "admin",
	})
	c.Params = []gin.Param{{Key: "id", Value: uuid.NewString()}}

	handler.DeleteUser(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.JSONEq(t, `{"error":"forbidden: owner only"}`, w.Body.String())
}

func TestAdminHandler_ChangePassword_BadRequestOnInvalidBody(t *testing.T) {
	handler := &AdminHandler{}
	c, w := createAdminContext(
		http.MethodPost,
		"/api/admin/change-password",
		[]byte(`{"old_password":"old"}`),
		&auth.Claims{UserID: uuid.New(), ClinicID: uuid.New(), Role: "owner"},
	)

	handler.ChangePassword(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"invalid body"}`, w.Body.String())
}
