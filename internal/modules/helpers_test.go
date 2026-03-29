package modules_test

import (
	"testing"

	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
	"github.com/felipeelias/claude-statusline/internal/modules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveBarChars_ViaContextModule(t *testing.T) {
	baseCfg := config.Default()
	baseCfg.Context.Format = "{{.Bar}}"
	baseCfg.Context.BarWidth = 5

	data := input.Data{
		ContextWindow: input.ContextWindow{UsedPercentage: 60.0},
	}

	t.Run("no bar_style uses classic defaults", func(t *testing.T) {
		result, err := modules.ContextModule{}.Render(data, baseCfg)
		require.NoError(t, err)
		assert.Contains(t, result, "\u2588\u2588\u2588\u2591\u2591")
	})

	t.Run("bar_style dots", func(t *testing.T) {
		cfg := baseCfg
		cfg.Context.BarStyle = "dots"
		result, err := modules.ContextModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Contains(t, result, "\u28ff\u28ff\u28ff\u28c0\u28c0")
	})

	t.Run("bar_style blocks", func(t *testing.T) {
		cfg := baseCfg
		cfg.Context.BarStyle = "blocks"
		result, err := modules.ContextModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Contains(t, result, "\u2588\u2588\u2588\u2592\u2592")
	})

	t.Run("bar_style line", func(t *testing.T) {
		cfg := baseCfg
		cfg.Context.BarStyle = "line"
		result, err := modules.ContextModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Contains(t, result, "\u2501\u2501\u2501\u2500\u2500")
	})

	t.Run("bar_style squares", func(t *testing.T) {
		cfg := baseCfg
		cfg.Context.BarStyle = "squares"
		result, err := modules.ContextModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Contains(t, result, "\u25fc\u25fc\u25fc\u25fb\u25fb")
	})

	t.Run("explicit bar_fill overrides bar_style", func(t *testing.T) {
		cfg := baseCfg
		cfg.Context.BarStyle = "dots"
		cfg.Context.BarFill = "#"
		result, err := modules.ContextModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Contains(t, result, "###\u28c0\u28c0")
	})

	t.Run("explicit bar_empty overrides bar_style", func(t *testing.T) {
		cfg := baseCfg
		cfg.Context.BarStyle = "dots"
		cfg.Context.BarEmpty = "O"
		result, err := modules.ContextModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Contains(t, result, "\u28ff\u28ff\u28ffOO")
	})

	t.Run("explicit both override bar_style completely", func(t *testing.T) {
		cfg := baseCfg
		cfg.Context.BarStyle = "dots"
		cfg.Context.BarFill = "#"
		cfg.Context.BarEmpty = "-"
		result, err := modules.ContextModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Contains(t, result, "###--")
	})

	t.Run("unknown bar_style falls back to classic", func(t *testing.T) {
		cfg := baseCfg
		cfg.Context.BarStyle = "unknown"
		result, err := modules.ContextModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Contains(t, result, "\u2588\u2588\u2588\u2591\u2591")
	})

	t.Run("explicit chars without bar_style", func(t *testing.T) {
		cfg := baseCfg
		cfg.Context.BarFill = "#"
		cfg.Context.BarEmpty = "-"
		result, err := modules.ContextModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Contains(t, result, "###--")
	})
}
