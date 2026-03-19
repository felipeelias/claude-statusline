package modules

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormatResetTime(t *testing.T) {
	tests := []struct {
		name     string
		resetAt  string
		expected string
	}{
		{name: "empty string", resetAt: "", expected: ""},
		{name: "invalid format", resetAt: "not-a-time", expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, formatResetTime(tt.resetAt))
		})
	}

	t.Run("past time returns 0m", func(t *testing.T) {
		past := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
		assert.Equal(t, "0m", formatResetTime(past))
	})

	t.Run("minutes only", func(t *testing.T) {
		future := time.Now().Add(45 * time.Minute).Format(time.RFC3339)
		result := formatResetTime(future)
		assert.Contains(t, result, "m")
		assert.NotContains(t, result, "h")
		assert.NotContains(t, result, "d")
	})

	t.Run("hours and minutes", func(t *testing.T) {
		future := time.Now().Add(2*time.Hour + 30*time.Minute).Format(time.RFC3339)
		result := formatResetTime(future)
		assert.Contains(t, result, "h")
		assert.Contains(t, result, "m")
		assert.NotContains(t, result, "d")
	})

	t.Run("days and hours", func(t *testing.T) {
		future := time.Now().Add(3*24*time.Hour + 5*time.Hour).Format(time.RFC3339)
		result := formatResetTime(future)
		assert.Contains(t, result, "d")
		assert.Contains(t, result, "h")
	})

	t.Run("RFC3339 with nanoseconds", func(t *testing.T) {
		future := time.Now().Add(2 * time.Hour).Format(time.RFC3339Nano)
		result := formatResetTime(future)
		assert.Contains(t, result, "h")
	})
}
