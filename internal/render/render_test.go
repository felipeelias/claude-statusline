package render_test

import (
	"testing"

	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
	"github.com/felipeelias/claude-statusline/internal/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderPlain(t *testing.T) {
	cfg := config.Default()
	data := input.Data{
		Model:         input.Model{DisplayName: "Claude Opus 4"},
		Cwd:           "/tmp/test",
		Cost:          input.Cost{TotalCostUSD: 0.42},
		ContextWindow: input.ContextWindow{UsedPercentage: 42.5},
	}
	result, err := render.Render(cfg, data)
	require.NoError(t, err)
	assert.Contains(t, result, "Claude Opus 4")
	assert.Contains(t, result, "/tmp/test")
	assert.Contains(t, result, "$0.42")
	assert.Contains(t, result, "42%")
	assert.Contains(t, result, " | ")
}

func TestRenderDisabledModule(t *testing.T) {
	cfg := config.Default()
	cfg.Format = "$model | $session_timer | $cost"
	data := input.Data{
		Model: input.Model{DisplayName: "Opus"},
		Cost:  input.Cost{TotalCostUSD: 1.0},
	}
	result, err := render.Render(cfg, data)
	require.NoError(t, err)
	assert.Contains(t, result, "Opus")
	assert.Contains(t, result, "$1.00")
}

func TestRenderEffortModule(t *testing.T) {
	cfg := config.Default()
	cfg.Format = "$effort"
	cfg.Effort.Disabled = false
	data := input.Data{Effort: input.Effort{Level: "max"}}

	result, err := render.Render(cfg, data)
	require.NoError(t, err)
	assert.Contains(t, result, "max")
}

func TestRenderOmitsEmptyPipeDelimitedSections(t *testing.T) {
	tests := []struct {
		name   string
		format string
		want   string
	}{
		{
			name:   "leading",
			format: "$effort | $model",
			want:   "Opus",
		},
		{
			name:   "middle",
			format: "$model | $effort | $cost",
			want:   "Opus | $1.00",
		},
		{
			name:   "trailing",
			format: "$model | $effort",
			want:   "Opus",
		},
		{
			name:   "adjacent",
			format: "$model | $effort | $usage | $cost",
			want:   "Opus | $1.00",
		},
		{
			name:   "repeated separator",
			format: "$model | | $cost",
			want:   "Opus | $1.00",
		},
		{
			name:   "styled whitespace",
			format: "$model | [ ](bold) | $cost",
			want:   "Opus | $1.00",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			cfg := config.Default()
			cfg.Format = testCase.format
			cfg.Model.Style = ""
			cfg.Effort.Disabled = false
			cfg.Cost.Style = ""

			data := input.Data{
				Model: input.Model{DisplayName: "Opus"},
				Cost:  input.Cost{TotalCostUSD: 1},
			}

			result, err := render.Render(cfg, data)
			require.NoError(t, err)
			assert.Equal(t, testCase.want, result)
		})
	}
}

func TestRenderPreservesPipeDelimitedContent(t *testing.T) {
	cfg := config.Default()
	cfg.Format = "$model | [left | right](bold) | literal | $cost"
	cfg.Model.Style = ""
	cfg.Cost.Style = ""
	data := input.Data{
		Model: input.Model{DisplayName: "Opus"},
		Cost:  input.Cost{TotalCostUSD: 1},
	}

	result, err := render.Render(cfg, data)
	require.NoError(t, err)
	assert.Equal(t, "Opus | \033[1mleft | right\033[0m | literal | $1.00", result)
}
func TestRenderPreservesWhitespaceWithoutPipeSeparators(t *testing.T) {
	cfg := config.Default()
	cfg.Format = " "

	result, err := render.Render(cfg, input.Data{})
	require.NoError(t, err)
	assert.Equal(t, " ", result)
}

func TestRenderEffortModuleWithoutPayload(t *testing.T) {
	cfg := config.Default()
	cfg.Format = "$effort"
	cfg.Effort.Disabled = false

	result, err := render.Render(cfg, input.Data{})
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestRenderStyledText(t *testing.T) {
	cfg := config.Default()
	cfg.Format = "[hello](bold green)"
	result, err := render.Render(cfg, input.Data{})
	require.NoError(t, err)
	assert.Contains(t, result, "\033[1;32m")
	assert.Contains(t, result, "hello")
}

func TestRenderUnknownModule(t *testing.T) {
	cfg := config.Default()
	cfg.Format = "$unknown_module"
	result, err := render.Render(cfg, input.Data{})
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestRenderPowerline(t *testing.T) {
	cfg := config.Default()
	cfg.Format = "[](bg:blue)$model[](fg:blue bg:cyan)$directory[](fg:cyan)"
	data := input.Data{
		Model: input.Model{DisplayName: "Opus"},
		Cwd:   "/tmp",
	}
	result, err := render.Render(cfg, data)
	require.NoError(t, err)
	assert.Contains(t, result, "Opus")
	assert.Contains(t, result, "/tmp")
	assert.Contains(t, result, "\033[") // ANSI codes present
}

func TestRenderLiteralText(t *testing.T) {
	cfg := config.Default()
	cfg.Format = "<<< $model >>>"
	data := input.Data{
		Model: input.Model{DisplayName: "Opus"},
	}
	result, err := render.Render(cfg, data)
	require.NoError(t, err)
	assert.Contains(t, result, "<<<")
	assert.Contains(t, result, ">>>")
	assert.Contains(t, result, "Opus")
}

func TestRenderEmptyFormat(t *testing.T) {
	cfg := config.Default()
	cfg.Format = ""
	result, err := render.Render(cfg, input.Data{})
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestRenderInlineStyle(t *testing.T) {
	cfg := config.Default()
	cfg.Format = "[text](cyan)"
	result, err := render.Render(cfg, input.Data{})
	require.NoError(t, err)
	assert.Contains(t, result, "\033[36m")
	assert.Contains(t, result, "text")
}
