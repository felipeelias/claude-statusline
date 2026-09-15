package modules

import (
	"time"

	"github.com/felipeelias/claude-statusline/internal/anthropic"
	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
)

// staleAfter is when a cached reading stops being worth trusting silently.
// Usually an expired login: the refresh keeps failing and the numbers freeze.
const staleAfter = 30 * time.Minute

// CreditsModule renders the credit pool a usage-based seat meters, read from
// Anthropic rather than estimated from token counts. Plans without a pool
// (Pro, Max) report it disabled and the module renders nothing, which is what
// $usage is for - that one reads the rate-limit windows.
type CreditsModule struct{}

func (CreditsModule) Name() string { return "credits" }

func (CreditsModule) Render(_ input.Data, cfg config.Config) (string, error) {
	usage, err := anthropic.Load()
	if err != nil {
		// No cache yet, or no credentials. A refresh has been triggered; stay
		// quiet rather than putting an error in the middle of the status line.
		return "", nil //nolint:nilerr // a missing reading is not a render error
	}

	if !usage.Spend.Enabled {
		return "", nil
	}

	spend := usage.Spend
	fill, empty := resolveBarChars(cfg.Credits.BarStyle, cfg.Credits.BarFill, cfg.Credits.BarEmpty)

	var limit float64
	if spend.Limit != nil {
		limit = spend.Limit.Major()
	}

	stale := ""
	if usage.Age > staleAfter {
		stale = " ⚠︎"
	}

	templateData := struct {
		Used     float64
		Limit    float64
		Pct      float64
		Bar      string
		Currency string
		Stale    string
	}{
		Used:     spend.Used.Major(),
		Limit:    limit,
		Pct:      spend.Percent,
		Bar:      buildBar(spend.Percent, cfg.Credits.BarWidth, fill, empty),
		Currency: spend.Used.Currency,
		Stale:    stale,
	}

	result, err := renderTemplate("credits", cfg.Credits.Format, templateData)
	if err != nil {
		return "", err
	}

	winningStyle := resolveThresholdStyle(spend.Percent, cfg.Credits.Thresholds, cfg.Credits.Style)

	return wrapStyle(result, winningStyle), nil
}
