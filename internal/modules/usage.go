package modules

import (
	"fmt"
	"math"
	"time"

	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
	"github.com/felipeelias/claude-statusline/internal/usage"
)

// UsageModule renders plan usage limits (5-hour block and weekly).
type UsageModule struct {
	fetcher usage.Fetcher
}

// NewUsageModule creates a UsageModule that fetches real data via the OAuth API.
func NewUsageModule() UsageModule {
	return UsageModule{fetcher: usage.DefaultFetcher{}}
}

// NewUsageModuleWithFetcher creates a UsageModule with a custom fetcher for testing.
func NewUsageModuleWithFetcher(f usage.Fetcher) UsageModule {
	return UsageModule{fetcher: f}
}

func (UsageModule) Name() string { return "usage" }

func (m UsageModule) Render(_ input.Data, cfg config.Config) (string, error) {
	if cfg.Usage.TestMode {
		return m.renderData(mockUsageData(), cfg)
	}

	data, err := m.fetcher.GetUsage(time.Duration(cfg.Usage.CacheTTLSeconds) * time.Second)
	if err != nil || data == nil {
		return "", nil
	}

	return m.renderData(data, cfg)
}

func (m UsageModule) renderData(data *usage.UsageData, cfg config.Config) (string, error) {
	blockPct := data.FiveHour.Utilisation
	weeklyPct := data.SevenDay.Utilisation

	templateData := struct {
		BlockPct    float64
		WeeklyPct   float64
		BlockBar    string
		WeeklyBar   string
		BlockResets string
		WeeklyResets string
	}{
		BlockPct:     blockPct,
		WeeklyPct:    weeklyPct,
		BlockBar:     buildBar(blockPct, cfg.Usage.BarWidth, cfg.Usage.BarFill, cfg.Usage.BarEmpty),
		WeeklyBar:    buildBar(weeklyPct, cfg.Usage.BarWidth, cfg.Usage.BarFill, cfg.Usage.BarEmpty),
		BlockResets:  formatResetTime(data.FiveHour.ResetsAt),
		WeeklyResets: formatResetTime(data.SevenDay.ResetsAt),
	}

	result, err := renderTemplate("usage", cfg.Usage.Format, templateData)
	if err != nil {
		return "", err
	}

	winningStyle := resolveThresholdStyle(blockPct, cfg.Usage.Thresholds, cfg.Usage.Style)

	return wrapStyle(result, winningStyle), nil
}

// formatResetTime converts an RFC3339 reset time to a human-readable duration like "2h13m" or "3d2h".
func formatResetTime(resetAt string) string {
	if resetAt == "" {
		return ""
	}

	t, err := time.Parse(time.RFC3339Nano, resetAt)
	if err != nil {
		t, err = time.Parse(time.RFC3339, resetAt)
		if err != nil {
			return ""
		}
	}

	d := time.Until(t)
	if d <= 0 {
		return "0m"
	}

	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(math.Mod(d.Minutes(), 60))

	if days > 0 {
		return fmt.Sprintf("%dd%dh", days, hours)
	}

	if hours > 0 {
		return fmt.Sprintf("%dh%02dm", hours, minutes)
	}

	return fmt.Sprintf("%dm", minutes)
}

//nolint:mnd // mock data uses literal values by design
func mockUsageData() *usage.UsageData {
	return &usage.UsageData{
		FiveHour: usage.Window{
			Utilisation: 42,
			ResetsAt:    time.Now().Add(2*time.Hour + 13*time.Minute).Format(time.RFC3339),
		},
		SevenDay: usage.Window{
			Utilisation: 15,
			ResetsAt:    time.Now().Add(3*24*time.Hour + 2*time.Hour).Format(time.RFC3339),
		},
	}
}
