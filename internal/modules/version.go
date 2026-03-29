package modules

import (
	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
)

// VersionModule renders the Claude Code version string.
type VersionModule struct{}

func (VersionModule) Name() string { return "version" }

func (VersionModule) Render(data input.Data, cfg config.Config) (string, error) {
	if data.Version == "" {
		return "", nil
	}

	templateData := struct{ Version string }{Version: data.Version}

	result, err := renderTemplate("version", cfg.Version.Format, templateData)
	if err != nil {
		return "", err
	}

	return wrapStyle(result, cfg.Version.Style), nil
}
