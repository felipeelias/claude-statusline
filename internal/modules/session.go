package modules

import (
	"fmt"

	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
)

// SessionTimerModule renders the session elapsed time.
type SessionTimerModule struct{}

func (SessionTimerModule) Name() string { return "session_timer" }

func (SessionTimerModule) Render(data input.Data, cfg config.Config) (string, error) {
	ms := data.Cost.TotalDurationMs
	if ms == 0 {
		return "", nil
	}

	elapsed := formatDuration(ms)

	templateData := struct{ Elapsed string }{Elapsed: elapsed}

	result, err := renderTemplate("session_timer", cfg.SessionTimer.Format, templateData)
	if err != nil {
		return "", err
	}

	return wrapStyle(result, cfg.SessionTimer.Style, cfg), nil
}

// formatDuration converts milliseconds to a human-readable duration.
// If >= 1 hour: H:MM:SS (e.g. 1:05:03), else: M:SS (e.g. 5:03).
func formatDuration(ms int) string {
	totalSeconds := ms / 1000
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}
