package modules_test

import (
	"strings"
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

	t.Run("no bar marker below first marker threshold", func(t *testing.T) {
		data := input.Data{
			ContextWindow: input.ContextWindow{
				UsedPercentage: 15.0,
			},
		}

		result, err := modules.ContextModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.NotContains(t, result, "▲")
	})

	t.Run("orange bar marker between warn and high marker thresholds", func(t *testing.T) {
		data := input.Data{
			ContextWindow: input.ContextWindow{
				UsedPercentage: 25.0,
			},
		}

		result, err := modules.ContextModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Contains(t, result, "\033[38;5;208m▲\033[0m")
		assert.True(t, strings.HasPrefix(result, "\033[38;5;208m▲\033[0m"))
		assert.True(t, strings.HasSuffix(result, "\033[38;5;208m▲\033[0m"))
	})

	t.Run("orange-red bar marker between high and crit marker thresholds", func(t *testing.T) {
		data := input.Data{
			ContextWindow: input.ContextWindow{
				UsedPercentage: 32.0,
			},
		}

		result, err := modules.ContextModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Contains(t, result, "\033[38;5;202m▲\033[0m")
		assert.True(t, strings.HasPrefix(result, "\033[38;5;202m▲\033[0m"))
		assert.True(t, strings.HasSuffix(result, "\033[38;5;202m▲\033[0m"))
	})

	t.Run("red bar marker above crit marker threshold", func(t *testing.T) {
		data := input.Data{
			ContextWindow: input.ContextWindow{
				UsedPercentage: 40.0,
			},
		}

		result, err := modules.ContextModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Contains(t, result, "\033[31m▲\033[0m")
		assert.True(t, strings.HasPrefix(result, "\033[31m▲\033[0m"))
		assert.True(t, strings.HasSuffix(result, "\033[31m▲\033[0m"))
	})

	t.Run("empty bar markers list yields no marker", func(t *testing.T) {
		cfgNoMarkers := config.Default()
		cfgNoMarkers.Context.BarMarkers = nil

		data := input.Data{
			ContextWindow: input.ContextWindow{
				UsedPercentage: 95.0,
			},
		}

		result, err := modules.ContextModule{}.Render(data, cfgNoMarkers)
		require.NoError(t, err)
		assert.NotContains(t, result, "▲")
	})
}
