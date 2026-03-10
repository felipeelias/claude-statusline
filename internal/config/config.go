package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config holds the full statusline configuration.
type Config struct {
	Format       string                       `toml:"format"`
	Palette      string                       `toml:"palette"`
	Palettes     map[string]map[string]string `toml:"palettes"`
	Model        ModelConfig                  `toml:"model"`
	Directory    DirectoryConfig              `toml:"directory"`
	Cost         CostConfig                   `toml:"cost"`
	Context      ContextConfig                `toml:"context"`
	GitBranch    GitBranchConfig              `toml:"git_branch"`
	SessionTimer SessionTimerConfig           `toml:"session_timer"`
	LinesChanged LinesChangedConfig           `toml:"lines_changed"`
}

// Threshold defines a conditional style based on a numeric value.
type Threshold struct {
	Above float64 `toml:"above"`
	Style string  `toml:"style"`
}

// ModelConfig holds model module settings.
type ModelConfig struct {
	Format   string `toml:"format"`
	Style    string `toml:"style"`
	Disabled bool   `toml:"disabled"`
}

// DirectoryConfig holds directory module settings.
type DirectoryConfig struct {
	Format           string `toml:"format"`
	Style            string `toml:"style"`
	Disabled         bool   `toml:"disabled"`
	TruncationLength int    `toml:"truncation_length"`
}

// CostConfig holds cost module settings.
type CostConfig struct {
	Format     string      `toml:"format"`
	Style      string      `toml:"style"`
	Disabled   bool        `toml:"disabled"`
	Thresholds []Threshold `toml:"thresholds"`
}

// ContextConfig holds context module settings.
type ContextConfig struct {
	Format     string      `toml:"format"`
	Style      string      `toml:"style"`
	Disabled   bool        `toml:"disabled"`
	BarWidth   int         `toml:"bar_width"`
	BarFill    string      `toml:"bar_fill"`
	BarEmpty   string      `toml:"bar_empty"`
	Thresholds []Threshold `toml:"thresholds"`
}

// GitBranchConfig holds git branch module settings.
type GitBranchConfig struct {
	Format   string `toml:"format"`
	Style    string `toml:"style"`
	Disabled bool   `toml:"disabled"`
}

// SessionTimerConfig holds session timer module settings.
type SessionTimerConfig struct {
	Format   string `toml:"format"`
	Style    string `toml:"style"`
	Disabled bool   `toml:"disabled"`
}

// LinesChangedConfig holds lines changed module settings.
type LinesChangedConfig struct {
	Format       string `toml:"format"`
	AddedStyle   string `toml:"added_style"`
	RemovedStyle string `toml:"removed_style"`
	Disabled     bool   `toml:"disabled"`
}

// Default returns a Config with hardcoded default values.
func Default() Config {
	return Config{
		Format:  "$model | $directory | $cost | $context",
		Palette: "default",
		Palettes: map[string]map[string]string{
			"default": {
				"accent":   "cyan",
				"cost_ok":  "green",
				"cost_warn": "yellow",
				"cost_high": "red",
				"ctx_ok":   "green",
				"ctx_warn": "yellow",
				"ctx_high": "red",
			},
		},
		Model: ModelConfig{
			Format: "{{.DisplayName}}",
			Style:  "bold",
		},
		Directory: DirectoryConfig{
			Format:           "{{.Dir}}",
			Style:            "palette:accent",
			TruncationLength: 3,
		},
		Cost: CostConfig{
			Format: `${{printf "%.2f" .TotalCostUSD}}`,
			Style:  "palette:cost_ok",
			Thresholds: []Threshold{
				{Above: 1.0, Style: "palette:cost_warn"},
				{Above: 5.0, Style: "palette:cost_high"},
			},
		},
		Context: ContextConfig{
			Format:   `{{.Bar}} {{printf "%.0f" .UsedPct}}%`,
			Style:    "palette:ctx_ok",
			BarWidth: 5,
			BarFill:  "\u2588",
			BarEmpty: "\u2591",
			Thresholds: []Threshold{
				{Above: 50, Style: "palette:ctx_warn"},
				{Above: 70, Style: "208"},
				{Above: 90, Style: "palette:ctx_high"},
			},
		},
		GitBranch: GitBranchConfig{
			Format:   "\ue0a0 {{.Branch}}",
			Style:    "palette:accent",
			Disabled: true,
		},
		SessionTimer: SessionTimerConfig{
			Format:   "{{.Elapsed}}",
			Style:    "dim",
			Disabled: true,
		},
		LinesChanged: LinesChangedConfig{
			Format:       "+{{.Added}} -{{.Removed}}",
			AddedStyle:   "green",
			RemovedStyle: "red",
			Disabled:     true,
		},
	}
}

// Load reads a TOML config file and merges it with defaults.
// If the file does not exist, Default() is returned with no error.
// If the file exists but has parse errors, an error is returned.
func Load(path string) (Config, error) {
	cfg := Default()

	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return Config{}, err
	}

	_, err = toml.DecodeFile(path, &cfg)
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// DefaultPath returns the default config file path: ~/.config/claude-statusline/config.toml
func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "claude-statusline", "config.toml")
}

// ResolveStyle resolves palette references in a style string.
// If s starts with "palette:", the key after the prefix is looked up in
// the active palette. If found, the palette value is returned.
// Otherwise s is returned unchanged.
func (c Config) ResolveStyle(s string) string {
	if !strings.HasPrefix(s, "palette:") {
		return s
	}

	key := s[len("palette:"):]

	palette, ok := c.Palettes[c.Palette]
	if !ok {
		return s
	}

	value, ok := palette[key]
	if !ok {
		return s
	}

	return value
}
