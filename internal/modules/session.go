package modules

import (
	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
)

const (
	msPerSecond      = 1000
	secondsPerMinute = 60
	secondsPerHour   = 3600
)

// SessionTimerModule renders the session elapsed time.
type SessionTimerModule struct{}

func (SessionTimerModule) Name() string { return "session_timer" }

func (SessionTimerModule) Render(data input.Data, cfg config.Config) (string, error) {
	ms := data.Cost.TotalDurationMs
	if ms == 0 {
		return "", nil
	}

	totalSeconds := ms / msPerSecond
	hours := totalSeconds / secondsPerHour
	minutes := (totalSeconds % secondsPerHour) / secondsPerMinute
	seconds := totalSeconds % secondsPerMinute

	templateData := struct {
		Hours   int
		Minutes int
		Seconds int
	}{
		Hours:   hours,
		Minutes: minutes,
		Seconds: seconds,
	}

	result, err := renderTemplate("session_timer", cfg.SessionTimer.Format, templateData)
	if err != nil {
		return "", err
	}

	return wrapStyle(result, cfg.SessionTimer.Style), nil
}
