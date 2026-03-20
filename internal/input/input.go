package input

import (
	"encoding/json"
	"io"
)

// Model represents the AI model being used.
type Model struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

// Cost represents usage cost and activity metrics.
type Cost struct {
	TotalCostUSD      float64 `json:"total_cost_usd"`
	TotalDurationMs   int     `json:"total_duration_ms"`
	TotalLinesAdded   int     `json:"total_lines_added"`
	TotalLinesRemoved int     `json:"total_lines_removed"`
}

// ContextWindow represents context window usage.
type ContextWindow struct {
	UsedPercentage      float64 `json:"used_percentage"`
	RemainingPercentage float64 `json:"remaining_percentage"`
	ContextWindowSize   int     `json:"context_window_size"`
}

// RateLimitWindow represents a single rate limit window (5-hour or 7-day).
type RateLimitWindow struct {
	UsedPercentage float64 `json:"used_percentage"`
	ResetsAt       int64   `json:"resets_at"`
}

// RateLimits represents Claude plan usage limits.
type RateLimits struct {
	FiveHour RateLimitWindow `json:"five_hour"`
	SevenDay RateLimitWindow `json:"seven_day"`
}

// Worktree represents active worktree details.
type Worktree struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Branch string `json:"branch"`
}

// Data represents the JSON input piped from Claude Code via stdin.
type Data struct {
	SessionID     string        `json:"session_id"`
	Version       string        `json:"version"`
	Model         Model         `json:"model"`
	Cwd           string        `json:"cwd"`
	Cost          Cost          `json:"cost"`
	ContextWindow ContextWindow `json:"context_window"`
	RateLimits    *RateLimits   `json:"rate_limits"`
	Worktree      *Worktree     `json:"worktree"`
}

// Parse decodes JSON from the given reader into a Data struct.
func Parse(reader io.Reader) (Data, error) {
	var data Data

	err := json.NewDecoder(reader).Decode(&data)
	if err != nil {
		return Data{}, err
	}

	return data, nil
}
