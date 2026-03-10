package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	appcli "github.com/felipeelias/claude-statusline/internal/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPromptCommand(t *testing.T) {
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
	assert.Contains(t, string(content), `palette = "default"`)
}

func TestInitCommandAlreadyExists(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.toml")
	err := os.WriteFile(configPath, []byte("existing"), 0644)
	require.NoError(t, err)

	app := appcli.New("test")
	err = app.Run([]string{"claude-statusline", "--config", configPath, "init"})
	assert.Error(t, err, "should fail if config already exists")
}

func TestTestCommand(t *testing.T) {
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
	var stdout bytes.Buffer
	app := appcli.New("test")
	app.Writer = &stdout

	err := app.Run([]string{"claude-statusline", "themes"})
	require.NoError(t, err)

	result := stdout.String()
	assert.Contains(t, result, "current:")
	assert.Contains(t, result, "default:")
	assert.Contains(t, result, "tokyo-night:")
	assert.Contains(t, result, "gruvbox:")
	assert.Contains(t, result, "catppuccin:")
}

func TestVersionFlag(t *testing.T) {
	var stdout bytes.Buffer
	app := appcli.New("1.2.3")
	app.Writer = &stdout

	err := app.Run([]string{"claude-statusline", "--version"})
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "1.2.3")
}
