package modules

import (
	"github.com/frank-bee/claude-statusline/internal/config"
	"github.com/frank-bee/claude-statusline/internal/input"
)

// Module is the interface that all statusline modules must implement.
type Module interface {
	Name() string
	Render(data input.Data, cfg config.Config) (string, error)
}
