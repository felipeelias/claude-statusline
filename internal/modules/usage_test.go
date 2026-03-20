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

func TestUsageModule_Name(t *testing.T) {
	m := modules.UsageModule{}
	assert.Equal(t, "usage", m.Name())
}

func TestUsageModule_Render(t *testing.T) {
	cfg := config.Default()

	mockData := input.Data{
		RateLimits: &input.RateLimits{
			FiveHour: input.RateLimitWindow{
				UsedPercentage: 42,
				ResetsAt:       time.Now().Add(2 * time.Hour).Unix(),
			},
			SevenDay: input.RateLimitWindow{
				UsedPercentage: 15,
				ResetsAt:       time.Now().Add(72 * time.Hour).Unix(),
			},
		},
	}

	t.Run("renders block percentage", func(t *testing.T) {
		result, err := modules.UsageModule{}.Render(mockData, cfg)

		require.NoError(t, err)
		assert.Contains(t, result, "42%")
	})

	t.Run("renders weekly percentage", func(t *testing.T) {
		result, err := modules.UsageModule{}.Render(mockData, cfg)

		require.NoError(t, err)
		assert.Contains(t, result, "W:15%")
	})

	t.Run("renders progress bar", func(t *testing.T) {
		result, err := modules.UsageModule{}.Render(mockData, cfg)

		require.NoError(t, err)
		assert.Contains(t, result, "\u2588\u2588\u2591\u2591\u2591")
	})

	t.Run("nil rate_limits returns empty", func(t *testing.T) {
		result, err := modules.UsageModule{}.Render(input.Data{}, cfg)

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("threshold below 75 uses base style", func(t *testing.T) {
		result, err := modules.UsageModule{}.Render(mockData, cfg)

		require.NoError(t, err)
		assert.Contains(t, result, "\033[32m")
	})

	t.Run("threshold above 75 uses warning style", func(t *testing.T) {
		highData := input.Data{
			RateLimits: &input.RateLimits{
				FiveHour: input.RateLimitWindow{UsedPercentage: 80},
				SevenDay: input.RateLimitWindow{UsedPercentage: 10},
			},
		}

		result, err := modules.UsageModule{}.Render(highData, cfg)

		require.NoError(t, err)
		assert.Contains(t, result, "\033[33m")
	})

	t.Run("threshold above 90 uses high style", func(t *testing.T) {
		criticalData := input.Data{
			RateLimits: &input.RateLimits{
				FiveHour: input.RateLimitWindow{UsedPercentage: 95},
				SevenDay: input.RateLimitWindow{UsedPercentage: 10},
			},
		}

		result, err := modules.UsageModule{}.Render(criticalData, cfg)

		require.NoError(t, err)
		assert.Contains(t, result, "\033[31m")
	})

	t.Run("custom format with weekly only", func(t *testing.T) {
		customCfg := config.Default()
		customCfg.Usage.Format = `{{printf "%.0f" .WeeklyPct}}%`

		result, err := modules.UsageModule{}.Render(mockData, customCfg)

		require.NoError(t, err)
		assert.Contains(t, result, "15%")
	})

	t.Run("format with reset time", func(t *testing.T) {
		customCfg := config.Default()
		customCfg.Usage.Format = `{{.BlockResets}}`

		result, err := modules.UsageModule{}.Render(mockData, customCfg)

		require.NoError(t, err)
		assert.Contains(t, result, "h")
	})

	t.Run("conditional format hides below threshold", func(t *testing.T) {
		customCfg := config.Default()
		customCfg.Usage.Format = `{{if ge .BlockPct 70.0}}{{printf "%.0f" .BlockPct}}%{{end}}{{if ge .WeeklyPct 80.0}} W:{{printf "%.0f" .WeeklyPct}}%{{end}}`

		lowData := input.Data{
			RateLimits: &input.RateLimits{
				FiveHour: input.RateLimitWindow{UsedPercentage: 30},
				SevenDay: input.RateLimitWindow{UsedPercentage: 20},
			},
		}

		result, err := modules.UsageModule{}.Render(lowData, customCfg)

		require.NoError(t, err)
		assert.NotContains(t, result, "30%")
		assert.NotContains(t, result, "20%")
	})

	t.Run("conditional format shows above threshold", func(t *testing.T) {
		customCfg := config.Default()
		customCfg.Usage.Format = `{{if ge .BlockPct 70.0}}{{printf "%.0f" .BlockPct}}%{{end}}{{if ge .WeeklyPct 80.0}} W:{{printf "%.0f" .WeeklyPct}}%{{end}}`

		highData := input.Data{
			RateLimits: &input.RateLimits{
				FiveHour: input.RateLimitWindow{UsedPercentage: 85},
				SevenDay: input.RateLimitWindow{UsedPercentage: 90},
			},
		}

		result, err := modules.UsageModule{}.Render(highData, customCfg)

		require.NoError(t, err)
		assert.Contains(t, result, "85%")
		assert.Contains(t, result, "W:90%")
	})
}
