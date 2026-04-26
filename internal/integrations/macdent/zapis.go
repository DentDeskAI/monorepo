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

// AddZapis creates a new patient appointment (zapis/add).
// doctorID is the MacDent integer doctor ID (stored as external_id in the local doctors table).
func (c *Client) AddZapis(ctx context.Context, doctorID int, startsAt, endsAt time.Time, complaint string) (*Zapis, error) {
	params := url.Values{
		"doctor":   {fmt.Sprint(doctorID)},
		"start":    {startsAt.Format(DateLayout)},
		"end":      {endsAt.Format(DateLayout)},
		"zhaloba":  {complaint},
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
