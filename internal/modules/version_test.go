package modules_test

import (
	"testing"

	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
	"github.com/felipeelias/claude-statusline/internal/modules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionModule_Name(t *testing.T) {
	m := modules.VersionModule{}
	assert.Equal(t, "version", m.Name())
}

func TestVersionModule_Render(t *testing.T) {
	cfg := config.Default()

	t.Run("renders version with default format", func(t *testing.T) {
		data := input.Data{Version: "1.0.33"}

		result, err := modules.VersionModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Contains(t, result, "v1.0.33")
	})

	t.Run("empty version renders empty", func(t *testing.T) {
		data := input.Data{Version: ""}

		result, err := modules.VersionModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("applies style", func(t *testing.T) {
		data := input.Data{Version: "2.1.80"}

		result, err := modules.VersionModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Contains(t, result, "\033[2m") // dim
	})

	t.Run("custom format", func(t *testing.T) {
		customCfg := cfg
		customCfg.Version.Format = "Claude {{.Version}}"

		data := input.Data{Version: "1.0.33"}

		result, err := modules.VersionModule{}.Render(data, customCfg)
		require.NoError(t, err)
		assert.Contains(t, result, "Claude 1.0.33")
	})
}
