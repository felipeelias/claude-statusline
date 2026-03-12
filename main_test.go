package main_test

import (
	"bytes"
	"context"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testBinary string

const mockJSON = `{
	"model": {"display_name": "Claude Opus 4"},
	"cwd": "/tmp/test",
	"cost": {"total_cost_usd": 0.42},
	"context_window": {"used_percentage": 42.5}
}`

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "claude-statusline-test-*")
	if err != nil {
		log.Fatal(err)
	}

	testBinary = filepath.Join(tmp, "claude-statusline")
	build := exec.CommandContext(context.Background(), "go", "build", "-o", testBinary, ".")
	build.Stderr = os.Stderr

	err = build.Run()
	if err != nil {
		_ = os.RemoveAll(tmp)
		log.Fatalf("failed to build binary: %v", err)
	}

	code := m.Run()
	_ = os.RemoveAll(tmp)
	os.Exit(code)
}

func TestEndToEnd(t *testing.T) {
	cmd := exec.CommandContext(context.Background(), testBinary)
	cmd.Stdin = strings.NewReader(mockJSON)
	out, err := cmd.Output()
	require.NoError(t, err)

	result := string(out)
	assert.Contains(t, result, "Claude Opus 4")
	assert.Contains(t, result, "/tmp/test")
	assert.Contains(t, result, "$0.42")
	assert.Contains(t, result, "42%")
}

func TestEndToEndPromptSubcommand(t *testing.T) {
	cmd := exec.CommandContext(context.Background(), testBinary, "prompt")
	cmd.Stdin = strings.NewReader(mockJSON)
	out, err := cmd.Output()
	require.NoError(t, err)

	result := string(out)
	assert.Contains(t, result, "Claude Opus 4")
	assert.Contains(t, result, "$0.42")
}

func TestEndToEndEmptyJSON(t *testing.T) {
	cmd := exec.CommandContext(context.Background(), testBinary)
	cmd.Stdin = strings.NewReader("{}")
	out, err := cmd.Output()
	require.NoError(t, err)
	assert.NotNil(t, out)
}

func TestEndToEndInitCommand(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "claude-statusline", "config.toml")

	cmd := exec.CommandContext(context.Background(), testBinary, "--config", configPath, "init")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err := cmd.Run()
	require.NoError(t, err)

	assert.Contains(t, stdout.String(), "Config created")

	content, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), `preset = "default"`)
	assert.Contains(t, string(content), "format =")
}

func TestEndToEndTestCommand(t *testing.T) {
	cmd := exec.CommandContext(context.Background(), testBinary, "test")
	out, err := cmd.Output()
	require.NoError(t, err)

	result := string(out)
	assert.Contains(t, result, "Claude Opus 4")
	assert.Contains(t, result, "$0.42")
	assert.Contains(t, result, "42%")
}

func TestEndToEndThemesCommand(t *testing.T) {
	cmd := exec.CommandContext(context.Background(), testBinary, "themes")
	out, err := cmd.Output()
	require.NoError(t, err)

	result := string(out)
	assert.Contains(t, result, "current:")
	assert.Contains(t, result, "default:")
	assert.Contains(t, result, "minimal:")
	assert.Contains(t, result, "pastel-powerline:")
	assert.Contains(t, result, "tokyo-night:")
	assert.Contains(t, result, "gruvbox-rainbow:")
	assert.Contains(t, result, "catppuccin:")
}

func TestEndToEndWithPresetConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")
	err := os.WriteFile(configPath, []byte(`
preset = "catppuccin"
`), 0o644)
	require.NoError(t, err)

	cmd := exec.CommandContext(context.Background(), testBinary, "--config", configPath)
	cmd.Stdin = strings.NewReader(mockJSON)
	out, err := cmd.Output()
	require.NoError(t, err)

	result := string(out)
	assert.Contains(t, result, "Claude Opus 4")
	assert.Contains(t, result, "$0.42")
}

func TestEndToEndVersion(t *testing.T) {
	cmd := exec.CommandContext(context.Background(), testBinary, "--version")
	out, err := cmd.Output()
	require.NoError(t, err)
	assert.Contains(t, string(out), "dev")
}
