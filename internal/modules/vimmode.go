package modules

import (
	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
)

// VimModeModule renders the current vim editor mode.
type VimModeModule struct{}

func (VimModeModule) Name() string { return "vim_mode" }

func (VimModeModule) Render(data input.Data, cfg config.Config) (string, error) {
	if data.Vim == nil || data.Vim.Mode == "" {
		return "", nil
	}

	templateData := struct{ Mode string }{Mode: data.Vim.Mode}

	result, err := renderTemplate("vim_mode", cfg.VimMode.Format, templateData)
	if err != nil {
		return "", err
	}

	return wrapStyle(result, cfg.VimMode.Style), nil
}
