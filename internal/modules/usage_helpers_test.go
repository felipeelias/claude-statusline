package modules_test

import (
	"testing"
	"time"

	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
	"github.com/felipeelias/claude-statusline/internal/modules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsageModule_ResetTimestamp(t *testing.T) {
	resetCfg := config.Default()
	resetCfg.Usage.Format = "{{.BlockResets}}"
	resetCfg.Usage.Style = ""

	t.Run("zero resets_at renders empty", func(t *testing.T) {
		data := input.Data{
			RateLimits: &input.RateLimits{
				FiveHour: input.RateLimitWindow{ResetsAt: 0},
			},
		}

		result, err := modules.UsageModule{}.Render(data, resetCfg)

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("past time renders 0m", func(t *testing.T) {
		data := input.Data{
			RateLimits: &input.RateLimits{
				FiveHour: input.RateLimitWindow{
					ResetsAt: time.Now().Add(-1 * time.Hour).Unix(),
				},
			},
		}

		result, err := modules.UsageModule{}.Render(data, resetCfg)

		require.NoError(t, err)
		assert.Contains(t, result, "0m")
	})

	t.Run("minutes only", func(t *testing.T) {
		data := input.Data{
			RateLimits: &input.RateLimits{
				FiveHour: input.RateLimitWindow{
					ResetsAt: time.Now().Add(45 * time.Minute).Unix(),
				},
			},
		}

		result, err := modules.UsageModule{}.Render(data, resetCfg)

		require.NoError(t, err)
		assert.Contains(t, result, "m")
		assert.NotContains(t, result, "h")
		assert.NotContains(t, result, "d")
	})

	t.Run("hours and minutes", func(t *testing.T) {
		data := input.Data{
			RateLimits: &input.RateLimits{
				FiveHour: input.RateLimitWindow{
					ResetsAt: time.Now().Add(2*time.Hour + 30*time.Minute).Unix(),
				},
			},
		}

		result, err := modules.UsageModule{}.Render(data, resetCfg)

		require.NoError(t, err)
		assert.Contains(t, result, "h")
		assert.Contains(t, result, "m")
		assert.NotContains(t, result, "d")
	})

	t.Run("days and hours", func(t *testing.T) {
		data := input.Data{
			RateLimits: &input.RateLimits{
				FiveHour: input.RateLimitWindow{
					ResetsAt: time.Now().Add(3*24*time.Hour + 5*time.Hour).Unix(),
				},
			},
		}

		result, err := modules.UsageModule{}.Render(data, resetCfg)

		require.NoError(t, err)
		assert.Contains(t, result, "d")
		assert.Contains(t, result, "h")
	})
}
