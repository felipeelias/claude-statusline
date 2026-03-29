package modules

import (
	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
)

// ContextModule renders the context window usage with a progress bar.
type ContextModule struct{}

func (ContextModule) Name() string { return "context" }

func (ContextModule) Render(data input.Data, cfg config.Config) (string, error) {
	pct := data.ContextWindow.UsedPercentage

	fill, empty := resolveBarChars(cfg.Context.BarStyle, cfg.Context.BarFill, cfg.Context.BarEmpty)
	bar := buildBar(pct, cfg.Context.BarWidth, fill, empty)

	templateData := struct {
		Bar     string
		UsedPct float64
	}{
		Bar:     bar,
		UsedPct: pct,
	}

	result, err := renderTemplate("context", cfg.Context.Format, templateData)
	if err != nil {
		return "", err
	}

	winningStyle := resolveThresholdStyle(pct, cfg.Context.Thresholds, cfg.Context.Style)

	return wrapStyle(result, winningStyle), nil
}
