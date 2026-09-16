package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/frank-bee/claude-statusline/internal/anthropic"
	appcli "github.com/frank-bee/claude-statusline/internal/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// isolate points config and cache lookup at a throwaway HOME, so a developer's
// own ~/.config/claude-statusline/config.toml and cached usage never decide
// whether these assertions hold.
//
// CLAUDE_CONFIG_DIR is pinned empty rather than left alone: credential and
// cache lookup follow it, and these tests are usually run from inside Claude
// Code, which exports it. Left set, the suite reads the developer's real
// credentials and caches under their real account key.
func isolate(t *testing.T) {
	t.Helper()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("CLAUDE_CONFIG_DIR", "")
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("XDG_STATE_HOME", filepath.Join(home, ".local", "state"))
}

func TestPromptCommand(t *testing.T) {
	isolate(t)
	jsonInput := `{
		"model": {"display_name": "Claude Opus 4"},
		"cwd": "/tmp/test",
		"cost": {"total_cost_usd": 0.42},
		"context_window": {"used_percentage": 42.5}
	}`

	var stdout bytes.Buffer
	app := appcli.New("test")
	app.Reader = strings.NewReader(jsonInput)
	app.Writer = &stdout

	err := app.Run([]string{"claude-statusline", "prompt"})
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Claude Opus 4")
	assert.Contains(t, stdout.String(), "$0.42")
	assert.Contains(t, stdout.String(), "42%")
}

func TestDefaultAction(t *testing.T) {
	isolate(t)

	jsonInput := `{
		"model": {"display_name": "Test Model"},
		"cwd": "/tmp",
		"cost": {"total_cost_usd": 0.10},
		"context_window": {"used_percentage": 10}
	}`

	var stdout bytes.Buffer
	app := appcli.New("test")
	app.Reader = strings.NewReader(jsonInput)
	app.Writer = &stdout

	err := app.Run([]string{"claude-statusline"})
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Test Model")
}

func TestInitCommand(t *testing.T) {
	isolate(t)

	dir := t.TempDir()
	configPath := filepath.Join(dir, "claude-statusline", "config.toml")

	var stdout bytes.Buffer
	app := appcli.New("test")
	app.Writer = &stdout

	err := app.Run([]string{"claude-statusline", "--config", configPath, "init"})
	require.NoError(t, err)

	assert.Contains(t, stdout.String(), "Config created")

	content, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "format =")
	assert.Contains(t, string(content), `preset = "default"`)
}

func TestInitCommandAlreadyExists(t *testing.T) {
	isolate(t)

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")
	err := os.WriteFile(configPath, []byte("existing"), 0644)
	require.NoError(t, err)

	app := appcli.New("test")
	err = app.Run([]string{"claude-statusline", "--config", configPath, "init"})
	assert.Error(t, err, "should fail if config already exists")
}

func TestTestCommand(t *testing.T) {
	isolate(t)

	var stdout bytes.Buffer
	app := appcli.New("test")
	app.Writer = &stdout

	err := app.Run([]string{"claude-statusline", "test"})
	require.NoError(t, err)

	result := stdout.String()
	assert.Contains(t, result, "Claude Opus 4")
	assert.Contains(t, result, "$0.42")
	assert.Contains(t, result, "42%")
}

func TestThemesCommand(t *testing.T) {
	isolate(t)

	var stdout bytes.Buffer
	app := appcli.New("test")
	app.Writer = &stdout

	err := app.Run([]string{"claude-statusline", "themes"})
	require.NoError(t, err)

	result := stdout.String()
	assert.Contains(t, result, "current:")

	// Preset names
	assert.Contains(t, result, "default:")
	assert.Contains(t, result, "minimal:")
	assert.Contains(t, result, "pastel-powerline:")
	assert.Contains(t, result, "tokyo-night:")
	assert.Contains(t, result, "gruvbox-rainbow:")
	assert.Contains(t, result, "catppuccin:")

	// The modules that are off by default are previewed too, so `themes` shows
	// what enabling them looks like and not only what the presets ship with.
	assert.Contains(t, result, "optional modules")
	assert.Contains(t, result, "usage (from the Claude Code payload):")
	assert.Contains(t, result, "windows (from Anthropic):")
	assert.Contains(t, result, "credits (from Anthropic):")
	assert.Contains(t, result, "$116/$200 (58%)", "the mock credit pool should render")
	assert.Contains(t, result, "wk 63%", "the mock weekly window should render")
}

func TestThemesLeavesTheRealCacheAlone(t *testing.T) {
	isolate(t)

	cache, err := anthropic.CachePath()
	require.NoError(t, err)

	original := `{"limits":[{"kind":"session","percent":7}],"spend":{"enabled":false}}`
	require.NoError(t, os.WriteFile(cache, []byte(original), 0o600))

	before := os.Getenv("XDG_STATE_HOME")

	var stdout bytes.Buffer
	app := appcli.New("test")
	app.Writer = &stdout

	require.NoError(t, app.Run([]string{"claude-statusline", "themes"}))

	after, err := os.ReadFile(cache)
	require.NoError(t, err)
	assert.JSONEq(t, original, string(after), "previewing must not overwrite the user's reading")
	assert.Equal(t, before, os.Getenv("XDG_STATE_HOME"), "XDG_STATE_HOME must be restored")
	assert.NotContains(t, stdout.String(), "7%", "the preview shows mock data, not the real cache")
}

func TestVersionFlag(t *testing.T) {
	var stdout bytes.Buffer
	app := appcli.New("1.2.3")
	app.Writer = &stdout

	err := app.Run([]string{"claude-statusline", "--version"})
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "1.2.3")
}
