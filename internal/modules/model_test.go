package modules_test

import (
	"testing"

	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
	"github.com/felipeelias/claude-statusline/internal/modules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModelModule_Name(t *testing.T) {
	m := modules.ModelModule{}
	assert.Equal(t, "model", m.Name())
}

func TestModelModule_Render(t *testing.T) {
	cfg := config.Default()

	t.Run("happy path with display name", func(t *testing.T) {
		data := input.Data{
			Model: input.Model{
				ID:          "claude-opus-4",
				DisplayName: "Claude Opus 4",
			},
		}

		result, err := modules.ModelModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Contains(t, result, "Claude Opus 4")
		assert.Contains(t, result, "\033[1m")
		assert.Contains(t, result, "\033[0m")
	})

	t.Run("empty display name returns empty string", func(t *testing.T) {
		data := input.Data{
			Model: input.Model{
				ID:          "claude-opus-4",
				DisplayName: "",
			},
		}

		result, err := modules.ModelModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("custom format template", func(t *testing.T) {
		customCfg := config.Default()
		customCfg.Model.Format = "model: {{.DisplayName}}"

		data := input.Data{
			Model: input.Model{DisplayName: "Sonnet"},
		}

		result, err := modules.ModelModule{}.Render(data, customCfg)
		require.NoError(t, err)
		assert.Contains(t, result, "model: Sonnet")
	})
}
