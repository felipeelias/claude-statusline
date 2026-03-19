package modules_test

import (
	"testing"
	"time"

	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
	"github.com/felipeelias/claude-statusline/internal/modules"
	"github.com/felipeelias/claude-statusline/internal/usage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsageModule_Name(t *testing.T) {
	m := modules.NewUsageModule()
	assert.Equal(t, "usage", m.Name())
}

func TestUsageModule_Render(t *testing.T) {
	cfg := config.Default()

	mockData := &usage.UsageData{
		FiveHour: usage.Window{
			Utilisation: 42,
			ResetsAt:    time.Now().Add(2 * time.Hour).Format(time.RFC3339),
		},
		SevenDay: usage.Window{
			Utilisation: 15,
			ResetsAt:    time.Now().Add(72 * time.Hour).Format(time.RFC3339),
		},
	}

	t.Run("renders block percentage", func(t *testing.T) {
		m := modules.NewUsageModuleWithFetcher(usage.MockFetcher{Data: mockData})
		result, err := m.Render(input.Data{}, cfg)

		require.NoError(t, err)
		assert.Contains(t, result, "42%")
	})

	t.Run("renders progress bar", func(t *testing.T) {
		m := modules.NewUsageModuleWithFetcher(usage.MockFetcher{Data: mockData})
		result, err := m.Render(input.Data{}, cfg)

		require.NoError(t, err)
		assert.Contains(t, result, "\u2588\u2588\u2591\u2591\u2591")
	})

	t.Run("nil data returns empty", func(t *testing.T) {
		m := modules.NewUsageModuleWithFetcher(usage.MockFetcher{})
		result, err := m.Render(input.Data{}, cfg)

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("threshold below 75 uses base style", func(t *testing.T) {
		m := modules.NewUsageModuleWithFetcher(usage.MockFetcher{Data: mockData})
		result, err := m.Render(input.Data{}, cfg)

		require.NoError(t, err)
		// green ANSI code
		assert.Contains(t, result, "\033[32m")
	})

	t.Run("threshold above 75 uses warning style", func(t *testing.T) {
		highData := &usage.UsageData{
			FiveHour: usage.Window{Utilisation: 80},
			SevenDay: usage.Window{Utilisation: 10},
		}

		m := modules.NewUsageModuleWithFetcher(usage.MockFetcher{Data: highData})
		result, err := m.Render(input.Data{}, cfg)

		require.NoError(t, err)
		// yellow ANSI code
		assert.Contains(t, result, "\033[33m")
	})

	t.Run("threshold above 90 uses high style", func(t *testing.T) {
		criticalData := &usage.UsageData{
			FiveHour: usage.Window{Utilisation: 95},
			SevenDay: usage.Window{Utilisation: 10},
		}

		m := modules.NewUsageModuleWithFetcher(usage.MockFetcher{Data: criticalData})
		result, err := m.Render(input.Data{}, cfg)

		require.NoError(t, err)
		// red ANSI code
		assert.Contains(t, result, "\033[31m")
	})

	t.Run("custom format with weekly", func(t *testing.T) {
		customCfg := config.Default()
		customCfg.Usage.Format = `{{printf "%.0f" .WeeklyPct}}%`

		m := modules.NewUsageModuleWithFetcher(usage.MockFetcher{Data: mockData})
		result, err := m.Render(input.Data{}, customCfg)

		require.NoError(t, err)
		assert.Contains(t, result, "15%")
	})

	t.Run("format with reset time", func(t *testing.T) {
		customCfg := config.Default()
		customCfg.Usage.Format = `{{.BlockResets}}`

		m := modules.NewUsageModuleWithFetcher(usage.MockFetcher{Data: mockData})
		result, err := m.Render(input.Data{}, customCfg)

		require.NoError(t, err)
		assert.Contains(t, result, "h")
	})

	t.Run("conditional format hides below threshold", func(t *testing.T) {
		customCfg := config.Default()
		customCfg.Usage.Format = `{{if ge .BlockPct 70.0}}{{printf "%.0f" .BlockPct}}%{{end}}{{if ge .WeeklyPct 80.0}} W:{{printf "%.0f" .WeeklyPct}}%{{end}}`

		lowData := &usage.UsageData{
			FiveHour: usage.Window{Utilisation: 30},
			SevenDay: usage.Window{Utilisation: 20},
		}

		m := modules.NewUsageModuleWithFetcher(usage.MockFetcher{Data: lowData})
		result, err := m.Render(input.Data{}, customCfg)

		require.NoError(t, err)
		assert.NotContains(t, result, "30%")
		assert.NotContains(t, result, "20%")
	})

	t.Run("conditional format shows above threshold", func(t *testing.T) {
		customCfg := config.Default()
		customCfg.Usage.Format = `{{if ge .BlockPct 70.0}}{{printf "%.0f" .BlockPct}}%{{end}}{{if ge .WeeklyPct 80.0}} W:{{printf "%.0f" .WeeklyPct}}%{{end}}`

		highData := &usage.UsageData{
			FiveHour: usage.Window{Utilisation: 85},
			SevenDay: usage.Window{Utilisation: 90},
		}

		m := modules.NewUsageModuleWithFetcher(usage.MockFetcher{Data: highData})
		result, err := m.Render(input.Data{}, customCfg)

		require.NoError(t, err)
		assert.Contains(t, result, "85%")
		assert.Contains(t, result, "W:90%")
	})

	t.Run("test mode uses mock data", func(t *testing.T) {
		testCfg := config.Default()
		testCfg.Usage.TestMode = true

		m := modules.NewUsageModule()
		result, err := m.Render(input.Data{}, testCfg)

		require.NoError(t, err)
		assert.Contains(t, result, "42%")
		assert.Contains(t, result, "W:15%")
	})
}
