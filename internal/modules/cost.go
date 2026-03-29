package modules

import (
	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
)

const msPerHour = 3_600_000

// CostModule renders the session cost with threshold-based styling.
type CostModule struct{}

func (CostModule) Name() string { return "cost" }

func (CostModule) Render(data input.Data, cfg config.Config) (string, error) {
	cost := data.Cost.TotalCostUSD

	var burnRate float64
	if data.Cost.TotalDurationMs > 0 {
		hours := float64(data.Cost.TotalDurationMs) / msPerHour
		burnRate = cost / hours
	}

	templateData := struct {
		TotalCostUSD   float64
		BurnRate       float64
		APIDurationMs  int
	}{
		TotalCostUSD:  cost,
		BurnRate:      burnRate,
		APIDurationMs: data.Cost.TotalAPIDurationMs,
	}

	result, err := renderTemplate("cost", cfg.Cost.Format, templateData)
	if err != nil {
		return "", err
	}

	winningStyle := resolveThresholdStyle(cost, cfg.Cost.Thresholds, cfg.Cost.Style)

	return wrapStyle(result, winningStyle), nil
}
