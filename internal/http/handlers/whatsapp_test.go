package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestWhatsAppHandler_Verify(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "valid challenge",
			query:      "/webhook/whatsapp?hub.mode=subscribe&hub.verify_token=secret&hub.challenge=abc123",
			wantStatus: http.StatusOK,
			wantBody:   "abc123",
		},
		{
			name:       "invalid token",
			query:      "/webhook/whatsapp?hub.mode=subscribe&hub.verify_token=wrong&hub.challenge=abc123",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "invalid mode",
			query:      "/webhook/whatsapp?hub.mode=ping&hub.verify_token=secret&hub.challenge=abc123",
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, tt.query, nil)

			handler := &WhatsAppHandler{VerifyToken: "secret"}
			handler.Verify(c)

			assert.Equal(t, tt.wantStatus, c.Writer.Status())
			assert.Equal(t, tt.wantBody, w.Body.String())
		})
	}
}
