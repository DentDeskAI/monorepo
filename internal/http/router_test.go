package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dentdesk/dentdesk/internal/auth"
	"github.com/dentdesk/dentdesk/internal/http/handlers"
)

func newTestRouter() *Router {
	return &Router{
		AuthSvc:    auth.NewService(nil, "test-secret"),
		Log:        zerolog.Nop(),
		Origin:     "http://localhost:5173",
		AuthH:      &handlers.AuthHandler{},
		AdminH:     &handlers.AdminHandler{},
		CRMH:       &handlers.CRMHandler{},
		ResourceH:  &handlers.ResourceHandler{},
		ScheduleH:  &handlers.SchedulingHandler{},
		DashboardH: &handlers.DashboardHandler{},
		WhatsApp:   &handlers.WhatsAppHandler{},
	}
}

func TestRouterBuild_Healthz(t *testing.T) {
	app := newTestRouter().Build()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	app.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status":"ok"}`, w.Body.String())
}

func TestRouterBuild_RegistersKeyRoutes(t *testing.T) {
	app := newTestRouter().Build()

	routes := app.Routes()
	require.NotEmpty(t, routes)

	routeSet := map[string]bool{}
	for _, rt := range routes {
		routeSet[rt.Method+" "+rt.Path] = true
	}

	expected := []string{
		"POST /api/register",
		"POST /api/auth/login",
		"GET /webhook/whatsapp",
		"POST /webhook/whatsapp",
		"GET /api/auth/me",
		"POST /api/auth/change-password",
		"GET /api/users",
		"POST /api/users",
		"GET /api/dashboard/today",
		"GET /api/dashboard/stats",
		"GET /api/dashboard/revenue",
		"GET /api/events",
	}

	for _, route := range expected {
		assert.Truef(t, routeSet[route], "missing route %s", route)
	}
}
