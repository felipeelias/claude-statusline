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

func TestShortName(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		displayName string
		want        string
	}{
		{"sonnet", "claude-sonnet-4-6-20250514", "Claude Sonnet 4.6", "Sonnet 4.6"},
		{"opus", "claude-opus-4-6-20250514", "Claude Opus 4.6", "Opus 4.6"},
		{"haiku", "claude-haiku-4-5-20251001", "Claude Haiku 4.5", "Haiku 4.5"},
		{"no date suffix", "claude-sonnet-4-6", "Claude Sonnet 4.6", "Sonnet 4.6"},
		{"unknown model falls back", "gpt-4o", "GPT-4o", "GPT-4o"},
		{"empty id falls back", "", "Some Model", "Some Model"},
		{"prefix mismatch falls back", "xclaude-sonnet-4-6", "X", "X"},
		{"suffix mismatch falls back", "claude-sonnet-4-6-foo", "X", "X"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, modules.ShortName(tt.id, tt.displayName))
		})
	}
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

	t.Run("empty display name and ID returns empty string", func(t *testing.T) {
		data := input.Data{
			Model: input.Model{},
		}

		result, err := modules.ModelModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("empty display name with ID renders via ID format", func(t *testing.T) {
		customCfg := config.Default()
		customCfg.Model.Format = "{{.ID}}"

		data := input.Data{
			Model: input.Model{
				ID:          "claude-sonnet-4-6-20250514",
				DisplayName: "",
			},
		}

		result, err := modules.ModelModule{}.Render(data, customCfg)
		require.NoError(t, err)
		assert.Contains(t, result, "claude-sonnet-4-6-20250514")
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

	t.Run("short name format", func(t *testing.T) {
		customCfg := config.Default()
		customCfg.Model.Format = "{{.Short}}"

		data := input.Data{
			Model: input.Model{
				ID:          "claude-sonnet-4-6-20250514",
				DisplayName: "Claude Sonnet 4.6",
			},
		}

		result, err := modules.ModelModule{}.Render(data, customCfg)
		require.NoError(t, err)
		assert.Contains(t, result, "Sonnet 4.6")
	})

	t.Run("ID format", func(t *testing.T) {
		customCfg := config.Default()
		customCfg.Model.Format = "{{.ID}}"

		data := input.Data{
			Model: input.Model{
				ID:          "claude-sonnet-4-6-20250514",
				DisplayName: "Claude Sonnet 4.6",
			},
		}

		result, err := modules.ModelModule{}.Render(data, customCfg)
		require.NoError(t, err)
		assert.Contains(t, result, "claude-sonnet-4-6-20250514")
	})
}
