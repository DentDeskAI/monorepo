package macdent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

type AppointmentRequest struct {
	PatientName  string
	PatientPhone string
	Start        time.Time
	End          time.Time
	WhereKnow    string // optional — "откуда узнал"
}

type AppointmentRequestResult struct {
	ID int `json:"id"`
}

// SendAppointmentRequest creates a заявка via appointment/send.
// Unlike zapis/add this does not require an existing patient record.
func (c *Client) SendAppointmentRequest(ctx context.Context, req AppointmentRequest) (*AppointmentRequestResult, error) {
	params := url.Values{
		"patient_name":  {req.PatientName},
		"patient_phone": {req.PatientPhone},
		"start":         {req.Start.Format(DateLayout)},
		"end":           {req.End.Format(DateLayout)},
	}
	if req.WhereKnow != "" {
		params.Set("where_know", req.WhereKnow)
	}
	b, err := c.Get(ctx, "/appointment/send", params)
	if err != nil {
		return nil, err
	}
	var resp struct {
		ID       int `json:"id"`
		Response int `json:"response"`
	}
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, fmt.Errorf("macdent appointment/send: %w", err)
	}
	return &AppointmentRequestResult{ID: resp.ID}, nil
}
