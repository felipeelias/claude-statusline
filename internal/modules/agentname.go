package modules

import (
	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
)

// AgentNameModule renders the agent name when running with --agent.
type AgentNameModule struct{}

func (AgentNameModule) Name() string { return "agent_name" }

func (AgentNameModule) Render(data input.Data, cfg config.Config) (string, error) {
	if data.Agent == nil || data.Agent.Name == "" {
		return "", nil
	}

	templateData := struct{ Name string }{Name: data.Agent.Name}

	result, err := renderTemplate("agent_name", cfg.AgentName.Format, templateData)
	if err != nil {
		return "", err
	}

	return wrapStyle(result, cfg.AgentName.Style), nil
}
