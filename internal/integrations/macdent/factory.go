package macdent

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// ClientFor fetches the clinic's MacDent API key from the local DB and returns
// a ready-to-use *Client that shares the given httpClient (connection pool reuse).
//
// This is the single entry point for code that needs to talk to MacDent on
// behalf of a tenant clinic.
func ClientFor(ctx context.Context, db *sqlx.DB, httpClient *http.Client, clinicID uuid.UUID) (*Client, error) {
	var apiKey string
	if err := db.GetContext(ctx, &apiKey,
		`SELECT macdent_api_key FROM clinics WHERE id = $1`, clinicID); err != nil {
		return nil, fmt.Errorf("macdent: get api key for clinic %s: %w", clinicID, err)
	}
	if apiKey == "" {
		return nil, fmt.Errorf("macdent: no api key set for clinic %s", clinicID)
	}
	return NewWithHTTP(apiKey, httpClient), nil
}
