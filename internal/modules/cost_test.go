package modules_test

import (
	"testing"

	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
	"github.com/felipeelias/claude-statusline/internal/modules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCostModule_Name(t *testing.T) {
	m := modules.CostModule{}
	assert.Equal(t, "cost", m.Name())
}

func TestCostModule_Render(t *testing.T) {
	cfg := config.Default()

	t.Run("happy path with cost", func(t *testing.T) {
		data := input.Data{
			Cost: input.Cost{TotalCostUSD: 0.42},
		}

		result, err := modules.CostModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Contains(t, result, "$0.42")
		assert.Contains(t, result, "\033[32m")
	})

	t.Run("zero cost", func(t *testing.T) {
		data := input.Data{
			Cost: input.Cost{TotalCostUSD: 0},
		}

		result, err := modules.CostModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Contains(t, result, "$0.00")
	})

	t.Run("threshold above 1.0 uses warning style", func(t *testing.T) {
		data := input.Data{
			Cost: input.Cost{TotalCostUSD: 2.50},
		}

		result, err := modules.CostModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Contains(t, result, "$2.50")
		assert.Contains(t, result, "\033[33m")
	})

	t.Run("threshold above 5.0 uses high style", func(t *testing.T) {
		data := input.Data{
			Cost: input.Cost{TotalCostUSD: 10.00},
		}

		result, err := modules.CostModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Contains(t, result, "$10.00")
		assert.Contains(t, result, "\033[31m")
	})

	t.Run("no threshold matches uses base style", func(t *testing.T) {
		data := input.Data{
			Cost: input.Cost{TotalCostUSD: 0.50},
		}

		result, err := modules.CostModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Contains(t, result, "\033[32m")
	})

	t.Run("burn rate template field", func(t *testing.T) {
		burnCfg := cfg
		burnCfg.Cost.Format = `${{printf "%.2f" .TotalCostUSD}} ({{printf "%.2f" .BurnRate}}/hr)`

		data := input.Data{
			Cost: input.Cost{
				TotalCostUSD:    1.00,
				TotalDurationMs: 1_800_000, // 30 minutes
			},
		}

		result, err := modules.CostModule{}.Render(data, burnCfg)
		require.NoError(t, err)
		assert.Contains(t, result, "$1.00")
		assert.Contains(t, result, "2.00/hr")
	})

	t.Run("burn rate zero when duration is zero", func(t *testing.T) {
		burnCfg := cfg
		burnCfg.Cost.Format = `{{printf "%.2f" .BurnRate}}/hr`

		data := input.Data{
			Cost: input.Cost{
				TotalCostUSD:    1.00,
				TotalDurationMs: 0,
			},
		}

		result, err := modules.CostModule{}.Render(data, burnCfg)
		require.NoError(t, err)
		assert.Contains(t, result, "0.00/hr")
	})

	t.Run("api duration template field", func(t *testing.T) {
		apiCfg := cfg
		apiCfg.Cost.Format = `{{.APIDurationMs}}ms`

		data := input.Data{
			Cost: input.Cost{TotalAPIDurationMs: 2300},
		}

		result, err := modules.CostModule{}.Render(data, apiCfg)
		require.NoError(t, err)
		assert.Contains(t, result, "2300ms")
	})
}
