package modules_test

import (
	"testing"

	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
	"github.com/felipeelias/claude-statusline/internal/modules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContextModule_Name(t *testing.T) {
	m := modules.ContextModule{}
	assert.Equal(t, "context", m.Name())
}

func TestContextModule_Render(t *testing.T) {
	cfg := config.Default()

	t.Run("happy path with usage", func(t *testing.T) {
		data := input.Data{
			ContextWindow: input.ContextWindow{
				UsedPercentage: 40.0,
			},
		}

		result, err := modules.ContextModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Contains(t, result, "40%")
		assert.Contains(t, result, "\u2588\u2588\u2591\u2591\u2591")
		assert.Contains(t, result, "\033[32m")
	})

	t.Run("zero usage", func(t *testing.T) {
		data := input.Data{
			ContextWindow: input.ContextWindow{
				UsedPercentage: 0,
			},
		}

		result, err := modules.ContextModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Contains(t, result, "0%")
		assert.Contains(t, result, "\u2591\u2591\u2591\u2591\u2591")
	})

	t.Run("full usage", func(t *testing.T) {
		data := input.Data{
			ContextWindow: input.ContextWindow{
				UsedPercentage: 100.0,
			},
		}

		result, err := modules.ContextModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Contains(t, result, "100%")
		assert.Contains(t, result, "\u2588\u2588\u2588\u2588\u2588")
	})

	t.Run("threshold above 50 uses warning style", func(t *testing.T) {
		data := input.Data{
			ContextWindow: input.ContextWindow{
				UsedPercentage: 60.0,
			},
		}

		result, err := modules.ContextModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Contains(t, result, "\033[33m")
	})

	t.Run("threshold above 50 still yellow at 75", func(t *testing.T) {
		data := input.Data{
			ContextWindow: input.ContextWindow{
				UsedPercentage: 75.0,
			},
		}

		result, err := modules.ContextModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Contains(t, result, "\033[33m")
	})

	t.Run("threshold above 90 uses high style", func(t *testing.T) {
		data := input.Data{
			ContextWindow: input.ContextWindow{
				UsedPercentage: 95.0,
			},
		}

		result, err := modules.ContextModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Contains(t, result, "\033[31m")
	})

	t.Run("no threshold matches uses base style", func(t *testing.T) {
		data := input.Data{
			ContextWindow: input.ContextWindow{
				UsedPercentage: 30.0,
			},
		}

		result, err := modules.ContextModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Contains(t, result, "\033[32m")
	})

	t.Run("bar_style dots", func(t *testing.T) {
		dotsCfg := cfg
		dotsCfg.Context.BarStyle = "dots"

		data := input.Data{
			ContextWindow: input.ContextWindow{
				UsedPercentage: 60.0,
			},
		}

		result, err := modules.ContextModule{}.Render(data, dotsCfg)
		require.NoError(t, err)
		assert.Contains(t, result, "\u28ff\u28ff\u28ff\u28c0\u28c0")
	})

	t.Run("bar_style line", func(t *testing.T) {
		lineCfg := cfg
		lineCfg.Context.BarStyle = "line"

		data := input.Data{
			ContextWindow: input.ContextWindow{
				UsedPercentage: 40.0,
			},
		}

		result, err := modules.ContextModule{}.Render(data, lineCfg)
		require.NoError(t, err)
		assert.Contains(t, result, "\u2501\u2501\u2500\u2500\u2500")
	})

	t.Run("explicit bar_fill overrides bar_style", func(t *testing.T) {
		overrideCfg := cfg
		overrideCfg.Context.BarStyle = "dots"
		overrideCfg.Context.BarFill = "#"

		data := input.Data{
			ContextWindow: input.ContextWindow{
				UsedPercentage: 60.0,
			},
		}

		result, err := modules.ContextModule{}.Render(data, overrideCfg)
		require.NoError(t, err)
		assert.Contains(t, result, "###\u28c0\u28c0")
	})
}
