package whatsapp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
)

// Client wraps the WhatsApp Cloud API.
type Client struct {
	httpClient    *resty.Client
	phoneNumberID string
}

// NewClient creates a ready-to-use WhatsApp Cloud API client.
func NewClient(apiURL, phoneNumberID, accessToken string) *Client {
	rc := resty.New().
		SetBaseURL(apiURL).
		SetAuthToken(accessToken).
		SetHeader("Content-Type", "application/json")

	return &Client{
		httpClient:    rc,
		phoneNumberID: phoneNumberID,
	}
}

// SendText sends a plain-text message to a phone number.
// to must be in E.164 format without the leading '+' (e.g. "77001234567").
func (c *Client) SendText(ctx context.Context, to, body string) (*SendMessageResponse, error) {
	endpoint := fmt.Sprintf("/%s/messages", c.phoneNumberID)

	payload := SendTextRequest{
		MessagingProduct: "whatsapp",
		RecipientType:    "individual",
		To:               to,
		Type:             "text",
		Text:             TextContent{Body: body},
	}

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetBody(payload).
		Post(endpoint)

	if err != nil {
		return nil, fmt.Errorf("whatsapp send request: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		log.Error().
			Int("status", resp.StatusCode()).
			Str("body", resp.String()).
			Msg("whatsapp API returned non-200")
		return nil, fmt.Errorf("whatsapp API error %d: %s", resp.StatusCode(), resp.String())
	}

	var result SendMessageResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("whatsapp decode response: %w", err)
	}

	log.Info().
		Str("to", to).
		Str("wamid", firstMessageID(result)).
		Msg("WhatsApp message sent")

	return &result, nil
}

// MarkAsRead marks an inbound message as read (shows blue ticks to sender).
func (c *Client) MarkAsRead(ctx context.Context, messageID string) error {
	endpoint := fmt.Sprintf("/%s/messages", c.phoneNumberID)

	payload := map[string]any{
		"messaging_product": "whatsapp",
		"status":            "read",
		"message_id":        messageID,
	}

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetBody(payload).
		Put(endpoint)

	if err != nil {
		return fmt.Errorf("whatsapp mark-as-read: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("mark-as-read failed %d: %s", resp.StatusCode(), resp.String())
	}

	return nil
}

func firstMessageID(r SendMessageResponse) string {
	if len(r.Messages) > 0 {
		return r.Messages[0].ID
	}
	return ""
}
