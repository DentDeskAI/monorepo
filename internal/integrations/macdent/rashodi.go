package macdent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

// Rashod is a single monetary transaction returned by rashodi/find.
// type: 1 = income, 2 = expense.
type Rashod struct {
	ID         int    `json:"id"`
	Date       string `json:"date"`
	Name       string `json:"name"`
	Summ       any    `json:"summ"` // MacDent returns string or number depending on version
	Type       int    `json:"type"`
	TypeOplata string `json:"typeOplata"`
	Filial     any    `json:"filial"`
	Comment    string `json:"comment"`
}

// SummFloat converts the Summ field (which may be a string or float) to float64.
func (r Rashod) SummFloat() float64 {
	switch v := r.Summ.(type) {
	case float64:
		return v
	case string:
		var f float64
		fmt.Sscanf(v, "%f", &f)
		return f
	}
	return 0
}

type RashodiResponse struct {
	Rashodi  []Rashod `json:"rashodi"`
	Count    string   `json:"count"`
	Response int      `json:"response"`
}

// GetRashodi fetches monetary transactions for the given date range from rashodi/find.
func (c *Client) GetRashodi(ctx context.Context, from, to time.Time) (*RashodiResponse, error) {
	params := url.Values{
		"dateFrom": {from.Format(DateOnlyLayout)},
		"dateTo":   {to.Format(DateOnlyLayout)},
	}
	b, err := c.Get(ctx, "/rashodi/find", params)
	if err != nil {
		return nil, err
	}
	var resp RashodiResponse
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, fmt.Errorf("macdent rashodi/find: %w", err)
	}
	return &resp, nil
}
