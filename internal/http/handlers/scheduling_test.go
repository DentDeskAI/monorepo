package handlers

import (
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dentdesk/dentdesk/internal/auth"
)

func TestSchedulingHandler_GetPatient_BadID(t *testing.T) {
	handler := &SchedulingHandler{}
	c, w := createAdminContext(http.MethodGet, "/api/patients/bad", nil, &auth.Claims{
		UserID:   uuid.New(),
		ClinicID: uuid.New(),
		Role:     "admin",
	})
	c.Params = []gin.Param{{Key: "id", Value: "bad"}}

	handler.GetPatient(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"bad id"}`, w.Body.String())
}

func TestSchedulingHandler_CreateAppointment_InvalidBody(t *testing.T) {
	handler := &SchedulingHandler{}
	c, w := createAdminContext(
		http.MethodPost,
		"/api/appointments",
		[]byte(`{"patient_id":"`+uuid.NewString()+`"}`),
		&auth.Claims{UserID: uuid.New(), ClinicID: uuid.New(), Role: "admin"},
	)

	handler.CreateAppointment(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"invalid body"}`, w.Body.String())
}

func TestSchedulingHandler_GetAppointment_BadID(t *testing.T) {
	handler := &SchedulingHandler{}
	c, w := createAdminContext(http.MethodGet, "/api/appointments/bad", nil, &auth.Claims{
		UserID:   uuid.New(),
		ClinicID: uuid.New(),
		Role:     "admin",
	})
	c.Params = []gin.Param{{Key: "id", Value: "bad"}}

	handler.GetAppointment(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"bad id"}`, w.Body.String())
}

func TestSchedulingHandler_UpdateAppointmentStatus_BadID(t *testing.T) {
	handler := &SchedulingHandler{}
	c, w := createAdminContext(
		http.MethodPut,
		"/api/appointments/bad/status",
		[]byte(`{"status":"confirmed"}`),
		&auth.Claims{UserID: uuid.New(), ClinicID: uuid.New(), Role: "admin"},
	)
	c.Params = []gin.Param{{Key: "id", Value: "bad"}}

	handler.UpdateAppointmentStatus(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"bad id"}`, w.Body.String())
}

func TestSchedulingHandler_GetConversation_BadID(t *testing.T) {
	handler := &SchedulingHandler{}
	c, w := createAdminContext(http.MethodGet, "/api/chats/bad", nil, &auth.Claims{
		UserID:   uuid.New(),
		ClinicID: uuid.New(),
		Role:     "admin",
	})
	c.Params = []gin.Param{{Key: "id", Value: "bad"}}

	handler.GetConversation(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"bad id"}`, w.Body.String())
}

func TestSchedulingHandler_SendAppointmentRequest_InvalidRange(t *testing.T) {
	handler := &SchedulingHandler{}
	c, w := createAdminContext(
		http.MethodPost,
		"/api/schedule/appointment-requests",
		[]byte(`{"patient_name":"John","patient_phone":"+7700","starts_at":"2026-04-30T10:00:00Z","ends_at":"2026-04-30T09:00:00Z"}`),
		&auth.Claims{UserID: uuid.New(), ClinicID: uuid.New(), Role: "admin"},
	)

	handler.SendAppointmentRequest(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"ends_at must be after starts_at"}`, w.Body.String())
}

func TestSchedulingHandler_CreateSchedulePatient_InvalidBody(t *testing.T) {
	handler := &SchedulingHandler{}
	c, w := createAdminContext(
		http.MethodPost,
		"/api/schedule/patients",
		[]byte(`{}`),
		&auth.Claims{UserID: uuid.New(), ClinicID: uuid.New(), Role: "admin"},
	)

	handler.CreateSchedulePatient(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"invalid body"}`, w.Body.String())
}

func TestSchedulingHandler_CreateScheduleAppointment_InvalidRange(t *testing.T) {
	handler := &SchedulingHandler{}
	c, w := createAdminContext(
		http.MethodPost,
		"/api/schedule/appointments",
		[]byte(`{"doctor_id":1,"patient_id":2,"starts_at":"2026-04-30T10:00:00Z","ends_at":"2026-04-30T09:00:00Z"}`),
		&auth.Claims{UserID: uuid.New(), ClinicID: uuid.New(), Role: "admin"},
	)

	handler.CreateScheduleAppointment(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"ends_at must be after starts_at"}`, w.Body.String())
}

func TestSchedulingHandler_SetScheduleAppointmentStatus_BadID(t *testing.T) {
	handler := &SchedulingHandler{}
	c, w := createAdminContext(
		http.MethodPut,
		"/api/schedule/appointments/bad/status",
		[]byte(`{"status":1}`),
		&auth.Claims{UserID: uuid.New(), ClinicID: uuid.New(), Role: "admin"},
	)
	c.Params = []gin.Param{{Key: "id", Value: "bad"}}

	handler.SetScheduleAppointmentStatus(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"bad id"}`, w.Body.String())
}

func TestSchedulingHandler_SetScheduleAppointmentStatus_InvalidStatus(t *testing.T) {
	handler := &SchedulingHandler{}
	c, w := createAdminContext(
		http.MethodPut,
		"/api/schedule/appointments/12/status",
		[]byte(`{"status":9}`),
		&auth.Claims{UserID: uuid.New(), ClinicID: uuid.New(), Role: "admin"},
	)
	c.Params = []gin.Param{{Key: "id", Value: "12"}}

	handler.SetScheduleAppointmentStatus(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"status must be 0..6"}`, w.Body.String())
}

func TestSchedulingHelpers(t *testing.T) {
	from, to := weekRange("2026-04-28T10:00:00.000Z", "2026-05-05T10:00:00Z")
	assert.Equal(t, "2026-04-28T10:00:00Z", from.UTC().Format(time.RFC3339))
	assert.Equal(t, "2026-05-05T10:00:00Z", to.UTC().Format(time.RFC3339))

	parsed, err := parseFlexTime("2026-04-28T10:00:00.000Z")
	require.NoError(t, err)
	assert.Equal(t, 2026, parsed.Year())

	_, err = parseFlexTime("not-a-time")
	assert.Error(t, err)
}
