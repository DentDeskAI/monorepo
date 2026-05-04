package whatsapp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookPayloadExtract(t *testing.T) {
	payload := WebhookPayload{
		Object: "whatsapp_business_account",
		Entry: []Entry{
			{
				ID: "entry-1",
				Changes: []Change{
					{
						Field: "messages",
						Value: Value{
							Metadata: Metadata{PhoneNumberID: "phone-123"},
							Contacts: []Contact{
								{WAID: "77001234567", Profile: struct {
									Name string `json:"name"`
								}{Name: "Aigerim"}},
							},
							Messages: []InMessage{
								{
									From:      "77001234567",
									ID:        "msg-1",
									Timestamp: "1714471200",
									Type:      "text",
									Text:      &TextBody{Body: "Hello"},
								},
								{
									From:      "77001234567",
									ID:        "msg-2",
									Timestamp: "1714471201",
									Type:      "image",
								},
								{
									From:      "77001234567",
									ID:        "msg-3",
									Timestamp: "bad-ts",
									Type:      "text",
									Text:      &TextBody{Body: "Fallback time"},
								},
							},
						},
					},
				},
			},
		},
	}

	before := time.Now().Add(-time.Second)
	got := payload.Extract()
	after := time.Now().Add(time.Second)

	require.Len(t, got, 2)

	assert.Equal(t, "phone-123", got[0].PhoneNumberID)
	assert.Equal(t, "77001234567", got[0].From)
	assert.Equal(t, "Aigerim", got[0].ProfileName)
	assert.Equal(t, "msg-1", got[0].MessageID)
	assert.Equal(t, "Hello", got[0].Text)
	assert.Equal(t, time.Unix(1714471200, 0), got[0].Timestamp)

	assert.Equal(t, "msg-3", got[1].MessageID)
	assert.Equal(t, "Fallback time", got[1].Text)
	assert.True(t, got[1].Timestamp.After(before))
	assert.True(t, got[1].Timestamp.Before(after))
}

func TestParseUnixSeconds(t *testing.T) {
	ts, err := parseUnixSeconds("1714471200")
	require.NoError(t, err)
	assert.Equal(t, time.Unix(1714471200, 0), ts)

	_, err = parseUnixSeconds("17x")
	assert.EqualError(t, err, "bad timestamp")
}
