// Package whatsapp provides types and a client for the WhatsApp Cloud API.
package whatsapp

// ─── Inbound webhook payload ─────────────────────────────────────────────────

// WebhookPayload is the top-level object sent by Meta to our webhook endpoint.
type WebhookPayload struct {
	Object string  `json:"object"`
	Entry  []Entry `json:"entry"`
}

type Entry struct {
	ID      string   `json:"id"`
	Changes []Change `json:"changes"`
}

type Change struct {
	Value ChangeValue `json:"value"`
	Field string      `json:"field"`
}

type ChangeValue struct {
	MessagingProduct string          `json:"messaging_product"`
	Metadata         Metadata        `json:"metadata"`
	Contacts         []Contact       `json:"contacts"`
	Messages         []InboundMessage `json:"messages"`
	Statuses         []StatusUpdate  `json:"statuses"`
}

type Metadata struct {
	DisplayPhoneNumber string `json:"display_phone_number"`
	PhoneNumberID      string `json:"phone_number_id"`
}

type Contact struct {
	Profile Profile `json:"profile"`
	WaID    string  `json:"wa_id"`
}

type Profile struct {
	Name string `json:"name"`
}

// InboundMessage represents a message received from a WhatsApp user.
type InboundMessage struct {
	From      string          `json:"from"`      // sender's phone number (E.164 without +)
	ID        string          `json:"id"`        // wamid — stable Meta message ID
	Timestamp string          `json:"timestamp"` // Unix epoch as string
	Type      string          `json:"type"`      // text | image | audio | document | ...
	Text      *TextContent    `json:"text,omitempty"`
	Image     *MediaContent   `json:"image,omitempty"`
	Audio     *MediaContent   `json:"audio,omitempty"`
	Document  *MediaContent   `json:"document,omitempty"`
	Context   *MessageContext `json:"context,omitempty"` // present if reply to another message
}

type TextContent struct {
	Body string `json:"body"`
}

type MediaContent struct {
	ID       string `json:"id"`
	MimeType string `json:"mime_type"`
	SHA256   string `json:"sha256"`
	Caption  string `json:"caption,omitempty"`
}

type MessageContext struct {
	From string `json:"from"`
	ID   string `json:"id"`
}

// StatusUpdate is sent by Meta when delivery state changes for outbound messages.
type StatusUpdate struct {
	ID           string    `json:"id"`
	Status       string    `json:"status"` // sent | delivered | read | failed
	Timestamp    string    `json:"timestamp"`
	RecipientID  string    `json:"recipient_id"`
	Conversation *Conversation `json:"conversation,omitempty"`
}

type Conversation struct {
	ID     string `json:"id"`
	Origin *Origin `json:"origin,omitempty"`
}

type Origin struct {
	Type string `json:"type"` // business_initiated | user_initiated
}

// ─── Outbound message requests ───────────────────────────────────────────────

// SendTextRequest is the payload for sending a plain-text WhatsApp message.
type SendTextRequest struct {
	MessagingProduct string      `json:"messaging_product"` // always "whatsapp"
	RecipientType    string      `json:"recipient_type"`    // always "individual"
	To               string      `json:"to"`
	Type             string      `json:"type"` // "text"
	Text             TextContent `json:"text"`
}

// SendMessageResponse is returned by the Cloud API after a successful send.
type SendMessageResponse struct {
	MessagingProduct string              `json:"messaging_product"`
	Contacts         []ResponseContact   `json:"contacts"`
	Messages         []ResponseMessage   `json:"messages"`
}

type ResponseContact struct {
	Input string `json:"input"`
	WaID  string `json:"wa_id"`
}

type ResponseMessage struct {
	ID string `json:"id"` // wamid of the sent message
}
