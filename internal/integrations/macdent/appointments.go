package macdent

import (
	"context"
	"encoding/json"
	"fmt"
)

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

func (c *Client) GetAppointments(ctx context.Context) (*AppointmentsResponse, error) {
	b, err := c.Get(ctx, "/zapis/find", nil)
	if err != nil {
		return nil, err
	}

	var resp AppointmentsResponse
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, fmt.Errorf("macdent zapis/find: %w", err)
	}

	return &resp, nil
}
