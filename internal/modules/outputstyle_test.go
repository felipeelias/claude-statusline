package modules_test

import (
	"testing"

	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
	"github.com/felipeelias/claude-statusline/internal/modules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOutputStyleModuleRendersName(t *testing.T) {
	cfg := config.Default()
	cfg.OutputStyle.Disabled = false
	data := input.Data{OutputStyle: input.OutputStyle{Name: "Concise"}}

	result, err := modules.OutputStyleModule{}.Render(data, cfg)
	require.NoError(t, err)
	assert.Contains(t, result, "Concise")
}

func TestOutputStyleModuleEmptyWithoutPayload(t *testing.T) {
	cfg := config.Default()
	cfg.OutputStyle.Disabled = false

	result, err := modules.OutputStyleModule{}.Render(input.Data{}, cfg)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestOutputStyleModuleHidesDefaultStyle(t *testing.T) {
	cfg := config.Default()
	cfg.OutputStyle.Disabled = false
	data := input.Data{OutputStyle: input.OutputStyle{Name: "default"}}

	result, err := modules.OutputStyleModule{}.Render(data, cfg)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestOutputStyleModuleShowsDefaultWhenAsked(t *testing.T) {
	cfg := config.Default()
	cfg.OutputStyle.Disabled = false
	cfg.OutputStyle.HideDefault = false
	data := input.Data{OutputStyle: input.OutputStyle{Name: "default"}}

	result, err := modules.OutputStyleModule{}.Render(data, cfg)
	require.NoError(t, err)
	assert.Contains(t, result, "default")
}

func TestOutputStyleModuleCustomFormat(t *testing.T) {
	cfg := config.Default()
	cfg.OutputStyle.Disabled = false
	cfg.OutputStyle.Format = "style:{{.Name}}"
	data := input.Data{OutputStyle: input.OutputStyle{Name: "Explanatory"}}

	result, err := modules.OutputStyleModule{}.Render(data, cfg)
	require.NoError(t, err)
	assert.Contains(t, result, "style:Explanatory")
}
