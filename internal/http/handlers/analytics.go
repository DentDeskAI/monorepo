package handlers

import (
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/dentdesk/dentdesk/internal/http/middleware"
	"github.com/dentdesk/dentdesk/internal/scheduler"
)

// MacDent appointment status codes.
const (
	mdStatusScheduled = 0
	mdStatusConfirmed = 1
	mdStatusCancelled = 2
	mdStatusCame      = 3
	mdStatusLeft      = 4
	mdStatusInProcess = 5
	mdStatusLate      = 6
)

type DashboardHandler struct {
	Sched *scheduler.Service
}

// ── /api/dashboard/today ─────────────────────────────────────────────────────

type statusCounts struct {
	Scheduled int `json:"scheduled"`
	Confirmed int `json:"confirmed"`
	Cancelled int `json:"cancelled"`
	Came      int `json:"came"`
	Left      int `json:"left"`
	InProcess int `json:"in_process"`
	Late      int `json:"late"`
	Completed int `json:"completed"`
}

type todayAppt struct {
	ID       int    `json:"id"`
	Start    string `json:"start"`
	End      string `json:"end"`
	Status   int    `json:"status"`
	DoctorID int    `json:"doctor_id"`
	Cabinet  string `json:"cabinet,omitempty"`
	IsFirst  bool   `json:"is_first"`
}

type todayResponse struct {
	Date            string        `json:"date"`
	Total           int           `json:"total"`
	Counts          statusCounts  `json:"counts"`
	Upcoming        []todayAppt   `json:"upcoming"`
	NewPatientsToday int          `json:"new_patients_today"`
}

// Today serves GET /api/dashboard/today.
// Returns a live status snapshot of today's appointments.
func (h *DashboardHandler) Today(c *gin.Context) {
	cl := middleware.ClaimsFrom(c)

	now := time.Now()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dayEnd := dayStart.Add(24 * time.Hour)

	resp, err := h.Sched.ListAppointments(c.Request.Context(), cl.ClinicID, dayStart, dayEnd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var counts statusCounts
	var upcoming []todayAppt
	newPatients := 0

	for _, a := range resp.Appointments {
		switch a.Status {
		case mdStatusScheduled:
			counts.Scheduled++
		case mdStatusConfirmed:
			counts.Confirmed++
		case mdStatusCancelled:
			counts.Cancelled++
		case mdStatusCame:
			counts.Came++
		case mdStatusLeft:
			counts.Left++
		case mdStatusInProcess:
			counts.InProcess++
		case mdStatusLate:
			counts.Late++
		default:
			counts.Completed++
		}

		if a.IsFirst {
			newPatients++
		}

		// Upcoming = confirmed or scheduled starting after now
		if (a.Status == mdStatusScheduled || a.Status == mdStatusConfirmed || a.Status == mdStatusLate) {
			upcoming = append(upcoming, todayAppt{
				ID:       a.ID,
				Start:    a.Start,
				End:      a.End,
				Status:   a.Status,
				DoctorID: a.Doctor,
				Cabinet:  a.Cabinet,
				IsFirst:  a.IsFirst,
			})
		}
	}

	sort.Slice(upcoming, func(i, j int) bool {
		return upcoming[i].Start < upcoming[j].Start
	})

	c.JSON(http.StatusOK, todayResponse{
		Date:             dayStart.Format("2006-01-02"),
		Total:            len(resp.Appointments),
		Counts:           counts,
		Upcoming:         upcoming,
		NewPatientsToday: newPatients,
	})
}

// ── /api/dashboard/stats ─────────────────────────────────────────────────────

type periodRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type doctorStats struct {
	DoctorID    int `json:"doctor_id"`
	Total       int `json:"total"`
	Completed   int `json:"completed"`
	Cancelled   int `json:"cancelled"`
	NewPatients int `json:"new_patients"`
}

type funnelData struct {
	Booked    int `json:"booked"`
	Confirmed int `json:"confirmed"`
	Came      int `json:"came"`
	Completed int `json:"completed"`
}

type statsResponse struct {
	Period          periodRange            `json:"period"`
	Total           int                    `json:"total"`
	Completed       int                    `json:"completed"`
	Cancelled       int                    `json:"cancelled"`
	NoShow          int                    `json:"no_show"`
	NewPatients     int                    `json:"new_patients"`
	CompletionRate  float64                `json:"completion_rate"`
	NewPatientRate  float64                `json:"new_patient_rate"`
	Funnel          funnelData             `json:"funnel"`
	ByDoctor        []doctorStats          `json:"by_doctor"`
	Heatmap         map[string][]int       `json:"heatmap"`
}

// Stats serves GET /api/dashboard/stats?from=&to=
func (h *DashboardHandler) Stats(c *gin.Context) {
	cl := middleware.ClaimsFrom(c)
	from, to := dashboardRange(c.Query("from"), c.Query("to"))

	resp, err := h.Sched.ListAppointments(c.Request.Context(), cl.ClinicID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	appts := resp.Appointments
	total := len(appts)

	byDoc := map[int]*doctorStats{}
	// heatmap: weekday name → 24-slot hour counts
	weekdays := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	heatmap := map[string][]int{}
	for _, d := range weekdays {
		heatmap[d] = make([]int, 24)
	}

	var completed, cancelled, noShow, newPatients int
	funnel := funnelData{Booked: total}

	for _, a := range appts {
		// per-doctor
		ds, ok := byDoc[a.Doctor]
		if !ok {
			ds = &doctorStats{DoctorID: a.Doctor}
			byDoc[a.Doctor] = ds
		}
		ds.Total++
		if a.IsFirst {
			ds.NewPatients++
			newPatients++
		}

		switch a.Status {
		case mdStatusConfirmed:
			funnel.Confirmed++
		case mdStatusCame, mdStatusLeft:
			funnel.Confirmed++
			funnel.Came++
		case mdStatusInProcess:
			funnel.Confirmed++
			funnel.Came++
		case mdStatusCancelled:
			cancelled++
			ds.Cancelled++
		case mdStatusScheduled, mdStatusLate:
			// no additional funnel step
		default:
			// status not in known codes — treat as completed
			funnel.Confirmed++
			funnel.Came++
			funnel.Completed++
			completed++
			ds.Completed++
		}

		// heatmap: parse start time
		startTime, parseErr := parseApptTime(a.Start)
		if parseErr == nil {
			wd := int(startTime.Weekday()+6) % 7 // Mon=0
			heatmap[weekdays[wd]][startTime.Hour()]++
		}
	}

	// no-show approximation: scheduled/late after the range has passed
	noShow = 0 // can't determine from status alone without comparing to current time

	docList := make([]doctorStats, 0, len(byDoc))
	for _, ds := range byDoc {
		docList = append(docList, *ds)
	}
	sort.Slice(docList, func(i, j int) bool {
		return docList[i].Total > docList[j].Total
	})

	var completionRate, newPatientRate float64
	if total > 0 {
		completionRate = float64(completed) / float64(total)
		newPatientRate = float64(newPatients) / float64(total)
	}

	c.JSON(http.StatusOK, statsResponse{
		Period:         periodRange{From: from.Format(time.RFC3339), To: to.Format(time.RFC3339)},
		Total:          total,
		Completed:      completed,
		Cancelled:      cancelled,
		NoShow:         noShow,
		NewPatients:    newPatients,
		CompletionRate: round2(completionRate),
		NewPatientRate: round2(newPatientRate),
		Funnel:         funnel,
		ByDoctor:       docList,
		Heatmap:        heatmap,
	})
}

// ── /api/dashboard/revenue ───────────────────────────────────────────────────

type paymentTypeEntry struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
}

type revenueTrendPoint struct {
	Date    string  `json:"date"`
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
}

type revenueResponse struct {
	Period        periodRange         `json:"period"`
	TotalIncome   float64             `json:"total_income"`
	TotalExpense  float64             `json:"total_expense"`
	Net           float64             `json:"net"`
	ByType        []paymentTypeEntry  `json:"by_type"`
	Trend         []revenueTrendPoint `json:"trend"`
}

// Revenue serves GET /api/dashboard/revenue?from=&to=
func (h *DashboardHandler) Revenue(c *gin.Context) {
	cl := middleware.ClaimsFrom(c)
	from, to := dashboardRange(c.Query("from"), c.Query("to"))

	rashodi, err := h.Sched.GetRashodi(c.Request.Context(), cl.ClinicID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	byType := map[string]float64{}
	trendMap := map[string]*revenueTrendPoint{}
	var totalIncome, totalExpense float64

	for _, r := range rashodi {
		summ := r.SummFloat()
		typeName := r.TypeOplata
		if typeName == "" {
			typeName = "Other"
		}

		date := rashodDate(r.Date)

		if r.Type == 1 { // income
			totalIncome += summ
			byType[typeName] += summ
			pt := trendPointFor(trendMap, date)
			pt.Income += summ
		} else if r.Type == 2 { // expense
			totalExpense += summ
			pt := trendPointFor(trendMap, date)
			pt.Expense += summ
		}
	}

	// Build sorted type list
	typeList := make([]paymentTypeEntry, 0, len(byType))
	for name, amount := range byType {
		typeList = append(typeList, paymentTypeEntry{Name: name, Amount: amount})
	}
	sort.Slice(typeList, func(i, j int) bool {
		return typeList[i].Amount > typeList[j].Amount
	})

	// Build sorted trend list
	trendList := make([]revenueTrendPoint, 0, len(trendMap))
	for _, pt := range trendMap {
		trendList = append(trendList, *pt)
	}
	sort.Slice(trendList, func(i, j int) bool {
		return trendList[i].Date < trendList[j].Date
	})

	c.JSON(http.StatusOK, revenueResponse{
		Period:       periodRange{From: from.Format(time.RFC3339), To: to.Format(time.RFC3339)},
		TotalIncome:  totalIncome,
		TotalExpense: totalExpense,
		Net:          totalIncome - totalExpense,
		ByType:       typeList,
		Trend:        trendList,
	})
}

// ── helpers ───────────────────────────────────────────────────────────────────

// dashboardRange parses RFC3339/RFC3339Nano from/to; defaults to the current calendar month.
// JavaScript toISOString() produces milliseconds (.000Z) which require RFC3339Nano.
func dashboardRange(fromStr, toStr string) (time.Time, time.Time) {
	from, err1 := parseFlexRFC3339(fromStr)
	to, err2 := parseFlexRFC3339(toStr)
	if err1 == nil && err2 == nil {
		return from, to
	}
	now := time.Now()
	from = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	to = from.AddDate(0, 1, 0)
	return from, to
}

// parseFlexRFC3339 tries RFC3339Nano first (handles JS toISOString() with .000Z),
// then falls back to plain RFC3339.
func parseFlexRFC3339(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339, s)
}

// parseApptTime parses MacDent datetime strings like "28.04.2026 10:45:00".
func parseApptTime(s string) (time.Time, error) {
	return time.Parse("02.01.2006 15:04:05", s)
}

// rashodDate extracts a YYYY-MM-DD date key from a MacDent date string.
func rashodDate(s string) string {
	t, err := time.Parse("02.01.2006 15:04:05", s)
	if err != nil {
		t, err = time.Parse("02.01.2006", s)
		if err != nil {
			return s
		}
	}
	return t.Format("2006-01-02")
}

func trendPointFor(m map[string]*revenueTrendPoint, date string) *revenueTrendPoint {
	pt, ok := m[date]
	if !ok {
		pt = &revenueTrendPoint{Date: date}
		m[date] = pt
	}
	return pt
}

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

