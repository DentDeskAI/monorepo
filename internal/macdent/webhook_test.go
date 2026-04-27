package macdent

import "testing"

func TestSummarizeWebhookFromTopLevelFields(t *testing.T) {
	body := []byte(`{
		"event":"onChange",
		"object":"patient",
		"id":12345
	}`)

	got := SummarizeWebhook(body)
	if got.Event != "onChange" {
		t.Fatalf("event = %q, want %q", got.Event, "onChange")
	}
	if got.Entity != "patient" {
		t.Fatalf("entity = %q, want %q", got.Entity, "patient")
	}
	if got.ObjectID != "12345" {
		t.Fatalf("object id = %q, want %q", got.ObjectID, "12345")
	}
}

func TestSummarizeWebhookFallsBackToNestedData(t *testing.T) {
	body := []byte(`{
		"action":"onCreate",
		"data":{
			"entity":"zapis",
			"external_id":"appt-42"
		}
	}`)

	got := SummarizeWebhook(body)
	if got.Event != "onCreate" {
		t.Fatalf("event = %q, want %q", got.Event, "onCreate")
	}
	if got.Entity != "zapis" {
		t.Fatalf("entity = %q, want %q", got.Entity, "zapis")
	}
	if got.ObjectID != "appt-42" {
		t.Fatalf("object id = %q, want %q", got.ObjectID, "appt-42")
	}
}
