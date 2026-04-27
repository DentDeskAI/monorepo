package macdent

import (
	"context"
	"encoding/json"
	"fmt"
)

type ProfileResponse struct {
	Stomatology Stomatology `json:"stomatology"`
	Response    int         `json:"response"`
}

type Stomatology struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (c *Client) GetClinic(ctx context.Context) (*Stomatology, error) {
	b, err := c.Get(ctx, "/profile/get", nil)
	if err != nil {
		return nil, err
	}

	var resp ProfileResponse
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, fmt.Errorf("macdent profile/get: %w", err)
	}

	if resp.Response != 1 {
		return nil, fmt.Errorf("macdent profile/get: expected 1 response, got %d", resp.Response)
	}

	return &resp.Stomatology, nil
}
