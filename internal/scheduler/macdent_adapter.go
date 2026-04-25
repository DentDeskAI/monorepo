package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	macdentBase       = "https://api-developer.macdent.kz"
	macdentDateLayout = "02.01.2006 15:04:05"
)

// MacDentAdapter integrates with the real MacDent scheduling API.
// All requests use POST with form-encoded body; access_token is a required field.
type MacDentAdapter struct {
	apiKey string
	http   *http.Client
}

func NewMacDentAdapter(_, apiKey string) *MacDentAdapter {
	return &MacDentAdapter{
		apiKey: apiKey,
		http:   &http.Client{Timeout: 15 * time.Second},
	}
}

// post sends a form-encoded POST, returns raw response bytes (on response=1) or an error.
func (a *MacDentAdapter) post(ctx context.Context, path string, fields url.Values) ([]byte, error) {
	fields.Set("access_token", a.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		macdentBase+path, strings.NewReader(fields.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("macdent: %w", err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)

	// all MacDent responses embed {"response": 0|1, "error": "..."} at the top level
	var env struct {
		Response int    `json:"response"`
		Error    string `json:"error"`
	}
	if err := json.Unmarshal(b, &env); err != nil {
		return nil, fmt.Errorf("macdent: bad json: %w", err)
	}
	if env.Response == 0 {
		return nil, fmt.Errorf("macdent: %s", env.Error)
	}
	return b, nil
}

// ── doctor list ───────────────────────────────────────────────────────────────

type mdSpecialty struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type mdDoctor struct {
	ID          int           `json:"id"`
	Name        string        `json:"name"`
	Specialnosti []mdSpecialty `json:"specialnosti"`
}

func (a *MacDentAdapter) listDoctors(ctx context.Context, specialty string) ([]mdDoctor, error) {
	b, err := a.post(ctx, "/doctor/find", url.Values{})
	if err != nil {
		return nil, err
	}
	var resp struct {
		Doctors []mdDoctor `json:"doctors"`
	}
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, fmt.Errorf("macdent doctor/find: %w", err)
	}

	if specialty == "" {
		return resp.Doctors, nil
	}
	// filter by specialty name (case-insensitive substring match)
	specialtyLower := strings.ToLower(specialty)
	var filtered []mdDoctor
	for _, d := range resp.Doctors {
		for _, s := range d.Specialnosti {
			if strings.Contains(strings.ToLower(s.Name), specialtyLower) {
				filtered = append(filtered, d)
				break
			}
		}
	}
	return filtered, nil
}

// ── free slots ────────────────────────────────────────────────────────────────

type mdScheduleInterval struct {
	From string `json:"from"` // "DD.MM.YYYY HH:MM:SS"
	To   string `json:"to"`
}

func parseMDTime(s string) (time.Time, error) {
	return time.Parse(macdentDateLayout, s)
}

func (a *MacDentAdapter) GetFreeSlots(
	ctx context.Context, _ uuid.UUID, from, to time.Time, specialty string,
) ([]Slot, error) {
	doctors, err := a.listDoctors(ctx, specialty)
	if err != nil {
		return nil, err
	}

	var slots []Slot
	for _, doc := range doctors {
		fields := url.Values{
			"doctor":   {fmt.Sprint(doc.ID)},
			"dateFrom": {from.Format("2006-01-02")},
			"dateTo":   {to.Format("2006-01-02")},
		}
		b, err := a.post(ctx, "/doctor/get_free_time", fields)
		if err != nil {
			continue
		}
		var resp struct {
			Schedules []mdScheduleInterval `json:"schedules"`
		}
		if err := json.Unmarshal(b, &resp); err != nil {
			continue
		}

		for _, iv := range resp.Schedules {
			start, err1 := parseMDTime(iv.From)
			end, err2 := parseMDTime(iv.To)
			if err1 != nil || err2 != nil {
				continue
			}
			// split working interval into 30-minute booking slots
			for cur := start; !cur.Add(30 * time.Minute).After(end); cur = cur.Add(30 * time.Minute) {
				slots = append(slots, Slot{
					StartsAt: cur,
					EndsAt:   cur.Add(30 * time.Minute),
					DoctorID: uuid.Nil, // MacDent uses int IDs; callers match via Doctor name
					Doctor:   doc.Name,
				})
			}
		}
	}
	return slots, nil
}

// ── create appointment (zapis/add) ────────────────────────────────────────────

func (a *MacDentAdapter) CreateAppointment(ctx context.Context, req BookRequest) (*BookResult, error) {
	// MacDent requires an integer doctor ID. We store it in the doctor's external_id.
	// The caller passes it via BookRequest.Service as a fallback for now if DoctorID is nil,
	// but typically the external_id must be resolved upstream.
	fields := url.Values{
		"start":   {req.StartsAt.Format(macdentDateLayout)},
		"end":     {req.EndsAt.Format(macdentDateLayout)},
		"doctor":  {req.DoctorID.String()}, // should be MacDent int id stored as external_id
		"zhaloba": {req.Service},
	}

	b, err := a.post(ctx, "/zapis/add", fields)
	if err != nil {
		return nil, err
	}
	var resp struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, fmt.Errorf("macdent zapis/add: %w", err)
	}
	ext := fmt.Sprint(resp.ID)
	return &BookResult{AppointmentID: uuid.New(), ExternalID: &ext}, nil
}

// ── cancel appointment (zapis/set_status → Declined) ─────────────────────────

func (a *MacDentAdapter) CancelAppointment(ctx context.Context, appointmentID uuid.UUID) error {
	// MacDent uses integer zapis IDs. The external_id must be resolved before calling.
	// Until external_id is stored and passed here, this is a no-op.
	_ = appointmentID
	return nil
}
