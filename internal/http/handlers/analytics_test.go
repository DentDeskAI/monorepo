package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dentdesk/dentdesk/internal/auth"
	"github.com/dentdesk/dentdesk/internal/http/middleware"
	"github.com/dentdesk/dentdesk/internal/scheduling"
)

type mockScheduler struct {
	listAppointmentsFn func(ctx context.Context, clinicID uuid.UUID, from, to time.Time) (*scheduling.AppointmentsResponse, error)
	getRevenueFn       func(ctx context.Context, clinicID uuid.UUID, from, to time.Time) ([]scheduling.RevenueRecord, error)
}

func (m *mockScheduler) ListDoctors(context.Context, uuid.UUID) ([]scheduling.Doctor, error) {
	panic("unexpected ListDoctors call")
}

func (m *mockScheduler) GetDoctor(context.Context, uuid.UUID, string) (*scheduling.Doctor, error) {
	panic("unexpected GetDoctor call")
}

func (m *mockScheduler) ListPatients(context.Context, uuid.UUID) ([]scheduling.Patient, error) {
	panic("unexpected ListPatients call")
}

func (m *mockScheduler) GetPatient(context.Context, uuid.UUID, int) (*scheduling.Patient, error) {
	panic("unexpected GetPatient call")
}

func (m *mockScheduler) CreatePatient(context.Context, uuid.UUID, scheduling.CreatePatientParams) (*scheduling.Patient, error) {
	panic("unexpected CreatePatient call")
}

func (m *mockScheduler) GetClinic(context.Context, uuid.UUID) (*scheduling.Stomatology, error) {
	panic("unexpected GetClinic call")
}

func (m *mockScheduler) ListAppointments(ctx context.Context, clinicID uuid.UUID, from, to time.Time) (*scheduling.AppointmentsResponse, error) {
	if m.listAppointmentsFn == nil {
		panic("unexpected ListAppointments call")
	}
	return m.listAppointmentsFn(ctx, clinicID, from, to)
}

func (m *mockScheduler) GetAppointmentByID(context.Context, uuid.UUID, int) (*scheduling.AppointmentDetail, error) {
	panic("unexpected GetAppointmentByID call")
}

func (m *mockScheduler) CreateAppointment(context.Context, scheduling.BookRequest) (*scheduling.BookResult, error) {
	panic("unexpected CreateAppointment call")
}

func (m *mockScheduler) CreateScheduleAppointment(context.Context, uuid.UUID, scheduling.ScheduleAppointmentParams) (*scheduling.ScheduleAppointmentResult, error) {
	panic("unexpected CreateScheduleAppointment call")
}

func (m *mockScheduler) UpdateAppointment(context.Context, uuid.UUID, int, scheduling.UpdateAppointmentParams) error {
	panic("unexpected UpdateAppointment call")
}

func (m *mockScheduler) RemoveAppointment(context.Context, uuid.UUID, int) error {
	panic("unexpected RemoveAppointment call")
}

func (m *mockScheduler) SetAppointmentStatus(context.Context, uuid.UUID, int, int) error {
	panic("unexpected SetAppointmentStatus call")
}

func (m *mockScheduler) SendAppointmentRequest(context.Context, uuid.UUID, scheduling.AppointmentRequestParams) (*scheduling.AppointmentRequestResult, error) {
	panic("unexpected SendAppointmentRequest call")
}

func (m *mockScheduler) GetFreeSlots(context.Context, uuid.UUID, time.Time, time.Time, string) ([]scheduling.Slot, error) {
	panic("unexpected GetFreeSlots call")
}

func (m *mockScheduler) GetRevenue(ctx context.Context, clinicID uuid.UUID, from, to time.Time) ([]scheduling.RevenueRecord, error) {
	if m.getRevenueFn == nil {
		panic("unexpected GetRevenue call")
	}
	return m.getRevenueFn(ctx, clinicID, from, to)
}

func (m *mockScheduler) GetHistory(context.Context, uuid.UUID, time.Time, time.Time) (*scheduling.AppointmentsResponse, error) {
	panic("unexpected GetHistory call")
}

func dashboardContext(path string, claims *auth.Claims) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, path, nil)
	c.Set(middleware.CtxClaims, claims)
	return c, w
}

func TestDashboardTodayAggregatesAppointments(t *testing.T) {
	clinicID := uuid.New()
	handler := &DashboardHandler{
		Sched: &mockScheduler{
			listAppointmentsFn: func(ctx context.Context, gotClinicID uuid.UUID, from, to time.Time) (*scheduling.AppointmentsResponse, error) {
				assert.Equal(t, clinicID, gotClinicID)
				assert.Equal(t, 0, from.Hour())
				assert.Equal(t, 0, from.Minute())
				assert.Equal(t, 24*time.Hour, to.Sub(from))
				return &scheduling.AppointmentsResponse{
					Appointments: []scheduling.Appointment{
						{ID: 1, Start: "30.04.2026 15:00:00", End: "30.04.2026 15:30:00", Status: mdStatusConfirmed, Doctor: 2, IsFirst: true, Cabinet: "A"},
						{ID: 2, Start: "30.04.2026 09:00:00", End: "30.04.2026 09:30:00", Status: mdStatusScheduled, Doctor: 1, Cabinet: "B"},
						{ID: 3, Start: "30.04.2026 18:00:00", End: "30.04.2026 18:30:00", Status: mdStatusLate, Doctor: 3},
						{ID: 4, Start: "30.04.2026 12:00:00", End: "30.04.2026 12:30:00", Status: mdStatusCancelled, Doctor: 4, IsFirst: true},
						{ID: 5, Start: "30.04.2026 13:00:00", End: "30.04.2026 13:30:00", Status: 99, Doctor: 5},
					},
				}, nil
			},
		},
	}

	c, w := dashboardContext("/api/dashboard/today", &auth.Claims{ClinicID: clinicID})
	handler.Today(c)

	require.Equal(t, http.StatusOK, w.Code)

	var resp todayResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, 5, resp.Total)
	assert.Equal(t, 1, resp.Counts.Scheduled)
	assert.Equal(t, 1, resp.Counts.Confirmed)
	assert.Equal(t, 1, resp.Counts.Cancelled)
	assert.Equal(t, 1, resp.Counts.Late)
	assert.Equal(t, 1, resp.Counts.Completed)
	assert.Equal(t, 2, resp.NewPatientsToday)
	require.Len(t, resp.Upcoming, 3)
	assert.Equal(t, []int{2, 1, 3}, []int{resp.Upcoming[0].ID, resp.Upcoming[1].ID, resp.Upcoming[2].ID})
}

func TestDashboardStatsAggregatesFunnelAndDoctorStats(t *testing.T) {
	clinicID := uuid.New()
	from := "2026-04-01T00:00:00Z"
	to := "2026-05-01T00:00:00Z"

	handler := &DashboardHandler{
		Sched: &mockScheduler{
			listAppointmentsFn: func(ctx context.Context, gotClinicID uuid.UUID, gotFrom, gotTo time.Time) (*scheduling.AppointmentsResponse, error) {
				assert.Equal(t, clinicID, gotClinicID)
				assert.Equal(t, from, gotFrom.UTC().Format(time.RFC3339))
				assert.Equal(t, to, gotTo.UTC().Format(time.RFC3339))
				return &scheduling.AppointmentsResponse{
					Appointments: []scheduling.Appointment{
						{ID: 1, Doctor: 1, Start: "28.04.2026 10:45:00", Status: mdStatusConfirmed, IsFirst: true},
						{ID: 2, Doctor: 1, Start: "28.04.2026 11:00:00", Status: mdStatusCancelled},
						{ID: 3, Doctor: 2, Start: "29.04.2026 14:00:00", Status: 99},
						{ID: 4, Doctor: 2, Start: "invalid", Status: mdStatusScheduled},
					},
				}, nil
			},
		},
	}

	c, w := dashboardContext("/api/dashboard/stats?from="+from+"&to="+to, &auth.Claims{ClinicID: clinicID})
	handler.Stats(c)

	require.Equal(t, http.StatusOK, w.Code)

	var resp statsResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, 4, resp.Total)
	assert.Equal(t, 1, resp.Completed)
	assert.Equal(t, 1, resp.Cancelled)
	assert.Equal(t, 1, resp.NewPatients)
	assert.Equal(t, 0.25, resp.CompletionRate)
	assert.Equal(t, 0.25, resp.NewPatientRate)
	assert.Equal(t, funnelData{Booked: 4, Confirmed: 2, Came: 1, Completed: 1}, resp.Funnel)
	require.Len(t, resp.ByDoctor, 2)
	assert.Equal(t, 1, resp.ByDoctor[0].DoctorID)
	assert.Equal(t, 2, resp.ByDoctor[0].Total)
	assert.Equal(t, 1, resp.ByDoctor[0].Cancelled)
	assert.Equal(t, 1, resp.ByDoctor[0].NewPatients)
	assert.Equal(t, 1, resp.Heatmap["Tue"][10])
	assert.Equal(t, 1, resp.Heatmap["Tue"][11])
	assert.Equal(t, 1, resp.Heatmap["Wed"][14])
}

func TestDashboardRevenueAggregatesTotalsAndTrend(t *testing.T) {
	clinicID := uuid.New()
	handler := &DashboardHandler{
		Sched: &mockScheduler{
			getRevenueFn: func(ctx context.Context, gotClinicID uuid.UUID, from, to time.Time) ([]scheduling.RevenueRecord, error) {
				assert.Equal(t, clinicID, gotClinicID)
				return []scheduling.RevenueRecord{
					{Date: "28.04.2026 10:00:00", Amount: 1000, Type: 1, PaymentType: "Cash"},
					{Date: "28.04.2026", Amount: 500, Type: 1, PaymentType: ""},
					{Date: "29.04.2026 09:00:00", Amount: 300, Type: 2, PaymentType: "Cash"},
				}, nil
			},
		},
	}

	c, w := dashboardContext("/api/dashboard/revenue", &auth.Claims{ClinicID: clinicID})
	handler.Revenue(c)

	require.Equal(t, http.StatusOK, w.Code)

	var resp revenueResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, 1500.0, resp.TotalIncome)
	assert.Equal(t, 300.0, resp.TotalExpense)
	assert.Equal(t, 1200.0, resp.Net)
	require.Len(t, resp.ByType, 2)
	assert.Equal(t, "Cash", resp.ByType[0].Name)
	assert.Equal(t, 1000.0, resp.ByType[0].Amount)
	assert.Equal(t, "Other", resp.ByType[1].Name)
	assert.Equal(t, 500.0, resp.ByType[1].Amount)
	require.Len(t, resp.Trend, 2)
	assert.Equal(t, "2026-04-28", resp.Trend[0].Date)
	assert.Equal(t, 1500.0, resp.Trend[0].Income)
	assert.Equal(t, "2026-04-29", resp.Trend[1].Date)
	assert.Equal(t, 300.0, resp.Trend[1].Expense)
}

func TestDashboardTodayReturnsInternalErrorOnSchedulerFailure(t *testing.T) {
	handler := &DashboardHandler{
		Sched: &mockScheduler{
			listAppointmentsFn: func(context.Context, uuid.UUID, time.Time, time.Time) (*scheduling.AppointmentsResponse, error) {
				return nil, errors.New("boom")
			},
		},
	}

	c, w := dashboardContext("/api/dashboard/today", &auth.Claims{ClinicID: uuid.New()})
	handler.Today(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.JSONEq(t, `{"error":"boom"}`, w.Body.String())
}

func TestDashboardHelpers(t *testing.T) {
	from, to := dashboardRange("2026-04-01T00:00:00.000Z", "2026-05-01T00:00:00Z")
	assert.Equal(t, "2026-04-01T00:00:00Z", from.UTC().Format(time.RFC3339))
	assert.Equal(t, "2026-05-01T00:00:00Z", to.UTC().Format(time.RFC3339))

	parsed, err := parseApptTime("28.04.2026 10:45:00")
	require.NoError(t, err)
	assert.Equal(t, 10, parsed.Hour())

	assert.Equal(t, "2026-04-28", rashodDate("28.04.2026 10:45:00"))
	assert.Equal(t, "2026-04-28", rashodDate("28.04.2026"))
	assert.Equal(t, "bad-date", rashodDate("bad-date"))

	m := map[string]*revenueTrendPoint{}
	pt := trendPointFor(m, "2026-04-28")
	pt.Income = 100
	assert.Same(t, pt, trendPointFor(m, "2026-04-28"))
	assert.Equal(t, 1.24, round2(1.235))
}
