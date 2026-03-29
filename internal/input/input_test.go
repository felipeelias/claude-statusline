package input_test

import (
	"strings"
	"testing"

	"github.com/felipeelias/claude-statusline/internal/input"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_FullJSON(t *testing.T) {
	jsonStr := `{
		"session_id": "abc-123",
		"transcript_path": "/path/to/transcript.jsonl",
		"version": "1.0.42",
		"model": {
			"id": "claude-sonnet-4-20250514",
			"display_name": "Claude Sonnet 4"
		},
		"cwd": "/home/user/project",
		"workspace": {
			"current_dir": "/home/user/project",
			"project_dir": "/home/user/project"
		},
		"cost": {
			"total_cost_usd": 0.1234,
			"total_duration_ms": 5000,
			"total_api_duration_ms": 2300,
			"total_lines_added": 100,
			"total_lines_removed": 50
		},
		"context_window": {
			"total_input_tokens": 15234,
			"total_output_tokens": 4521,
			"used_percentage": 35.5,
			"remaining_percentage": 64.5,
			"context_window_size": 200000,
			"current_usage": {
				"input_tokens": 8500,
				"output_tokens": 1200,
				"cache_creation_input_tokens": 5000,
				"cache_read_input_tokens": 2000
			}
		},
		"exceeds_200k_tokens": true,
		"output_style": {
			"name": "default"
		},
		"vim": {
			"mode": "NORMAL"
		},
		"agent": {
			"name": "security-reviewer"
		},
		"rate_limits": {
			"five_hour": {
				"used_percentage": 12.5,
				"resets_at": 1710000000
			},
			"seven_day": {
				"used_percentage": 48.0,
				"resets_at": 1710500000
			}
		},
		"worktree": {
			"name": "my-feature",
			"path": "/path/to/.claude/worktrees/my-feature",
			"branch": "worktree-my-feature",
			"original_cwd": "/path/to/project",
			"original_branch": "main"
		}
	}`

	data, err := input.Parse(strings.NewReader(jsonStr))
	require.NoError(t, err)

	assert.Equal(t, "abc-123", data.SessionID)
	assert.Equal(t, "/path/to/transcript.jsonl", data.TranscriptPath)
	assert.Equal(t, "1.0.42", data.Version)

	assert.Equal(t, "claude-sonnet-4-20250514", data.Model.ID)
	assert.Equal(t, "Claude Sonnet 4", data.Model.DisplayName)

	assert.Equal(t, "/home/user/project", data.Cwd)
	assert.Equal(t, "/home/user/project", data.Workspace.CurrentDir)
	assert.Equal(t, "/home/user/project", data.Workspace.ProjectDir)

	assert.InDelta(t, 0.1234, data.Cost.TotalCostUSD, 0.0001)
	assert.Equal(t, 5000, data.Cost.TotalDurationMs)
	assert.Equal(t, 2300, data.Cost.TotalAPIDurationMs)
	assert.Equal(t, 100, data.Cost.TotalLinesAdded)
	assert.Equal(t, 50, data.Cost.TotalLinesRemoved)

	assert.Equal(t, 15234, data.ContextWindow.TotalInputTokens)
	assert.Equal(t, 4521, data.ContextWindow.TotalOutputTokens)
	assert.InDelta(t, 35.5, data.ContextWindow.UsedPercentage, 0.01)
	assert.InDelta(t, 64.5, data.ContextWindow.RemainingPercentage, 0.01)
	assert.Equal(t, 200000, data.ContextWindow.ContextWindowSize)
	require.NotNil(t, data.ContextWindow.CurrentUsage)
	assert.Equal(t, 8500, data.ContextWindow.CurrentUsage.InputTokens)
	assert.Equal(t, 1200, data.ContextWindow.CurrentUsage.OutputTokens)
	assert.Equal(t, 5000, data.ContextWindow.CurrentUsage.CacheCreationInputTokens)
	assert.Equal(t, 2000, data.ContextWindow.CurrentUsage.CacheReadInputTokens)

	assert.True(t, data.Exceeds200kTokens)

	assert.Equal(t, "default", data.OutputStyle.Name)

	require.NotNil(t, data.Vim)
	assert.Equal(t, "NORMAL", data.Vim.Mode)

	require.NotNil(t, data.Agent)
	assert.Equal(t, "security-reviewer", data.Agent.Name)

	require.NotNil(t, data.RateLimits)
	assert.InDelta(t, 12.5, data.RateLimits.FiveHour.UsedPercentage, 0.01)
	assert.Equal(t, int64(1710000000), data.RateLimits.FiveHour.ResetsAt)
	assert.InDelta(t, 48.0, data.RateLimits.SevenDay.UsedPercentage, 0.01)
	assert.Equal(t, int64(1710500000), data.RateLimits.SevenDay.ResetsAt)

	require.NotNil(t, data.Worktree)
	assert.Equal(t, "my-feature", data.Worktree.Name)
	assert.Equal(t, "/path/to/.claude/worktrees/my-feature", data.Worktree.Path)
	assert.Equal(t, "worktree-my-feature", data.Worktree.Branch)
	assert.Equal(t, "/path/to/project", data.Worktree.OriginalCwd)
	assert.Equal(t, "main", data.Worktree.OriginalBranch)
}

func TestParse_EmptyJSON(t *testing.T) {
	data, err := input.Parse(strings.NewReader("{}"))
	require.NoError(t, err)

	assert.Empty(t, data.SessionID)
	assert.Empty(t, data.TranscriptPath)
	assert.Empty(t, data.Version)
	assert.Empty(t, data.Model.ID)
	assert.Empty(t, data.Model.DisplayName)
	assert.Empty(t, data.Cwd)
	assert.Empty(t, data.Workspace.CurrentDir)
	assert.Empty(t, data.Workspace.ProjectDir)
	assert.InDelta(t, 0.0, data.Cost.TotalCostUSD, 0.0001)
	assert.Equal(t, 0, data.Cost.TotalDurationMs)
	assert.Equal(t, 0, data.Cost.TotalAPIDurationMs)
	assert.Equal(t, 0, data.Cost.TotalLinesAdded)
	assert.Equal(t, 0, data.Cost.TotalLinesRemoved)
	assert.Equal(t, 0, data.ContextWindow.TotalInputTokens)
	assert.Equal(t, 0, data.ContextWindow.TotalOutputTokens)
	assert.InDelta(t, 0.0, data.ContextWindow.UsedPercentage, 0.01)
	assert.InDelta(t, 0.0, data.ContextWindow.RemainingPercentage, 0.01)
	assert.Equal(t, 0, data.ContextWindow.ContextWindowSize)
	assert.Nil(t, data.ContextWindow.CurrentUsage)
	assert.False(t, data.Exceeds200kTokens)
	assert.Empty(t, data.OutputStyle.Name)
	assert.Nil(t, data.Vim)
	assert.Nil(t, data.Agent)
	assert.Nil(t, data.Worktree)
	assert.Nil(t, data.RateLimits)
}

func TestParse_InvalidJSON(t *testing.T) {
	_, err := input.Parse(strings.NewReader("not json"))
	assert.Error(t, err)
}
