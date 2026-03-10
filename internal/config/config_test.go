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
	assert.Equal(t, "$model | $directory | $cost | $context", cfg.Format)
	assert.Equal(t, "default", cfg.Palette)
	assert.False(t, cfg.Model.Disabled)
	assert.True(t, cfg.GitBranch.Disabled)
	assert.True(t, cfg.SessionTimer.Disabled)
	assert.True(t, cfg.LinesChanged.Disabled)
	assert.Equal(t, 5, cfg.Context.BarWidth)
	assert.Equal(t, "\u2588", cfg.Context.BarFill)
	assert.Len(t, cfg.Cost.Thresholds, 2)
	assert.Len(t, cfg.Context.Thresholds, 3)
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	os.WriteFile(path, []byte(`
format = "$model | $cost"
palette = "custom"

[palettes.custom]
accent = "#ff0000"

[model]
format = "M: {{.DisplayName}}"

[git_branch]
disabled = false
`), 0o644)

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
	assert.Equal(t, "$model | $directory | $cost | $context", cfg.Format)
}

func TestResolveStyle(t *testing.T) {
	cfg := config.Default()
	assert.Equal(t, "cyan", cfg.ResolveStyle("palette:accent"))
	assert.Equal(t, "bold green", cfg.ResolveStyle("bold green"))
	assert.Equal(t, "palette:nonexistent", cfg.ResolveStyle("palette:nonexistent"))
}

func TestDefaultPath(t *testing.T) {
	path := config.DefaultPath()
	assert.Contains(t, path, ".config/claude-statusline/config.toml")
}
