package config_test

import (
	"testing"

	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestThemeNames(t *testing.T) {
	names := config.ThemeNames()
	assert.Equal(t, []string{"default", "minimal", "powerline", "rounded"}, names)
}

func TestApplyThemeDefault(t *testing.T) {
	cfg, ok := config.ApplyTheme("default")
	assert.True(t, ok)
	assert.Equal(t, "default", cfg.Theme)
	assert.Equal(t, config.Default().Format, cfg.Format)
}

func TestApplyThemePowerline(t *testing.T) {
	cfg, ok := config.ApplyTheme("powerline")
	assert.True(t, ok)
	assert.Equal(t, "powerline", cfg.Theme)
	assert.Contains(t, cfg.Format, "\ue0b0")
	assert.Contains(t, cfg.Format, "$directory")
	assert.Contains(t, cfg.Format, "$git_branch")
	assert.Contains(t, cfg.Format, "$model")
	assert.Contains(t, cfg.Format, "$cost")
	assert.Contains(t, cfg.Format, "$context")
	// Module formats should have padding
	assert.Contains(t, cfg.Directory.Format, " ")
	assert.Contains(t, cfg.Directory.Style, "bg:palette:dir_bg")
}

func TestApplyThemeRounded(t *testing.T) {
	cfg, ok := config.ApplyTheme("rounded")
	assert.True(t, ok)
	assert.Equal(t, "rounded", cfg.Theme)
	assert.Contains(t, cfg.Format, "\ue0b6")
	assert.Contains(t, cfg.Format, "\ue0b4")
	assert.Contains(t, cfg.Format, "$directory")
}

func TestApplyThemeMinimal(t *testing.T) {
	cfg, ok := config.ApplyTheme("minimal")
	assert.True(t, ok)
	assert.Equal(t, "minimal", cfg.Theme)
	assert.Equal(t, "$directory  $git_branch  $model  $cost  $context", cfg.Format)
	// Minimal should not have bg styles
	assert.Equal(t, "dim", cfg.Directory.Style)
	assert.Equal(t, "dim", cfg.GitBranch.Style)
	// Minimal git_branch should not have icons
	assert.NotContains(t, cfg.GitBranch.Format, "\ue0a0")
}

func TestApplyThemeUnknown(t *testing.T) {
	cfg, ok := config.ApplyTheme("nonexistent")
	assert.False(t, ok)
	assert.Equal(t, config.Default().Format, cfg.Format)
}

func TestThemesHavePalettes(t *testing.T) {
	for _, name := range config.ThemeNames() {
		cfg, ok := config.ApplyTheme(name)
		assert.True(t, ok, "theme %s should exist", name)
		assert.Len(t, cfg.Palettes, 4, "theme %s should include all palettes", name)
	}
}
