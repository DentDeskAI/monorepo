package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/dentdesk/dentdesk/internal/auth"
	"github.com/dentdesk/dentdesk/internal/http/middleware"
	errs "github.com/dentdesk/dentdesk/internal/platform/errors"
)

// MockAuthService implements AuthService for testing
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(ctx context.Context, email, password string) (string, *auth.User, error) {
	args := m.Called(ctx, email, password)
	token := args.String(0)
	err := args.Error(2)

	if err != nil {
		return "", nil, err
	}

	user, _ := args.Get(1).(*auth.User)
	return token, user, nil
}

func (m *MockAuthService) Parse(token string) (*auth.Claims, error) {
	args := m.Called(token)
	return args.Get(0).(*auth.Claims), args.Error(1)
}

// createGinContextWithClaims is a helper to create a gin.Context with mocked claims
func createGinContextWithClaims(claims *auth.Claims) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)

	// Inject claims into the context
	c.Set(middleware.CtxClaims, claims)

	return c, w
}

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockLoginErr   error
		mockUser       *auth.User
		mockToken      string
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "Success - Login with valid credentials",
			requestBody: loginReq{
				Email:    "admin@test.kz",
				Password: "demo1234",
			},
			mockToken: "valid-jwt-token",
			mockUser: &auth.User{
				ID:       uuid.New(),
				Email:    "admin@test.kz",
				Name:     "Admin User",
				Role:     "admin",
				ClinicID: uuid.New(),
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"token": "valid-jwt-token",
			},
		},
		{
			name: "Failure - Invalid credentials",
			requestBody: loginReq{
				Email:    "bad@email.com",
				Password: "wrongpass",
			},
			mockLoginErr:   errs.ErrUnauthorized,
			expectedStatus: http.StatusUnauthorized,
			expectedBody: map[string]interface{}{
				"error": "invalid credentials",
			},
		},
		{
			name: "Failure - Internal server error",
			requestBody: loginReq{
				Email:    "user@test.com",
				Password: "pass123",
			},
			mockLoginErr:   errors.New("database connection failed"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error": "internal",
			},
		},
		{
			name:           "Failure - Invalid JSON body",
			requestBody:    nil, // Will send invalid JSON
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "invalid body",
			},
		},
		{
			name: "Failure - Missing required fields",
			requestBody: loginReq{
				Email:    "",
				Password: "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "invalid body",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockAuthService)
			handler := &AuthHandler{Svc: mockSvc}

			var reqBody []byte
			var req *http.Request

			if tt.requestBody == nil {
				// Simulate invalid JSON
				reqBody = []byte(`{invalid json}`)
				req = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(reqBody))
				req.Header.Set("Content-Type", "application/json")
			} else {
				reqBody, _ = json.Marshal(tt.requestBody)
				req = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(reqBody))
				req.Header.Set("Content-Type", "application/json")
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Setup mock expectations
			if tt.name != "Failure - Invalid JSON body" && tt.name != "Failure - Missing required fields" {
				if tt.mockLoginErr != nil {
					mockSvc.On("Login", mock.Anything, mock.Anything, mock.Anything).Return("", nil, tt.mockLoginErr)
				} else {
					mockSvc.On("Login", mock.Anything, mock.Anything, mock.Anything).Return(tt.mockToken, tt.mockUser, nil)
				}
			}

			// Call the handler
			handler.Login(c)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Assert response body
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.expectedBody != nil {
				for k, v := range tt.expectedBody {
					assert.Equal(t, v, response[k], "Mismatch in key %s", k)
				}
			}

			// Verify mock was called correctly
			if tt.name != "Failure - Invalid JSON body" && tt.name != "Failure - Missing required fields" {
				mockSvc.AssertExpectations(t)
			}
		})
	}
}

func TestAuthHandler_Me(t *testing.T) {
	tests := []struct {
		name           string
		claims         *auth.Claims
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "Success - Valid claims",
			claims: &auth.Claims{
				UserID:   uuid.New(),
				ClinicID: uuid.New(),
				Role:     "admin",
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"role": "admin",
			},
		},
		{
			name:           "Failure - No claims present",
			claims:         nil,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   nil, // No body expected, just status
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockAuthService)
			handler := &AuthHandler{Svc: mockSvc}

			var c *gin.Context
			var w *httptest.ResponseRecorder
			if tt.claims != nil {
				c, w = createGinContextWithClaims(tt.claims)
			} else {
				gin.SetMode(gin.TestMode)
				w = httptest.NewRecorder()
				c, _ = gin.CreateTestContext(w)
				c.Request = httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
				// Don't set any claims so ClaimsFrom returns nil
			}

			handler.Me(c)

			// Use c.Writer.Status() to get the status code set by the handler
			assert.Equal(t, tt.expectedStatus, c.Writer.Status())

			if tt.expectedBody != nil {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				for k, v := range tt.expectedBody {
					assert.Equal(t, v, response[k], "Mismatch in key %s", k)
				}
			}
		})
	}
}
