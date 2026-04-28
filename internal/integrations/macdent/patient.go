package macdent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

type Patient struct {
	Name      string  `json:"name"`
	Gender    *string `json:"gender"` // nullable
	ID        int     `json:"id"`
	IIN       *string `json:"iin"` // nullable
	Number    string  `json:"number"`
	Phone     *string `json:"phone"` // nullable
	Birth     *string `json:"birth"` // nullable (could be time.Time if formatted)
	IsChild   bool    `json:"isChild"`
	Comment   string  `json:"comment"`
	WhereKnow string  `json:"whereKnow"`
}

type ListPatientsResponse struct {
	Patients []Patient `json:"patients"`
	Count    string    `json:"count"`
	AtPage   int       `json:"atPage"`
	MaxPage  int       `json:"maxPage"`
	Response int       `json:"response"`
}

func (c *Client) GetPatientByID(ctx context.Context, id int) (*Patient, error) {
	b, err := c.Get(ctx, "/patient/get", url.Values{"id": {fmt.Sprint(id)}})
	if err != nil {
		return nil, err
	}
	var resp struct {
		Patient  Patient `json:"patient"`
		Response int     `json:"response"`
	}
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, fmt.Errorf("macdent patient/get: %w", err)
	}
	return &resp.Patient, nil
}

func (c *Client) ListPatients(ctx context.Context) (*ListPatientsResponse, error) {
	b, err := c.Get(ctx, "/patient/find", nil)
	if err != nil {
		return nil, err
	}

	var resp ListPatientsResponse
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, fmt.Errorf("macdent patient/find: %w", err)
	}

	return &resp, nil
}
