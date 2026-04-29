package macdent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

type Zapis struct {
	ID int `json:"id"`
}

// AddZapisParams holds the inputs for zapis/add.
type AddZapisParams struct {
	DoctorID  int
	PatientID int
	Start     time.Time
	End       time.Time
	Zhaloba   string
	Cabinet   string
	IsFirst   bool
}

// AddZapis creates a new patient appointment (zapis/add).
// DoctorID and PatientID are MacDent integer IDs.
func (c *Client) AddZapis(ctx context.Context, p AddZapisParams) (*Zapis, error) {
	params := url.Values{
		"doctor":  {fmt.Sprint(p.DoctorID)},
		"patient": {fmt.Sprint(p.PatientID)},
		"start":   {p.Start.Format(DateLayout)},
		"end":     {p.End.Format(DateLayout)},
	}
	if p.Zhaloba != "" {
		params.Set("zhaloba", p.Zhaloba)
	}
	if p.Cabinet != "" {
		params.Set("cabinet", p.Cabinet)
	}
	if p.IsFirst {
		params.Set("isFirst", "1")
	}
	b, err := c.Get(ctx, "/zapis/add", params)
	if err != nil {
		return nil, err
	}
	var resp struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, fmt.Errorf("macdent zapis/add: %w", err)
	}
	return &Zapis{ID: resp.ID}, nil
}

// SetStatus updates the status of a zapis.
// status values: 0=DEFAULT, 1=CONFIRM, 2=DECLINED, 3=CAME, 4=LEFT, 5=IN_PROCESS, 6=LATE
func (c *Client) SetStatus(ctx context.Context, zapisID, status int) error {
	params := url.Values{
		"id":     {fmt.Sprint(zapisID)},
		"status": {fmt.Sprint(status)},
	}
	_, err := c.Get(ctx, "/zapis/set_status", params)
	return err
}
