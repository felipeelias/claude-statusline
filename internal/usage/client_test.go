package usage

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Fetch(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Contains(t, r.Header.Get("Authorization"), "Bearer test-token")
			assert.Equal(t, betaHeader, r.Header.Get("anthropic-beta"))

			resp := UsageData{
				FiveHour: Window{Utilisation: 42, ResetsAt: "2025-01-15T10:00:00Z"},
				SevenDay: Window{Utilisation: 15, ResetsAt: "2025-01-20T00:00:00Z"},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer srv.Close()

		c := &client{httpClient: srv.Client(), endpoint: srv.URL}
		data, err := c.fetch("test-token")

		require.NoError(t, err)
		assert.InDelta(t, 42, data.FiveHour.Utilisation, 0.1)
		assert.InDelta(t, 15, data.SevenDay.Utilisation, 0.1)
		assert.Equal(t, "2025-01-15T10:00:00Z", data.FiveHour.ResetsAt)
	})

	t.Run("401 token expired", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer srv.Close()

		c := &client{httpClient: srv.Client(), endpoint: srv.URL}
		_, err := c.fetch("bad-token")

		assert.ErrorIs(t, err, errTokenExpired)
	})

	t.Run("429 rate limited with Retry-After", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
		}))
		defer srv.Close()

		c := &client{httpClient: srv.Client(), endpoint: srv.URL}
		_, err := c.fetch("test-token")

		assert.ErrorIs(t, err, errRateLimited)

		var rle *rateLimitError
		require.ErrorAs(t, err, &rle)
		assert.Equal(t, 60*time.Second, rle.retryAfter)
	})

	t.Run("429 rate limited without Retry-After", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
		}))
		defer srv.Close()

		c := &client{httpClient: srv.Client(), endpoint: srv.URL}
		_, err := c.fetch("test-token")

		assert.ErrorIs(t, err, errRateLimited)

		var rle *rateLimitError
		require.ErrorAs(t, err, &rle)
		assert.Equal(t, 2*time.Minute, rle.retryAfter)
	})

	t.Run("500 server error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		c := &client{httpClient: srv.Client(), endpoint: srv.URL}
		_, err := c.fetch("test-token")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "500")
	})

	t.Run("invalid JSON response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("not json"))
		}))
		defer srv.Close()

		c := &client{httpClient: srv.Client(), endpoint: srv.URL}
		_, err := c.fetch("test-token")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode response")
	})

	t.Run("connection refused", func(t *testing.T) {
		c := &client{
			httpClient: &http.Client{Timeout: 100 * time.Millisecond},
			endpoint:   "http://127.0.0.1:1",
		}

		_, err := c.fetch("test-token")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request failed")
	})
}

func TestParseRetryAfter(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected time.Duration
	}{
		{name: "empty", header: "", expected: 2 * time.Minute},
		{name: "seconds", header: "120", expected: 120 * time.Second},
		{name: "invalid", header: "garbage", expected: 2 * time.Minute},
		{name: "zero", header: "0", expected: 2 * time.Minute},
		{name: "negative", header: "-5", expected: 2 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseRetryAfter(tt.header)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestErrors(t *testing.T) {
	t.Run("rateLimitError unwraps to errRateLimited", func(t *testing.T) {
		err := &rateLimitError{retryAfter: 30 * time.Second}
		assert.True(t, errors.Is(err, errRateLimited))
	})

	t.Run("rateLimitError message", func(t *testing.T) {
		err := &rateLimitError{retryAfter: 30 * time.Second}
		assert.Contains(t, err.Error(), "30s")
	})
}
