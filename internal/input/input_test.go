package input_test

import (
	"strings"
	"testing"

	"github.com/felipeelias/claude-statusline/internal/input"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_FullJSON(t *testing.T) {
	jsonStr := `{
		"session_id": "abc-123",
		"version": "1.0.42",
		"model": {
			"id": "claude-sonnet-4-20250514",
			"display_name": "Claude Sonnet 4"
		},
		"cwd": "/home/user/project",
		"cost": {
			"total_cost_usd": 0.1234,
			"total_duration_ms": 5000,
			"total_lines_added": 100,
			"total_lines_removed": 50
		},
		"context_window": {
			"used_percentage": 35.5,
			"remaining_percentage": 64.5,
			"context_window_size": 200000
		}
	}`

	data, err := input.Parse(strings.NewReader(jsonStr))
	require.NoError(t, err)

	assert.Equal(t, "abc-123", data.SessionID)
	assert.Equal(t, "1.0.42", data.Version)

	assert.Equal(t, "claude-sonnet-4-20250514", data.Model.ID)
	assert.Equal(t, "Claude Sonnet 4", data.Model.DisplayName)

	assert.Equal(t, "/home/user/project", data.Cwd)

	assert.InDelta(t, 0.1234, data.Cost.TotalCostUSD, 0.0001)
	assert.Equal(t, 5000, data.Cost.TotalDurationMs)
	assert.Equal(t, 100, data.Cost.TotalLinesAdded)
	assert.Equal(t, 50, data.Cost.TotalLinesRemoved)

	assert.InDelta(t, 35.5, data.ContextWindow.UsedPercentage, 0.01)
	assert.InDelta(t, 64.5, data.ContextWindow.RemainingPercentage, 0.01)
	assert.Equal(t, 200000, data.ContextWindow.ContextWindowSize)
}

func TestParse_EmptyJSON(t *testing.T) {
	data, err := input.Parse(strings.NewReader("{}"))
	require.NoError(t, err)

	assert.Equal(t, "", data.SessionID)
	assert.Equal(t, "", data.Version)
	assert.Equal(t, "", data.Model.ID)
	assert.Equal(t, "", data.Model.DisplayName)
	assert.Equal(t, "", data.Cwd)
	assert.InDelta(t, 0.0, data.Cost.TotalCostUSD, 0.0001)
	assert.Equal(t, 0, data.Cost.TotalDurationMs)
	assert.Equal(t, 0, data.Cost.TotalLinesAdded)
	assert.Equal(t, 0, data.Cost.TotalLinesRemoved)
	assert.InDelta(t, 0.0, data.ContextWindow.UsedPercentage, 0.01)
	assert.InDelta(t, 0.0, data.ContextWindow.RemainingPercentage, 0.01)
	assert.Equal(t, 0, data.ContextWindow.ContextWindowSize)
}

func TestParse_InvalidJSON(t *testing.T) {
	_, err := input.Parse(strings.NewReader("not json"))
	assert.Error(t, err)
}
