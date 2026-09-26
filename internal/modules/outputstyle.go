package modules

import (
	"strings"

	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
)

// defaultOutputStyleName is the built-in style Claude Code reports when the
// user has not chosen one.
const defaultOutputStyleName = "default"

// OutputStyleModule renders Claude Code's current output style.
type OutputStyleModule struct{}

func (OutputStyleModule) Name() string { return "output_style" }

func (OutputStyleModule) Render(data input.Data, cfg config.Config) (string, error) {
	name := data.OutputStyle.Name
	if name == "" {
		return "", nil
	}

	if cfg.OutputStyle.HideDefault && strings.EqualFold(name, defaultOutputStyleName) {
		return "", nil
	}

	templateData := struct{ Name string }{Name: name}

	result, err := renderTemplate("output_style", cfg.OutputStyle.Format, templateData)
	if err != nil {
		return "", err
	}

	return wrapStyle(result, cfg.OutputStyle.Style), nil
}
