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

type AddPatientParams struct {
	Name      string // required
	Phone     string
	IIN       string
	Birth     string // dd.mm.yyyy
	Gender    string // "M" or "F"
	Comment   string
	WhereKnow string
	IsChild   bool
}

func (c *Client) AddPatient(ctx context.Context, p AddPatientParams) (*Patient, error) {
	params := url.Values{"name": {p.Name}}
	if p.Phone != "" {
		params.Set("phone", p.Phone)
	}
	if p.IIN != "" {
		params.Set("iin", p.IIN)
	}
	if p.Birth != "" {
		params.Set("birth", p.Birth)
	}
	if p.Gender != "" {
		params.Set("gender", p.Gender)
	}
	if p.Comment != "" {
		params.Set("comment", p.Comment)
	}
	if p.WhereKnow != "" {
		params.Set("whereKnow", p.WhereKnow)
	}
	if p.IsChild {
		params.Set("isChild", "1")
	}

	b, err := c.Get(ctx, "/patient/add", params)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Patient  Patient `json:"patient"`
		ID       int     `json:"id"`
		Response int     `json:"response"`
	}
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, fmt.Errorf("macdent patient/add: %w", err)
	}
	if resp.Patient.ID == 0 && resp.ID != 0 {
		resp.Patient.ID = resp.ID
		resp.Patient.Name = p.Name
		if p.Phone != "" {
			resp.Patient.Phone = &p.Phone
		}
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
