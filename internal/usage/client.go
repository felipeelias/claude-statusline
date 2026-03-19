package usage

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

const (
	usageEndpoint = "https://api.anthropic.com/api/oauth/usage"
	httpTimeout   = 5 * time.Second
	betaHeader    = "oauth-2025-04-20"
)

var (
	errTokenExpired = errors.New("OAuth token expired")
	errRateLimited  = errors.New("rate limited by API")
)

// rateLimitError wraps errRateLimited with a Retry-After duration.
type rateLimitError struct {
	retryAfter time.Duration
}

func (e *rateLimitError) Error() string {
	return fmt.Sprintf("rate limited (retry after %s)", e.retryAfter)
}

func (e *rateLimitError) Unwrap() error { return errRateLimited }

// Window represents a usage window (5-hour or 7-day).
type Window struct {
	Utilisation float64 `json:"utilization"`
	ResetsAt    string  `json:"resets_at"`
}

// UsageData represents the API response from the OAuth usage endpoint.
type UsageData struct {
	FiveHour Window `json:"five_hour"`
	SevenDay Window `json:"seven_day"`
}

// client fetches usage data from the Anthropic OAuth API.
type client struct {
	httpClient *http.Client
	endpoint   string
}

func newClient() *client {
	return &client{
		httpClient: &http.Client{Timeout: httpTimeout},
		endpoint:   usageEndpoint,
	}
}

func (c *client) fetch(token string) (*UsageData, error) {
	req, err := http.NewRequest("GET", c.endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("anthropic-beta", betaHeader)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return nil, errTokenExpired
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			return nil, &rateLimitError{retryAfter: parseRetryAfter(resp.Header.Get("Retry-After"))}
		}

		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var data UsageData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &data, nil
}

// parseRetryAfter parses the Retry-After header value.
// Returns a default of 2 minutes if missing or unparseable.
func parseRetryAfter(header string) time.Duration {
	const defaultRetry = 2 * time.Minute

	if header == "" {
		return defaultRetry
	}

	if seconds, err := strconv.Atoi(header); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}

	if t, err := time.Parse(time.RFC1123, header); err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}

	return defaultRetry
}
