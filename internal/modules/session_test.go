package modules_test

import (
	"testing"

	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
	"github.com/felipeelias/claude-statusline/internal/modules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionTimerModule_Name(t *testing.T) {
	m := modules.SessionTimerModule{}
	assert.Equal(t, "session_timer", m.Name())
}

func TestSessionTimerModule_Render(t *testing.T) {
	cfg := config.Default()

	t.Run("formats hours minutes seconds", func(t *testing.T) {
		data := input.Data{
			Cost: input.Cost{TotalDurationMs: 3661000},
		}

		result, err := modules.SessionTimerModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Contains(t, result, "1h01m01s")
	})

	t.Run("formats minutes seconds without hours", func(t *testing.T) {
		data := input.Data{
			Cost: input.Cost{TotalDurationMs: 125000},
		}

		result, err := modules.SessionTimerModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Contains(t, result, "02m05s")
	})

	t.Run("zero duration returns empty string", func(t *testing.T) {
		data := input.Data{
			Cost: input.Cost{TotalDurationMs: 0},
		}

		result, err := modules.SessionTimerModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}
