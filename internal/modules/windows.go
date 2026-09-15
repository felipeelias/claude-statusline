package modules

import (
	"strings"
	"time"

	"github.com/felipeelias/claude-statusline/internal/anthropic"
	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
)

// windowNames maps the API's window kinds to the short labels the status line
// shows. An unknown kind falls back to its own name, so a window Anthropic adds
// later still renders instead of disappearing.
var windowNames = map[string]string{
	"session":    "5h",
	"weekly_all": "wk",
}

// WindowsModule renders the rate-limit windows a plan meters, read from
// Anthropic rather than from the status line payload. The payload only carries
// what Claude Code was told at session start; this reflects the account.
type WindowsModule struct{}

func (WindowsModule) Name() string { return "windows" }

func (WindowsModule) Render(_ input.Data, cfg config.Config) (string, error) {
	usage, err := anthropic.Load()
	if err != nil || len(usage.Limits) == 0 {
		// Same as credits: no reading yet means render nothing, not an error.
		return "", nil //nolint:nilerr // a missing reading is not a render error
	}

	fill, empty := resolveBarChars(cfg.Windows.BarStyle, cfg.Windows.BarFill, cfg.Windows.BarEmpty)

	parts := make([]string, 0, len(usage.Limits))
	worst := 0.0

	for _, limit := range usage.Limits {
		name, ok := windowNames[limit.Kind]
		if !ok {
			name = strings.TrimPrefix(limit.Kind, "weekly_")
		}

		templateData := struct {
			Name   string
			Pct    float64
			Bar    string
			Resets string
		}{
			Name:   name,
			Pct:    limit.Percent,
			Bar:    buildBar(limit.Percent, cfg.Windows.BarWidth, fill, empty),
			Resets: formatResetsAt(limit.ResetsAt),
		}

		rendered, renderErr := renderTemplate("windows", cfg.Windows.Format, templateData)
		if renderErr != nil {
			return "", renderErr
		}

		parts = append(parts, rendered)
		worst = max(worst, limit.Percent)
	}

	result := strings.Join(parts, cfg.Windows.Separator)

	if usage.Age > staleAfter {
		result += " ⚠︎"
	}

	return wrapStyle(result, resolveThresholdStyle(worst, cfg.Windows.Thresholds, cfg.Windows.Style)), nil
}

// formatResetsAt renders when a window resets as a weekday, or as a clock time
// when that is today - a date says nothing useful about a 5-hour window.
func formatResetsAt(value string) string {
	if value == "" {
		return ""
	}

	resetsAt, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return ""
	}

	// Local time on purpose: this is read by the person at the terminal, and
	// a window that resets at 18:00 their time should say so.
	resetsAt = resetsAt.Local() //nolint:gosmopolitan // rendered for the local user

	if resetsAt.YearDay() == time.Now().YearDay() {
		return resetsAt.Format("15:04")
	}

	return resetsAt.Format("Mon")
}
