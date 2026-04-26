package macdent

import (
	"encoding/json"
	"fmt"
	"strings"
)

type WebhookSummary struct {
	Event    string
	Entity   string
	ObjectID string
}

// SummarizeWebhook extracts a small, stable subset of fields from a MacDent
// webhook payload so handlers can log useful context even if the exact payload
// schema changes between event types.
func SummarizeWebhook(body []byte) WebhookSummary {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return WebhookSummary{}
	}

	summary := WebhookSummary{
		Event:    pickString(payload, "event", "action", "method"),
		Entity:   pickString(payload, "object", "entity", "model", "type"),
		ObjectID: pickString(payload, "id", "externalId", "external_id"),
	}

	data, _ := payload["data"].(map[string]any)
	if data == nil {
		return summary
	}

	if summary.Entity == "" {
		summary.Entity = pickString(data, "object", "entity", "model", "type")
	}
	if summary.ObjectID == "" {
		summary.ObjectID = pickString(data, "id", "externalId", "external_id")
	}

	return summary
}

func pickString(src map[string]any, keys ...string) string {
	for _, key := range keys {
		if val, ok := src[key]; ok {
			if s := stringify(val); s != "" {
				return s
			}
		}
	}
	return ""
}

func stringify(v any) string {
	if v == nil {
		return ""
	}
	s := strings.TrimSpace(fmt.Sprint(v))
	if s == "<nil>" {
		return ""
	}
	return s
}
