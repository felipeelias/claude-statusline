package modules

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormatResetTimestamp(t *testing.T) {
	t.Run("zero returns empty", func(t *testing.T) {
		assert.Equal(t, "", formatResetTimestamp(0))
	})

	t.Run("past time returns 0m", func(t *testing.T) {
		past := time.Now().Add(-1 * time.Hour).Unix()
		assert.Equal(t, "0m", formatResetTimestamp(past))
	})

	t.Run("minutes only", func(t *testing.T) {
		future := time.Now().Add(45 * time.Minute).Unix()
		result := formatResetTimestamp(future)
		assert.Contains(t, result, "m")
		assert.NotContains(t, result, "h")
		assert.NotContains(t, result, "d")
	})

	t.Run("hours and minutes", func(t *testing.T) {
		future := time.Now().Add(2*time.Hour + 30*time.Minute).Unix()
		result := formatResetTimestamp(future)
		assert.Contains(t, result, "h")
		assert.Contains(t, result, "m")
		assert.NotContains(t, result, "d")
	})

	t.Run("days and hours", func(t *testing.T) {
		future := time.Now().Add(3*24*time.Hour + 5*time.Hour).Unix()
		result := formatResetTimestamp(future)
		assert.Contains(t, result, "d")
		assert.Contains(t, result, "h")
	})
}
