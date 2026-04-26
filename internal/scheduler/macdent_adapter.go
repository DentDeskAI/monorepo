package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const (
	macdentBase       = "https://api-developer.macdent.kz"
	macdentDateLayout = "02.01.2006 15:04:05"
)

// MacDentAdapter integrates with the real MacDent scheduling API.
// All requests use GET with access_token as a query parameter.
type MacDentAdapter struct {
	db   *sqlx.DB
	http *http.Client
}

func NewMacDentAdapter(db *sqlx.DB) *MacDentAdapter {
	return &MacDentAdapter{
		db: db,
		http: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (a *MacDentAdapter) GetClinicAPI(ctx context.Context, clinicID uuid.UUID) (string, error) {
	var api string
	err := a.db.GetContext(ctx, &api, `SELECT macdent_api_key FROM clinics WHERE id = $1`, clinicID)
	return api, err
}

// get sends a GET request with access_token and any extra params as query string.
func (a *MacDentAdapter) get(ctx context.Context, apiKey, path string, params url.Values) ([]byte, error) {
	if params == nil {
		params = url.Values{}
	}
	params.Set("access_token", apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		macdentBase+path+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

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
	ID           int           `json:"id"`
	Name         string        `json:"name"`
	Specialnosti []mdSpecialty `json:"specialnosti"`
}

type doctorsResponse struct {
	Doctors []mdDoctor `json:"doctors"`
}

func (a *MacDentAdapter) ListDoctors(ctx context.Context, clinicID uuid.UUID) ([]Doctor, error) {
	apiKey, err := a.GetClinicAPI(ctx, clinicID)
	if err != nil {
		return nil, fmt.Errorf("macdent: get api key: %w", err)
	}
	mds, err := a.listDoctors(ctx, apiKey)
	if err != nil {
		return nil, err
	}
	out := make([]Doctor, 0, len(mds))
	for _, d := range mds {
		spec := make([]string, 0, len(d.Specialnosti))
		for _, s := range d.Specialnosti {
			spec = append(spec, s.Name)
		}
		out = append(out, Doctor{
			ID:          fmt.Sprint(d.ID),
			Name:        d.Name,
			Specialties: spec,
		})
	}
	return out, nil
}

func (a *MacDentAdapter) listDoctors(ctx context.Context, apiKey string) ([]mdDoctor, error) {
	b, err := a.get(ctx, apiKey, "/doctor/find", nil)
	if err != nil {
		return nil, err
	}

	var resp doctorsResponse
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, fmt.Errorf("macdent doctor/find: %w", err)
	}

	return resp.Doctors, nil
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
	ctx context.Context, clinicID uuid.UUID, from, to time.Time, specialty string,
) ([]Slot, error) {
	apiKey, err := a.GetClinicAPI(ctx, clinicID)
	if err != nil {
		return nil, fmt.Errorf("macdent: get api key: %w", err)
	}

	doctors, err := a.listDoctors(ctx, apiKey)
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
		b, err := a.get(ctx, apiKey, "/doctor/get_free_time", fields)
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
	apiKey, err := a.GetClinicAPI(ctx, req.ClinicID)
	if err != nil {
		return nil, fmt.Errorf("macdent: get api key: %w", err)
	}

	// MacDent requires an integer doctor ID stored in doctors.external_id.
	fields := url.Values{
		"start":   {req.StartsAt.Format(macdentDateLayout)},
		"end":     {req.EndsAt.Format(macdentDateLayout)},
		"doctor":  {req.DoctorID.String()},
		"zhaloba": {req.Service},
	}

	b, err := a.get(ctx, apiKey, "/zapis/add", fields)
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
