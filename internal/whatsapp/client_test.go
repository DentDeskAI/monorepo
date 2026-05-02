package whatsapp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestNewClientDefaultsAPIVersion(t *testing.T) {
	c := NewClient("token", "phone-id", "")
	assert.Equal(t, "v20.0", c.apiVersion)
}

func TestSendText_NoCredentialsIsNoOp(t *testing.T) {
	c := NewClient("", "", "")
	err := c.SendText(context.Background(), "77001234567", "hello")
	assert.NoError(t, err)
}

func TestSendText_SendsExpectedRequest(t *testing.T) {
	c := NewClient("token-123", "phone-456", "v21.0")
	c.http = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, http.MethodPost, req.Method)
			assert.Equal(t, "https://graph.facebook.com/v21.0/phone-456/messages", req.URL.String())
			assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
			assert.Equal(t, "Bearer token-123", req.Header.Get("Authorization"))

			body, err := io.ReadAll(req.Body)
			require.NoError(t, err)

			var payload map[string]any
			require.NoError(t, json.Unmarshal(body, &payload))
			assert.Equal(t, "whatsapp", payload["messaging_product"])
			assert.Equal(t, "77001234567", payload["to"])
			assert.Equal(t, "text", payload["type"])

			text, ok := payload["text"].(map[string]any)
			require.True(t, ok)
			assert.Equal(t, "hello there", text["body"])
			assert.Equal(t, false, text["preview_url"])

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"messages":[{"id":"1"}]}`)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	err := c.SendText(context.Background(), "77001234567", "hello there")
	assert.NoError(t, err)
}

func TestSendText_TransportError(t *testing.T) {
	c := NewClient("token", "phone", "")
	c.http = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("network down")
		}),
	}

	err := c.SendText(context.Background(), "77001234567", "hello")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "whatsapp send")
}

func TestSendText_HTTPError(t *testing.T) {
	c := NewClient("token", "phone", "")
	c.http = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(strings.NewReader(`bad request`)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	err := c.SendText(context.Background(), "77001234567", "hello")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "whatsapp status 400: bad request")
}
