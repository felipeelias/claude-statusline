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

// Workspace represents the working directory context.
type Workspace struct {
	CurrentDir string `json:"current_dir"`
	ProjectDir string `json:"project_dir"`
}

// Cost represents usage cost and activity metrics.
type Cost struct {
	TotalCostUSD        float64 `json:"total_cost_usd"`
	TotalDurationMs     int     `json:"total_duration_ms"`
	TotalAPIDurationMs  int     `json:"total_api_duration_ms"`
	TotalLinesAdded     int     `json:"total_lines_added"`
	TotalLinesRemoved   int     `json:"total_lines_removed"`
}

// CurrentUsage represents token usage for the current API call.
type CurrentUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
}

// ContextWindow represents context window usage.
type ContextWindow struct {
	TotalInputTokens    int           `json:"total_input_tokens"`
	TotalOutputTokens   int           `json:"total_output_tokens"`
	UsedPercentage      float64       `json:"used_percentage"`
	RemainingPercentage float64       `json:"remaining_percentage"`
	ContextWindowSize   int           `json:"context_window_size"`
	CurrentUsage        *CurrentUsage `json:"current_usage"`
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

// OutputStyle represents Claude Code's current output style.
type OutputStyle struct {
	Name string `json:"name"`
}

// Vim represents vim mode state when vim mode is enabled.
type Vim struct {
	Mode string `json:"mode"`
}

// Agent represents the active agent when running with --agent.
type Agent struct {
	Name string `json:"name"`
}

// Worktree represents active worktree details.
type Worktree struct {
	Name           string `json:"name"`
	Path           string `json:"path"`
	Branch         string `json:"branch"`
	OriginalCwd    string `json:"original_cwd"`
	OriginalBranch string `json:"original_branch"`
}

// Data represents the JSON input piped from Claude Code via stdin.
type Data struct {
	SessionID         string        `json:"session_id"`
	TranscriptPath    string        `json:"transcript_path"`
	Version           string        `json:"version"`
	Model             Model         `json:"model"`
	Cwd               string        `json:"cwd"`
	Workspace         Workspace     `json:"workspace"`
	Cost              Cost          `json:"cost"`
	ContextWindow     ContextWindow `json:"context_window"`
	Exceeds200kTokens bool          `json:"exceeds_200k_tokens"`
	RateLimits        *RateLimits   `json:"rate_limits"`
	OutputStyle       OutputStyle   `json:"output_style"`
	Vim               *Vim          `json:"vim"`
	Agent             *Agent        `json:"agent"`
	Worktree          *Worktree     `json:"worktree"`
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
