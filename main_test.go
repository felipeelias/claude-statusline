package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	binary := filepath.Join(tmpDir, "claude-statusline")
	build := exec.CommandContext(t.Context(), "go", "build", "-o", binary, ".")
	build.Stderr = os.Stderr
	require.NoError(t, build.Run())

	jsonInput := `{
		"model": {"display_name": "Claude Opus 4"},
		"cwd": "/tmp/test",
		"cost": {"total_cost_usd": 0.42},
		"context_window": {"used_percentage": 42.5}
	}`

	cmd := exec.CommandContext(t.Context(), binary)
	cmd.Stdin = strings.NewReader(jsonInput)
	out, err := cmd.Output()
	require.NoError(t, err)

	result := string(out)
	assert.Contains(t, result, "Claude Opus 4")
	assert.Contains(t, result, "/tmp/test")
	assert.Contains(t, result, "$0.42")
	assert.Contains(t, result, "42%")
	assert.Contains(t, result, "|")
}

func TestIntegrationEmptyJSON(t *testing.T) {
	tmpDir := t.TempDir()
	binary := filepath.Join(tmpDir, "claude-statusline")
	build := exec.CommandContext(t.Context(), "go", "build", "-o", binary, ".")
	build.Stderr = os.Stderr
	require.NoError(t, build.Run())

	cmd := exec.CommandContext(t.Context(), binary)
	cmd.Stdin = strings.NewReader("{}")
	out, err := cmd.Output()
	require.NoError(t, err)
	assert.NotNil(t, out)
}
