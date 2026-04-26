// Package macdent is a typed Go client for the MacDent external API.
// Each API group (doctor, patient, zapis, …) lives in its own file.
// All methods are on *Client so new groups are added without touching existing files.
package macdent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	BaseURL    = "https://api-developer.macdent.kz"
	DateLayout = "02.01.2006 15:04:05"
)

type Client struct {
	apiKey string
	http   *http.Client
}

func New(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		http:   &http.Client{Timeout: 15 * time.Second},
	}
}

// NewWithHTTP lets callers share a single http.Client (connection pool reuse).
func NewWithHTTP(apiKey string, httpClient *http.Client) *Client {
	return &Client{apiKey: apiKey, http: httpClient}
}

// Get sends a GET request; access_token and params go in the query string.
func (c *Client) Get(ctx context.Context, path string, params url.Values) ([]byte, error) {
	if params == nil {
		params = url.Values{}
	}
	params.Set("access_token", c.apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		BaseURL+path+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("macdent %s: %w", path, err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)

	var env struct {
		Response int    `json:"response"`
		Error    string `json:"error"`
	}
	if err := json.Unmarshal(b, &env); err != nil {
		return nil, fmt.Errorf("macdent %s: bad json: %w", path, err)
	}
	if env.Response == 0 {
		return nil, fmt.Errorf("macdent %s: %s", path, env.Error)
	}
	return b, nil
}
