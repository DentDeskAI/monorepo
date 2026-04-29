package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/dentdesk/dentdesk/internal/auth"
	"github.com/dentdesk/dentdesk/internal/http/middleware"
)

// MockAdminService implements the interface needed by AdminHandler
type MockAdminService struct {
	mock.Mock
}

func (m *MockAdminService) Register(ctx interface{}, clinicName, timezone, ownerName, email, password string) (interface{}, interface{}, string, error) {
	args := m.Called(ctx, clinicName, timezone, ownerName, email, password)
	return args.Get(0), args.Get(1), args.String(2), args.Error(3)
}

func (m *MockAdminService) GetClinic(ctx interface{}, clinicID uuid.UUID) (interface{}, error) {
	args := m.Called(ctx, clinicID)
	return args.Get(0), args.Error(1)
}

func (m *MockAdminService) UpdateClinic(ctx interface{}, clinicID uuid.UUID, name, timezone, workingHours, schedulerType string, slotDuration int) (interface{}, error) {
	args := m.Called(ctx, clinicID, name, timezone, workingHours, schedulerType, slotDuration)
	return args.Get(0), args.Error(1)
}

func (m *MockAdminService) ListUsers(ctx interface{}, clinicID uuid.UUID) (interface{}, error) {
	args := m.Called(ctx, clinicID)
	return args.Get(0), args.Error(1)
}

func (m *MockAdminService) CreateUser(ctx interface{}, clinicID uuid.UUID, email, password, role, name string) (interface{}, error) {
	args := m.Called(ctx, clinicID, email, password, role, name)
	return args.Get(0), args.Error(1)
}

func (m *MockAdminService) GetUser(ctx interface{}, id uuid.UUID) (interface{}, error) {
	args := m.Called(ctx, id)
	return args.Get(0), args.Error(1)
}

func (m *MockAdminService) UpdateUser(ctx interface{}, id uuid.UUID, name, role string) error {
	args := m.Called(ctx, id, name, role)
	return args.Error(0)
}

func (m *MockAdminService) DeleteUser(ctx interface{}, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAdminService) ChangePassword(ctx interface{}, userID uuid.UUID, oldPassword, newPassword string) error {
	args := m.Called(ctx, userID, oldPassword, newPassword)
	return args.Error(0)
}

func createAuthContext(role string, clinicID uuid.UUID, userID uuid.UUID) *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	claims := &auth.Claims{
		UserID:   userID,
		ClinicID: clinicID,
		Role:     role,
	}
	c.Set(middleware.CtxClaims, claims)
	return c
}

func TestAdminHandler_Register(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		mockError      error
		expectedStatus int
	}{
		{
			name:           "Failure - Invalid JSON",
			requestBody:    `{invalid json`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Failure - Missing required fields",
			requestBody:    `{}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Failure - Registration error",
			requestBody:    `{"clinic_name":"Test","timezone":"UTC","owner_name":"John","email":"test@test.com","password":"pass"}`,
			mockError:      assert.AnError,
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockAdminService)
			handler := &AdminHandler{Svc: mockSvc}

			c := createAuthContext("owner", uuid.New(), uuid.New())
			c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/register", bytes.NewBufferString(tt.requestBody))
			c.Request.Header.Set("Content-Type", "application/json")

			if tt.mockError != nil {
				mockSvc.On("Register", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, nil, "", tt.mockError)
			}

			handler.Register(c)

			assert.Equal(t, tt.expectedStatus, c.Writer.Status())
		})
	}
}

func TestAdminHandler_GetClinic(t *testing.T) {
	clinicID := uuid.New()
	userID := uuid.New()
	mockSvc := new(MockAdminService)
	handler := &AdminHandler{Svc: mockSvc}

	c := createAuthContext("owner", clinicID, userID)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/clinic", nil)

	mockClinic := map[string]interface{}{
		"id":         clinicID.String(),
		"name":       "Test Clinic",
		"timezone":   "UTC",
		"clinic_id":  clinicID.String(),
	}
	mockSvc.On("GetClinic", mock.Anything, clinicID).Return(mockClinic, nil)

	handler.GetClinic(c)

	assert.Equal(t, http.StatusOK, c.Writer.Status())
}

func TestAdminHandler_UpdateClinic(t *testing.T) {
	clinicID := uuid.New()
	userID := uuid.New()
	mockSvc := new(MockAdminService)
	handler := &AdminHandler{Svc: mockSvc}

	c := createAuthContext("owner", clinicID, userID)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/admin/clinic", bytes.NewBufferString(`{"name":"New Name","timezone":"UTC"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	mockSvc.On("UpdateClinic", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)

	handler.UpdateClinic(c)

	assert.Equal(t, http.StatusOK, c.Writer.Status())
}

func TestAdminHandler_ListUsers(t *testing.T) {
	clinicID := uuid.New()
	userID := uuid.New()
	mockSvc := new(MockAdminService)
	handler := &AdminHandler{Svc: mockSvc}

	c := createAuthContext("admin", clinicID, userID)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)

	mockSvc.On("ListUsers", mock.Anything, clinicID).Return([]interface{}{}, nil)

	handler.ListUsers(c)

	assert.Equal(t, http.StatusOK, c.Writer.Status())
}

func TestAdminHandler_CreateUser(t *testing.T) {
	clinicID := uuid.New()
	userID := uuid.New()
	mockSvc := new(MockAdminService)
	handler := &AdminHandler{Svc: mockSvc}

	c := createAuthContext("owner", clinicID, userID)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/users", bytes.NewBufferString(`{"email":"test@test.com","password":"pass","name":"John","role":"admin"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	mockSvc.On("CreateUser", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)

	handler.CreateUser(c)

	assert.Equal(t, http.StatusCreated, c.Writer.Status())
}

func TestAdminHandler_GetUser(t *testing.T) {
	clinicID := uuid.New()
	userID := uuid.New()
	mockSvc := new(MockAdminService)
	handler := &AdminHandler{Svc: mockSvc}

	testID := uuid.New()
	c := createAuthContext("owner", clinicID, userID)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/users/"+testID.String(), nil)
	c.Params = []gin.Param{{Key: "id", Value: testID.String()}}

	mockSvc.On("GetUser", mock.Anything, testID).Return(nil, nil)

	handler.GetUser(c)

	assert.Equal(t, http.StatusOK, c.Writer.Status())
}

func TestAdminHandler_UpdateUser(t *testing.T) {
	clinicID := uuid.New()
	userID := uuid.New()
	mockSvc := new(MockAdminService)
	handler := &AdminHandler{Svc: mockSvc}

	testID := uuid.New()
	c := createAuthContext("owner", clinicID, userID)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/admin/users/"+testID.String(), bytes.NewBufferString(`{"name":"New Name","role":"admin"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = []gin.Param{{Key: "id", Value: testID.String()}}

	mockSvc.On("UpdateUser", mock.Anything, testID, mock.Anything, mock.Anything).Return(nil)

	handler.UpdateUser(c)

	assert.Equal(t, http.StatusOK, c.Writer.Status())
}

func TestAdminHandler_DeleteUser(t *testing.T) {
	clinicID := uuid.New()
	userID := uuid.New()
	mockSvc := new(MockAdminService)
	handler := &AdminHandler{Svc: mockSvc}

	testID := uuid.New()
	c := createAuthContext("owner", clinicID, userID)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/admin/users/"+testID.String(), nil)
	c.Params = []gin.Param{{Key: "id", Value: testID.String()}}

	mockSvc.On("DeleteUser", mock.Anything, testID).Return(nil)

	handler.DeleteUser(c)

	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
}

func TestAdminHandler_ChangePassword(t *testing.T) {
	clinicID := uuid.New()
	userID := uuid.New()
	mockSvc := new(MockAdminService)
	handler := &AdminHandler{Svc: mockSvc}

	c := createAuthContext("owner", clinicID, userID)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/change-password", bytes.NewBufferString(`{"old_password":"old","new_password":"new"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	mockSvc.On("ChangePassword", mock.Anything, userID, mock.Anything, mock.Anything).Return(nil)

	handler.ChangePassword(c)

	assert.Equal(t, http.StatusOK, c.Writer.Status())
}
