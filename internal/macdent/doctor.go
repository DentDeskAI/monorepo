package macdent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

type Specialty struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Doctor struct {
	ID           int         `json:"id"`
	Name         string      `json:"name"`
	Specialnosti []Specialty `json:"specialnosti"`
}

type Schedule struct {
	From string `json:"from"` // "DD.MM.YYYY HH:MM:SS"
	To   string `json:"to"`
}

func (c *Client) ListDoctors(ctx context.Context) ([]Doctor, error) {
	b, err := c.Get(ctx, "/doctor/find", nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Doctors []Doctor `json:"doctors"`
	}
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, fmt.Errorf("macdent doctor/find: %w", err)
	}
	return resp.Doctors, nil
}

func (c *Client) GetFreeTime(ctx context.Context, doctorID int, from, to time.Time) ([]Schedule, error) {
	params := url.Values{
		"doctor":   {fmt.Sprint(doctorID)},
		"dateFrom": {from.Format("2006-01-02")},
		"dateTo":   {to.Format("2006-01-02")},
	}
	b, err := c.Get(ctx, "/doctor/get_free_time", params)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Schedules []Schedule `json:"schedules"`
	}
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, fmt.Errorf("macdent doctor/get_free_time: %w", err)
	}
	return resp.Schedules, nil
}
