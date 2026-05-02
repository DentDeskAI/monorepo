package macdent

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestClientGet_Success(t *testing.T) {
	client := NewWithHTTP("api-key", &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, http.MethodGet, req.Method)
			assert.Equal(t, BaseURL+"/doctor/find?access_token=api-key&foo=bar", req.URL.String())

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"response":1,"doctors":[]}`)),
				Header:     make(http.Header),
			}, nil
		}),
	})

	body, err := client.Get(context.Background(), "/doctor/find", mapToValues(map[string]string{"foo": "bar"}))
	require.NoError(t, err)
	assert.JSONEq(t, `{"response":1,"doctors":[]}`, string(body))
}

func TestClientGet_TransportError(t *testing.T) {
	client := NewWithHTTP("api-key", &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("boom")
		}),
	})

	_, err := client.Get(context.Background(), "/doctor/find", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "macdent /doctor/find:")
	assert.Contains(t, err.Error(), "boom")
}

func TestClientGet_BadJSON(t *testing.T) {
	client := NewWithHTTP("api-key", &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`not-json`)),
				Header:     make(http.Header),
			}, nil
		}),
	})

	_, err := client.Get(context.Background(), "/doctor/find", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bad json")
}

func TestClientGet_APIErrorEnvelope(t *testing.T) {
	client := NewWithHTTP("api-key", &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"response":0,"error":"denied"}`)),
				Header:     make(http.Header),
			}, nil
		}),
	})

	_, err := client.Get(context.Background(), "/doctor/find", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "macdent /doctor/find: denied")
}

func mapToValues(in map[string]string) url.Values {
	v := url.Values{}
	for k, val := range in {
		v.Set(k, val)
	}
	return v
}
