package whatsapp

import (
	"time"
)

// WebhookPayload описывает входящий webhook от Meta.
// Справочник: https://developers.facebook.com/docs/graph-api/webhooks/reference/whatsapp-business-account
type WebhookPayload struct {
	Object string  `json:"object"`
	Entry  []Entry `json:"entry"`
}

type Entry struct {
	ID      string   `json:"id"`
	Changes []Change `json:"changes"`
}

type Change struct {
	Field string `json:"field"`
	Value Value  `json:"value"`
}

type Value struct {
	MessagingProduct string        `json:"messaging_product"`
	Metadata         Metadata      `json:"metadata"`
	Contacts         []Contact     `json:"contacts,omitempty"`
	Messages         []InMessage   `json:"messages,omitempty"`
	Statuses         []interface{} `json:"statuses,omitempty"`
}

type Metadata struct {
	DisplayPhoneNumber string `json:"display_phone_number"`
	PhoneNumberID      string `json:"phone_number_id"`
}

type Contact struct {
	Profile struct {
		Name string `json:"name"`
	} `json:"profile"`
	WAID string `json:"wa_id"`
}

type InMessage struct {
	From      string    `json:"from"`
	ID        string    `json:"id"`
	Timestamp string    `json:"timestamp"`
	Type      string    `json:"type"`
	Text      *TextBody `json:"text,omitempty"`
}

type TextBody struct {
	Body string `json:"body"`
}

// Extracted — нормализованное сообщение, с которым работает наш код.
type Extracted struct {
	PhoneNumberID string
	From          string // "77011234567" (без +)
	ProfileName   string
	MessageID     string
	Text          string
	Timestamp     time.Time
}

// Extract — безопасно достать все пригодные для обработки сообщения из payload'а.
// Фильтрует служебные updates (статусы доставки и т.п.).
func (p *WebhookPayload) Extract() []Extracted {
	var out []Extracted
	for _, e := range p.Entry {
		for _, ch := range e.Changes {
			phoneID := ch.Value.Metadata.PhoneNumberID
			// Карта wa_id → profile.name
			nameByID := map[string]string{}
			for _, c := range ch.Value.Contacts {
				nameByID[c.WAID] = c.Profile.Name
			}
			for _, m := range ch.Value.Messages {
				if m.Type != "text" || m.Text == nil {
					continue
				}
				ts := time.Now()
				// Meta шлёт unix-секунды как строку — разбираем best-effort.
				if n, err := parseUnixSeconds(m.Timestamp); err == nil {
					ts = n
				}
				out = append(out, Extracted{
					PhoneNumberID: phoneID,
					From:          m.From,
					ProfileName:   nameByID[m.From],
					MessageID:     m.ID,
					Text:          m.Text.Body,
					Timestamp:     ts,
				})
			}
		}
	}
	return out
}

func parseUnixSeconds(s string) (time.Time, error) {
	var sec int64
	for _, c := range s {
		if c < '0' || c > '9' {
			return time.Time{}, errInvalidTs
		}
		sec = sec*10 + int64(c-'0')
	}
	return time.Unix(sec, 0), nil
}

var errInvalidTs = &tsErr{}

type tsErr struct{}

func (*tsErr) Error() string { return "bad timestamp" }
