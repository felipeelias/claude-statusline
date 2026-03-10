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

// Data represents the JSON input piped from Claude Code via stdin.
type Data struct {
	SessionID     string        `json:"session_id"`
	Version       string        `json:"version"`
	Model         Model         `json:"model"`
	Cwd           string        `json:"cwd"`
	Cost          Cost          `json:"cost"`
	ContextWindow ContextWindow `json:"context_window"`
}

// Parse decodes JSON from the given reader into a Data struct.
func Parse(r io.Reader) (Data, error) {
	var data Data
	if err := json.NewDecoder(r).Decode(&data); err != nil {
		return Data{}, err
	}
	return data, nil
}
