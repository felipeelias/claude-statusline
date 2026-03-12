package config_test

import (
	"testing"

	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestPresetNames(t *testing.T) {
	names := config.PresetNames()
	assert.Equal(t, []string{
		"catppuccin", "default", "gruvbox-rainbow", "minimal",
		"pastel-powerline", "tokyo-night",
	}, names)
}

func TestApplyPresetDefault(t *testing.T) {
	cfg, ok := config.ApplyPreset("default")
	assert.True(t, ok)
	assert.Equal(t, "default", cfg.Preset)
	assert.Equal(t, config.Default().Format, cfg.Format)
}

func TestApplyPresetMinimal(t *testing.T) {
	cfg, ok := config.ApplyPreset("minimal")
	assert.True(t, ok)
	assert.Equal(t, "minimal", cfg.Preset)
	assert.Equal(t, "blue", cfg.Directory.Style)
	assert.NotContains(t, cfg.Format, "\ue0b0")
	assert.NotContains(t, cfg.Format, "|")
}

func TestApplyPresetCapsulePowerline(t *testing.T) {
	for _, name := range []string{"pastel-powerline", "gruvbox-rainbow", "catppuccin"} {
		t.Run(name, func(t *testing.T) {
			cfg, ok := config.ApplyPreset(name)
			assert.True(t, ok)
			assert.Equal(t, name, cfg.Preset)
			// Left half-circle cap
			assert.Contains(t, cfg.Format, "\ue0b6")
			// Arrow transitions
			assert.Contains(t, cfg.Format, "\ue0b0")
			assert.Contains(t, cfg.Format, "$directory")
			assert.Contains(t, cfg.Format, "$git_branch")
			assert.Contains(t, cfg.Format, "$model")
			assert.Contains(t, cfg.Format, "$cost")
			assert.Contains(t, cfg.Format, "$context")
			assert.Contains(t, cfg.Directory.Format, " ")
			assert.Contains(t, cfg.Directory.Style, "bg:")
		})
	}
}

func TestApplyPresetPastelTrailingArrow(t *testing.T) {
	cfg, _ := config.ApplyPreset("pastel-powerline")
	// Pastel Powerline ends with right arrow (not rounded half-circle), last color is dark blue
	assert.Contains(t, cfg.Format, "\ue0b0 ](fg:#33658A)")
}

func TestApplyPresetGruvboxTrailingRounded(t *testing.T) {
	cfg, _ := config.ApplyPreset("gruvbox-rainbow")
	// Gruvbox ends with rounded right half-circle
	assert.Contains(t, cfg.Format, "\ue0b4 ](fg:#3c3836)")
}

func TestApplyPresetTokyoNight(t *testing.T) {
	cfg, ok := config.ApplyPreset("tokyo-night")
	assert.True(t, ok)
	assert.Equal(t, "tokyo-night", cfg.Preset)
	// Gradient leading
	assert.Contains(t, cfg.Format, "░▒▓")
	// All rounded half-circle transitions (not arrows)
	assert.Contains(t, cfg.Format, "\ue0b4")
	assert.NotContains(t, cfg.Format, "\ue0b0")
	assert.Contains(t, cfg.Format, "$directory")
	assert.Contains(t, cfg.Directory.Style, "bg:")
}

func TestApplyPresetUnknown(t *testing.T) {
	cfg, ok := config.ApplyPreset("nonexistent")
	assert.False(t, ok)
	assert.Equal(t, config.Default().Format, cfg.Format)
}
