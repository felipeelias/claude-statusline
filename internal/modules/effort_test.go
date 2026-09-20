package modules_test

import (
	"testing"

	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
	"github.com/felipeelias/claude-statusline/internal/modules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEffortModule_Name(t *testing.T) {
	assert.Equal(t, "effort", modules.EffortModule{}.Name())
}

func TestEffortModule_Render(t *testing.T) {
	cfg := config.Default()

	t.Run("renders effective level", func(t *testing.T) {
		data := input.Data{Effort: input.Effort{Level: "xhigh"}}

		result, err := modules.EffortModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Contains(t, result, "xhigh")
		assert.Contains(t, result, "\033[1;33m")
		assert.Contains(t, result, "\033[0m")
	})

	t.Run("empty level renders empty", func(t *testing.T) {
		result, err := modules.EffortModule{}.Render(input.Data{}, cfg)
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("custom format", func(t *testing.T) {
		customCfg := cfg
		customCfg.Effort.Format = "effort:{{.Level}}"

		data := input.Data{Effort: input.Effort{Level: "low"}}

		result, err := modules.EffortModule{}.Render(data, customCfg)
		require.NoError(t, err)
		assert.Contains(t, result, "effort:low")
	})
}
