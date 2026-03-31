package modules_test

import (
	"testing"

	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
	"github.com/felipeelias/claude-statusline/internal/modules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentNameModule_Name(t *testing.T) {
	m := modules.AgentNameModule{}
	assert.Equal(t, "agent_name", m.Name())
}

func TestAgentNameModule_Render(t *testing.T) {
	cfg := config.Default()

	t.Run("renders agent name with default format", func(t *testing.T) {
		data := input.Data{Agent: &input.Agent{Name: "security-reviewer"}}

		result, err := modules.AgentNameModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Contains(t, result, "security-reviewer")
	})

	t.Run("nil agent renders empty", func(t *testing.T) {
		data := input.Data{Agent: nil}

		result, err := modules.AgentNameModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("empty agent name renders empty", func(t *testing.T) {
		data := input.Data{Agent: &input.Agent{Name: ""}}

		result, err := modules.AgentNameModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("applies style", func(t *testing.T) {
		data := input.Data{Agent: &input.Agent{Name: "code-reviewer"}}

		result, err := modules.AgentNameModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Contains(t, result, "\033[1;35m") // bold magenta
		assert.Contains(t, result, "\033[0m")    // reset
	})

	t.Run("custom format", func(t *testing.T) {
		customCfg := cfg
		customCfg.AgentName.Format = "agent:{{.Name}}"

		data := input.Data{Agent: &input.Agent{Name: "security-reviewer"}}

		result, err := modules.AgentNameModule{}.Render(data, customCfg)
		require.NoError(t, err)
		assert.Contains(t, result, "agent:security-reviewer")
	})
}
