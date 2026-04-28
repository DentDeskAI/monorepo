package macdent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

// DateOnlyLayout is the date-only format MacDent accepts for range filters.
const DateOnlyLayout = "02.01.2006"

// Appointment is a flat record returned by zapis/find.
type Appointment struct {
	ID      int    `json:"id"`
	Doctor  int    `json:"doctor"`
	Patient int    `json:"patient"`
	Date    string `json:"date"`
	Start   string `json:"start"`
	End     string `json:"end"`
	Status  int    `json:"status"`
	Zhaloba string `json:"zhaloba"`
	Comment string `json:"comment"`
	IsFirst bool   `json:"isFirst"`
	Cabinet string `json:"cabinet"`
	Rasp    string `json:"rasp"`
}

type AppointmentsResponse struct {
	Appointments []Appointment `json:"zapisi"`
	Count        string        `json:"count"`
	AtPage       int           `json:"atPage"`
	MaxPage      int           `json:"maxPage"`
	Response     int           `json:"response"`
}

// ZapisDetail is a rich record returned by zapis/get with embedded doctor/patient objects.
type ZapisDetail struct {
	ID      int             `json:"id"`
	Doctor  ZapisDoctorRef  `json:"doctor"`
	Patient ZapisPatientRef `json:"patient"`
	Date    string          `json:"date"`
	Start   string          `json:"start"`
	End     string          `json:"end"`
	Status  int             `json:"status"`
	Zhaloba string          `json:"zhaloba"`
	Comment string          `json:"comment"`
	IsFirst bool            `json:"isFirst"`
	Cabinet string          `json:"cabinet"`
	Rasp    string          `json:"rasp"`
}

type ZapisDoctorRef struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ZapisPatientRef struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Phone *string `json:"phone"`
}

// UpdateZapisParams holds optional fields for zapis/update.
// Only non-nil fields are sent to MacDent.
type UpdateZapisParams struct {
	DoctorID *int
	Start    *time.Time
	End      *time.Time
	Zhaloba  *string
	Comment  *string
}

// GetAppointments fetches appointments for the given date range from zapis/find.
func (c *Client) GetAppointments(ctx context.Context, from, to time.Time) (*AppointmentsResponse, error) {
	params := url.Values{
		"dateFrom": {from.Format(DateOnlyLayout)},
		"dateTo":   {to.Format(DateOnlyLayout)},
	}
	b, err := c.Get(ctx, "/zapis/find", params)
	if err != nil {
		return nil, err
	}
	var resp AppointmentsResponse
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, fmt.Errorf("macdent zapis/find: %w", err)
	}
	return &resp, nil
}

// GetAppointmentByID fetches a single appointment with embedded doctor and patient objects.
func (c *Client) GetAppointmentByID(ctx context.Context, id int) (*ZapisDetail, error) {
	b, err := c.Get(ctx, "/zapis/get", url.Values{"id": {fmt.Sprint(id)}})
	if err != nil {
		return nil, err
	}
	var resp struct {
		Zapis    ZapisDetail `json:"zapis"`
		Response int         `json:"response"`
	}
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, fmt.Errorf("macdent zapis/get: %w", err)
	}
	return &resp.Zapis, nil
}

// UpdateAppointment calls zapis/update with only the non-nil fields.
func (c *Client) UpdateAppointment(ctx context.Context, id int, p UpdateZapisParams) error {
	params := url.Values{"id": {fmt.Sprint(id)}}
	if p.DoctorID != nil {
		params.Set("doctor", fmt.Sprint(*p.DoctorID))
	}
	if p.Start != nil {
		params.Set("start", p.Start.Format(DateLayout))
	}
	if p.End != nil {
		params.Set("end", p.End.Format(DateLayout))
	}
	if p.Zhaloba != nil {
		params.Set("zhaloba", *p.Zhaloba)
	}
	if p.Comment != nil {
		params.Set("comment", *p.Comment)
	}
	_, err := c.Get(ctx, "/zapis/update", params)
	return err
}

// RemoveAppointment calls zapis/remove for the given record ID.
func (c *Client) RemoveAppointment(ctx context.Context, id int) error {
	_, err := c.Get(ctx, "/zapis/remove", url.Values{"id": {fmt.Sprint(id)}})
	return err
}
