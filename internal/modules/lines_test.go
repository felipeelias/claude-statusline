package modules

import (
	"testing"

	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinesChangedModule_Name(t *testing.T) {
	m := LinesChangedModule{}
	assert.Equal(t, "lines_changed", m.Name())
}

func TestLinesChangedModule_Render(t *testing.T) {
	cfg := config.Default()

	t.Run("renders added and removed counts", func(t *testing.T) {
		data := input.Data{
			Cost: input.Cost{TotalLinesAdded: 42, TotalLinesRemoved: 7},
		}

		result, err := LinesChangedModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Contains(t, result, "+42")
		assert.Contains(t, result, "-7")
	})

	t.Run("both zero returns empty string", func(t *testing.T) {
		data := input.Data{
			Cost: input.Cost{TotalLinesAdded: 0, TotalLinesRemoved: 0},
		}

		result, err := LinesChangedModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("only added lines", func(t *testing.T) {
		data := input.Data{
			Cost: input.Cost{TotalLinesAdded: 10, TotalLinesRemoved: 0},
		}

		result, err := LinesChangedModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Contains(t, result, "+10")
		assert.Contains(t, result, "-0")
	})
}
