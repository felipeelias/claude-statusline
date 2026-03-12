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
	assert.Equal(t, "default", cfg.Preset)
	assert.Equal(t, "$directory | $git_branch | $model | $cost | $context", cfg.Format)
	assert.Equal(t, "cyan", cfg.Directory.Style)
	assert.Equal(t, "bold", cfg.Model.Style)
	assert.False(t, cfg.Model.Disabled)
	assert.True(t, cfg.SessionTimer.Disabled)
	assert.True(t, cfg.LinesChanged.Disabled)
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	require.NoError(t, os.WriteFile(path, []byte(`
format = "$model | $cost"
[model]
style = "italic"
`), 0o644))

	cfg, err := config.Load(path)
	require.NoError(t, err)
	assert.Equal(t, "$model | $cost", cfg.Format)
	assert.Equal(t, "italic", cfg.Model.Style)
}

func TestLoadMissingFileReturnsDefaults(t *testing.T) {
	cfg, err := config.Load("/nonexistent/path/config.toml")
	require.NoError(t, err)
	assert.Equal(t, config.Default().Format, cfg.Format)
}

func TestLoadWithPreset(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	require.NoError(t, os.WriteFile(path, []byte(`
preset = "catppuccin"
`), 0o644))

	cfg, err := config.Load(path)
	require.NoError(t, err)
	assert.Equal(t, "catppuccin", cfg.Preset)
	assert.NotEqual(t, config.Default().Format, cfg.Format)
	assert.Contains(t, cfg.Format, "$directory")
}

func TestLoadWithPresetAndOverrides(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	require.NoError(t, os.WriteFile(path, []byte(`
preset = "pure"

[model]
format = "CUSTOM: {{.DisplayName}}"
`), 0o644))

	cfg, err := config.Load(path)
	require.NoError(t, err)
	assert.Equal(t, "pure", cfg.Preset)
	assert.Equal(t, "CUSTOM: {{.DisplayName}}", cfg.Model.Format)
}

func TestLoadWithUnknownPreset(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	require.NoError(t, os.WriteFile(path, []byte(`
preset = "nonexistent"
`), 0o644))

	cfg, err := config.Load(path)
	require.NoError(t, err)
	assert.Equal(t, config.Default().Format, cfg.Format)
}

func TestSampleConfig(t *testing.T) {
	sample := config.SampleConfig()
	assert.Contains(t, sample, `preset = "default"`)
	assert.Contains(t, sample, "format =")
	assert.Contains(t, sample, "minimal")
	assert.Contains(t, sample, "pastel-powerline")
	assert.Contains(t, sample, "tokyo-night")
	assert.Contains(t, sample, "gruvbox-rainbow")
	assert.Contains(t, sample, "catppuccin")
	assert.Contains(t, sample, "# [model]")
	assert.Contains(t, sample, "# [cost]")
	assert.Contains(t, sample, "# [context]")
	assert.Contains(t, sample, "# [session_timer]")
}

func TestDefaultPath(t *testing.T) {
	path := config.DefaultPath()
	assert.Contains(t, path, "claude-statusline")
	assert.Contains(t, path, "config.toml")
}
