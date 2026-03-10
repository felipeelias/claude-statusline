package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefault(t *testing.T) {
	cfg := config.Default()
	assert.Equal(t, "default", cfg.Theme)
	assert.Equal(t, "$directory | $git_branch | $model | $cost | $context", cfg.Format)
	assert.Equal(t, "default", cfg.Palette)
	assert.False(t, cfg.Model.Disabled)
	assert.False(t, cfg.GitBranch.Disabled)
	assert.True(t, cfg.SessionTimer.Disabled)
	assert.True(t, cfg.LinesChanged.Disabled)
	assert.Equal(t, 5, cfg.Context.BarWidth)
	assert.Equal(t, "\u2588", cfg.Context.BarFill)
	assert.Len(t, cfg.Cost.Thresholds, 2)
	assert.Len(t, cfg.Context.Thresholds, 3)
}

func TestDefaultPalettes(t *testing.T) {
	cfg := config.Default()
	expected := []string{"default", "tokyo-night", "gruvbox", "catppuccin"}
	for _, name := range expected {
		palette, ok := cfg.Palettes[name]
		assert.True(t, ok, "missing built-in palette: %s", name)
		assert.Len(t, palette, 13, "palette %s should have 13 keys", name)
	}
	assert.Len(t, cfg.Palettes, len(expected))
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	require.NoError(t, os.WriteFile(path, []byte(`
format = "$model | $cost"
palette = "custom"

[palettes.custom]
accent = "#ff0000"

[model]
format = "M: {{.DisplayName}}"

[git_branch]
disabled = false
`), 0o644))

	cfg, err := config.Load(path)
	require.NoError(t, err)
	assert.Equal(t, "$model | $cost", cfg.Format)
	assert.Equal(t, "custom", cfg.Palette)
	assert.Equal(t, "M: {{.DisplayName}}", cfg.Model.Format)
	assert.False(t, cfg.GitBranch.Disabled)
	// Non-overridden fields keep defaults
	assert.Equal(t, "bold", cfg.Model.Style)
}

func TestLoadMissingFileReturnsDefaults(t *testing.T) {
	cfg, err := config.Load("/nonexistent/path/config.toml")
	require.NoError(t, err)
	assert.Equal(t, "$directory | $git_branch | $model | $cost | $context", cfg.Format)
}

func TestResolveStyle(t *testing.T) {
	cfg := config.Default()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"bare palette ref", "palette:accent", "cyan"},
		{"no palette ref", "bold green", "bold green"},
		{"missing key", "palette:nonexistent", "palette:nonexistent"},
		{"fg palette ref", "fg:palette:accent", "fg:cyan"},
		{"bg palette ref", "bg:palette:accent", "bg:cyan"},
		{"compound style", "fg:palette:accent bg:palette:cost_ok bold", "fg:cyan bg:green bold"},
		{"mixed tokens", "bold palette:accent italic", "bold cyan italic"},
		{"no palette fast path", "bold italic dim", "bold italic dim"},
		{"fg non-palette", "fg:#ff0000 bold", "fg:#ff0000 bold"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, cfg.ResolveStyle(tt.input))
		})
	}
}

func TestLoadWithTheme(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	require.NoError(t, os.WriteFile(path, []byte(`
theme = "powerline"
palette = "catppuccin"
`), 0o644))

	cfg, err := config.Load(path)
	require.NoError(t, err)
	assert.Equal(t, "powerline", cfg.Theme)
	assert.Equal(t, "catppuccin", cfg.Palette)
	// Format should come from the powerline theme, not the default
	assert.NotEqual(t, "$directory | $git_branch | $model | $cost | $context", cfg.Format)
	assert.Contains(t, cfg.Format, "$directory")
}

func TestLoadWithThemeAndOverrides(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	require.NoError(t, os.WriteFile(path, []byte(`
theme = "minimal"
palette = "gruvbox"

[model]
format = "CUSTOM: {{.DisplayName}}"
`), 0o644))

	cfg, err := config.Load(path)
	require.NoError(t, err)
	assert.Equal(t, "minimal", cfg.Theme)
	assert.Equal(t, "gruvbox", cfg.Palette)
	// User override should take precedence over theme's module config
	assert.Equal(t, "CUSTOM: {{.DisplayName}}", cfg.Model.Format)
}

func TestLoadWithUnknownTheme(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	require.NoError(t, os.WriteFile(path, []byte(`
theme = "nonexistent"
`), 0o644))

	cfg, err := config.Load(path)
	require.NoError(t, err)
	// Unknown theme falls back to default
	assert.Equal(t, "$directory | $git_branch | $model | $cost | $context", cfg.Format)
}

func TestSampleConfig(t *testing.T) {
	sample := config.SampleConfig()
	assert.Contains(t, sample, `theme = "default"`)
	assert.Contains(t, sample, "format =")
	assert.Contains(t, sample, `palette = "default"`)
	assert.Contains(t, sample, "powerline")
	assert.Contains(t, sample, "rounded")
	assert.Contains(t, sample, "minimal")
	assert.Contains(t, sample, "seg_fg")
	assert.Contains(t, sample, "dir_bg")
	assert.Contains(t, sample, "# [model]")
	assert.Contains(t, sample, "# [cost]")
	assert.Contains(t, sample, "# [context]")
	assert.Contains(t, sample, "# [session_timer]")
}

func TestDefaultPath(t *testing.T) {
	path := config.DefaultPath()
	assert.Contains(t, path, ".config/claude-statusline/config.toml")
}
