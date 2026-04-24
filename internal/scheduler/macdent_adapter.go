package scheduler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// MacDentAdapter — реальная интеграция с MacDent API.
// Конкретные endpoints уточняются у вендора; здесь — контракт и HTTP-плитка.
type MacDentAdapter struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

func NewMacDentAdapter(baseURL, apiKey string) *MacDentAdapter {
	return &MacDentAdapter{
		baseURL: baseURL,
		apiKey:  apiKey,
		http:    &http.Client{Timeout: 15 * time.Second},
	}
}

type mdSlotResp struct {
	DoctorID   string    `json:"doctor_id"`
	DoctorName string    `json:"doctor_name"`
	StartsAt   time.Time `json:"starts_at"`
	EndsAt     time.Time `json:"ends_at"`
}

func (a *MacDentAdapter) GetFreeSlots(ctx context.Context, clinicID uuid.UUID, from, to time.Time, specialty string) ([]Slot, error) {
	url := fmt.Sprintf("%s/api/v1/slots?from=%s&to=%s&specialty=%s",
		a.baseURL,
		from.UTC().Format(time.RFC3339),
		to.UTC().Format(time.RFC3339),
		specialty,
	)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+a.apiKey)

	resp, err := a.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("macdent: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("macdent status %d: %s", resp.StatusCode, string(b))
	}
	var items []mdSlotResp
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, err
	}
	out := make([]Slot, 0, len(items))
	for _, s := range items {
		// external_id маппим в наши doctors по external_id (делается на уровне выше, если нужно).
		// Для простоты MVP возвращаем nil uuid — вызывающий код должен сопоставить.
		out = append(out, Slot{
			StartsAt: s.StartsAt,
			EndsAt:   s.EndsAt,
			DoctorID: uuid.Nil,
			Doctor:   s.DoctorName,
		})
	}
	return out, nil
}

func (a *MacDentAdapter) CreateAppointment(ctx context.Context, req BookRequest) (*BookResult, error) {
	body, _ := json.Marshal(map[string]any{
		"patient_id": req.PatientID,
		"doctor_id":  req.DoctorID,
		"starts_at":  req.StartsAt.UTC().Format(time.RFC3339),
		"ends_at":    req.EndsAt.UTC().Format(time.RFC3339),
		"service":    req.Service,
	})
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/api/v1/appointments", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+a.apiKey)

	resp, err := a.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("macdent: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("macdent status %d: %s", resp.StatusCode, string(b))
	}
	var out struct {
		ID         string `json:"id"`
		ExternalID string `json:"external_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	id, err := uuid.Parse(out.ID)
	if err != nil {
		return nil, errors.New("macdent returned bad uuid")
	}
	ext := out.ExternalID
	return &BookResult{AppointmentID: id, ExternalID: &ext}, nil
}

func (a *MacDentAdapter) CancelAppointment(ctx context.Context, appointmentID uuid.UUID) error {
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodDelete,
		a.baseURL+"/api/v1/appointments/"+appointmentID.String(), nil)
	httpReq.Header.Set("Authorization", "Bearer "+a.apiKey)
	resp, err := a.http.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("macdent status %d: %s", resp.StatusCode, string(b))
	}
	return nil
}
