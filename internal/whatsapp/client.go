// Package whatsapp — минимальный клиент WhatsApp Cloud API (Meta).
// Отправка текстовых сообщений. Интерактивные кнопки намеренно опущены в MVP.
package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	token         string
	phoneNumberID string
	apiVersion    string
	http          *http.Client
}

func NewClient(token, phoneNumberID, apiVersion string) *Client {
	if apiVersion == "" {
		apiVersion = "v20.0"
	}
	return &Client{
		token:         token,
		phoneNumberID: phoneNumberID,
		apiVersion:    apiVersion,
		http:          &http.Client{Timeout: 10 * time.Second},
	}
}

type sendTextBody struct {
	MessagingProduct string `json:"messaging_product"`
	To               string `json:"to"`
	Type             string `json:"type"`
	Text             struct {
		PreviewURL bool   `json:"preview_url"`
		Body       string `json:"body"`
	} `json:"text"`
}

func (c *Client) SendText(ctx context.Context, to, body string) error {
	if c.token == "" || c.phoneNumberID == "" {
		// dev/demo режим — просто лог, без попыток сети.
		return nil
	}
	payload := sendTextBody{
		MessagingProduct: "whatsapp",
		To:               to,
		Type:             "text",
	}
	payload.Text.Body = body

	buf, _ := json.Marshal(payload)
	url := fmt.Sprintf("https://graph.facebook.com/%s/%s/messages", c.apiVersion, c.phoneNumberID)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(buf))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("whatsapp send: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("whatsapp status %d: %s", resp.StatusCode, string(b))
	}
	return nil
}
