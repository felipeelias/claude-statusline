package modules

import (
	"fmt"
	"time"

	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
)

const (
	hoursPerDay    = 24
	minutesPerHour = 60
)

// UsageModule renders plan usage limits (5-hour block and weekly).
type UsageModule struct{}

func (UsageModule) Name() string { return "usage" }

func (UsageModule) Render(data input.Data, cfg config.Config) (string, error) {
	rateLimits := data.RateLimits
	if rateLimits == nil {
		return "", nil
	}

	blockPct := rateLimits.FiveHour.UsedPercentage
	weeklyPct := rateLimits.SevenDay.UsedPercentage

	fill, empty := resolveBarChars(cfg.Usage.BarStyle, cfg.Usage.BarFill, cfg.Usage.BarEmpty)

	templateData := struct {
		BlockPct     float64
		WeeklyPct    float64
		BlockBar     string
		WeeklyBar    string
		BlockResets  string
		WeeklyResets string
	}{
		BlockPct:     blockPct,
		WeeklyPct:    weeklyPct,
		BlockBar:     buildBar(blockPct, cfg.Usage.BarWidth, fill, empty),
		WeeklyBar:    buildBar(weeklyPct, cfg.Usage.BarWidth, fill, empty),
		BlockResets:  formatResetTimestamp(rateLimits.FiveHour.ResetsAt),
		WeeklyResets: formatResetTimestamp(rateLimits.SevenDay.ResetsAt),
	}

	result, err := renderTemplate("usage", cfg.Usage.Format, templateData)
	if err != nil {
		return "", err
	}

	winningStyle := resolveThresholdStyle(max(blockPct, weeklyPct), cfg.Usage.Thresholds, cfg.Usage.Style)

	return wrapStyle(result, winningStyle), nil
}

// formatResetTimestamp converts a Unix timestamp to a human-readable duration like "2h13m" or "3d2h".
func formatResetTimestamp(ts int64) string {
	if ts == 0 {
		return ""
	}

	remaining := time.Until(time.Unix(ts, 0))
	if remaining <= 0 {
		return "0m"
	}

	days := int(remaining.Hours()) / hoursPerDay
	hours := int(remaining.Hours()) % hoursPerDay
	minutes := int(remaining.Minutes()) % minutesPerHour

	if days > 0 {
		return fmt.Sprintf("%dd%dh", days, hours)
	}

	if hours > 0 {
		return fmt.Sprintf("%dh%02dm", hours, minutes)
	}

	return fmt.Sprintf("%dm", minutes)
}
