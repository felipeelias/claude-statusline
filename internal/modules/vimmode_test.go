package modules_test

import (
	"testing"

	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
	"github.com/felipeelias/claude-statusline/internal/modules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVimModeModule_Name(t *testing.T) {
	m := modules.VimModeModule{}
	assert.Equal(t, "vim_mode", m.Name())
}

func TestVimModeModule_Render(t *testing.T) {
	cfg := config.Default()

	t.Run("renders NORMAL mode", func(t *testing.T) {
		data := input.Data{Vim: &input.Vim{Mode: "NORMAL"}}

		result, err := modules.VimModeModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Contains(t, result, "NORMAL")
	})

	t.Run("renders INSERT mode", func(t *testing.T) {
		data := input.Data{Vim: &input.Vim{Mode: "INSERT"}}

		result, err := modules.VimModeModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Contains(t, result, "INSERT")
	})

	t.Run("nil vim renders empty", func(t *testing.T) {
		data := input.Data{Vim: nil}

		result, err := modules.VimModeModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("empty mode renders empty", func(t *testing.T) {
		data := input.Data{Vim: &input.Vim{Mode: ""}}

		result, err := modules.VimModeModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("applies style", func(t *testing.T) {
		data := input.Data{Vim: &input.Vim{Mode: "NORMAL"}}

		result, err := modules.VimModeModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Contains(t, result, "\033[1;33m") // bold yellow
		assert.Contains(t, result, "\033[0m")    // reset
	})

	t.Run("custom format", func(t *testing.T) {
		customCfg := cfg
		customCfg.VimMode.Format = "[{{.Mode}}]"

		data := input.Data{Vim: &input.Vim{Mode: "NORMAL"}}

		result, err := modules.VimModeModule{}.Render(data, customCfg)
		require.NoError(t, err)
		assert.Contains(t, result, "[NORMAL]")
	})
}
