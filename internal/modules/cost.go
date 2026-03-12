package modules

import (
	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
)

// CostModule renders the session cost with threshold-based styling.
type CostModule struct{}

func (CostModule) Name() string { return "cost" }

func (CostModule) Render(data input.Data, cfg config.Config) (string, error) {
	cost := data.Cost.TotalCostUSD

	templateData := struct{ TotalCostUSD float64 }{TotalCostUSD: cost}

	result, err := renderTemplate("cost", cfg.Cost.Format, templateData)
	if err != nil {
		return "", err
	}

	winningStyle := resolveThresholdStyle(cost, cfg.Cost.Thresholds, cfg.Cost.Style)

	return wrapStyle(result, winningStyle), nil
}
