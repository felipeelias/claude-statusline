package modules

import (
	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
)

// ModelModule renders the AI model name.
type ModelModule struct{}

func (ModelModule) Name() string { return "model" }

func (ModelModule) Render(data input.Data, cfg config.Config) (string, error) {
	displayName := data.Model.DisplayName
	if displayName == "" {
		return "", nil
	}

	templateData := struct{ DisplayName string }{DisplayName: displayName}

	result, err := renderTemplate("model", cfg.Model.Format, templateData)
	if err != nil {
		return "", err
	}

	return wrapStyle(result, cfg.Model.Style), nil
}
