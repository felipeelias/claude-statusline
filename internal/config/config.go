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

const (
	defaultTruncationLength = 3
	defaultBarWidth         = 5
	costWarnThreshold       = 5.0
	ctxWarnThreshold        = 50
	ctxMedThreshold         = 70
	ctxHighThreshold        = 90
)

// defaultPalettes returns all built-in palette definitions.
func defaultPalettes() map[string]map[string]string {
	return map[string]map[string]string{
		"default": {
			"accent":    "cyan",
			"cost_ok":   "green",
			"cost_warn": "yellow",
			"cost_high": "red",
			"ctx_ok":    "green",
			"ctx_warn":  "yellow",
			"ctx_high":  "red",
		},
		"tokyo-night": {
			"accent":    "#769ff0",
			"cost_ok":   "#73daca",
			"cost_warn": "#e0af68",
			"cost_high": "#f7768e",
			"ctx_ok":    "#73daca",
			"ctx_warn":  "#e0af68",
			"ctx_high":  "#f7768e",
		},
		"gruvbox": {
			"accent":    "#83a598",
			"cost_ok":   "#b8bb26",
			"cost_warn": "#fabd2f",
			"cost_high": "#fb4934",
			"ctx_ok":    "#b8bb26",
			"ctx_warn":  "#fabd2f",
			"ctx_high":  "#fb4934",
		},
		"catppuccin": {
			"accent":    "#89b4fa",
			"cost_ok":   "#a6e3a1",
			"cost_warn": "#f9e2af",
			"cost_high": "#f38ba8",
			"ctx_ok":    "#a6e3a1",
			"ctx_warn":  "#f9e2af",
			"ctx_high":  "#f38ba8",
		},
	}
}

// Default returns a Config with hardcoded default values.
func Default() Config {
	return Config{
		Format:   "$directory | $git_branch | $model | $cost | $context",
		Palette:  "default",
		Palettes: defaultPalettes(),
		Model: ModelConfig{
			Format: "{{.DisplayName}}",
			Style:  "bold",
		},
		Directory: DirectoryConfig{
			Format:           "{{.Dir}}",
			Style:            "palette:accent",
			TruncationLength: defaultTruncationLength,
		},
		Cost: CostConfig{
			Format: `${{printf "%.2f" .TotalCostUSD}}`,
			Style:  "palette:cost_ok",
			Thresholds: []Threshold{
				{Above: 1.0, Style: "palette:cost_warn"},
				{Above: costWarnThreshold, Style: "palette:cost_high"},
			},
		},
		Context: ContextConfig{
			Format:   `{{.Bar}} {{printf "%.0f" .UsedPct}}%`,
			Style:    "palette:ctx_ok",
			BarWidth: defaultBarWidth,
			BarFill:  "\u2588",
			BarEmpty: "\u2591",
			Thresholds: []Threshold{
				{Above: ctxWarnThreshold, Style: "palette:ctx_warn"},
				{Above: ctxMedThreshold, Style: "208"},
				{Above: ctxHighThreshold, Style: "palette:ctx_high"},
			},
		},
		GitBranch: GitBranchConfig{
			Format: "\ue0a0 {{.Branch}}{{if .InWorktree}} \uf0e8{{end}}",
			Style:  "palette:accent",
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

// DefaultPath returns the default config file path: ~/.config/claude-statusline/config.toml.
func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return filepath.Join(home, ".config", "claude-statusline", "config.toml")
}

// sampleConfigTemplate is the commented TOML config template for the init command.
const sampleConfigTemplate = `# claude-statusline configuration
# Docs: https://github.com/felipeelias/claude-statusline

# Format string controls the layout. Modules are referenced with $name.
# Styled text groups use [text](style) syntax.
format = "$directory | $git_branch | $model | $cost | $context"

# Built-in palettes: "default", "tokyo-night", "gruvbox", "catppuccin"
# Run 'claude-statusline themes' to preview all palettes.
palette = "default"

# Custom palette: override or add your own palette colors.
# [palettes.my-theme]
# accent = "#ff5500"
# cost_ok = "green"
# cost_warn = "yellow"
# cost_high = "red"
# ctx_ok = "green"
# ctx_warn = "yellow"
# ctx_high = "red"

# Module configuration. Each module supports format, style, and disabled.
# Styles: "bold", "dim", "italic", "fg:#hex", "bg:#hex", "palette:name"

# [model]
# format = "{{.DisplayName}}"
# style = "bold"

# [directory]
# format = "{{.Dir}}"
# style = "palette:accent"
# truncation_length = 3

# [cost]
# format = '${{printf "%.2f" .TotalCostUSD}}'
# style = "palette:cost_ok"
# thresholds = [
#   { above = 1.0, style = "palette:cost_warn" },
#   { above = 5.0, style = "palette:cost_high" },
# ]

# [context]
# format = '{{.Bar}} {{printf "%.0f" .UsedPct}}%%'
# style = "palette:ctx_ok"
# bar_width = 5
# bar_fill = "█"
# bar_empty = "░"
# thresholds = [
#   { above = 50, style = "palette:ctx_warn" },
#   { above = 70, style = "208" },
#   { above = 90, style = "palette:ctx_high" },
# ]

# [git_branch]
# format = " {{.Branch}}{{if .InWorktree}} {{end}}"
# style = "palette:accent"

# Disabled by default. Set disabled = false and add to format string to enable.
# [session_timer]
# disabled = false
# format = "{{.Elapsed}}"
# style = "dim"

# [lines_changed]
# disabled = false
# format = "+{{.Added}} -{{.Removed}}"
# added_style = "green"
# removed_style = "red"
`

// SampleConfig returns a commented TOML config template for the init command.
func SampleConfig() string {
	return sampleConfigTemplate
}

// ResolveStyle resolves palette references in a style string.
// If styleStr starts with "palette:", the key after the prefix is looked up in
// the active palette. If found, the palette value is returned.
// Otherwise styleStr is returned unchanged.
func (c Config) ResolveStyle(styleStr string) string {
	key, found := strings.CutPrefix(styleStr, "palette:")
	if !found {
		return styleStr
	}

	palette, paletteExists := c.Palettes[c.Palette]
	if !paletteExists {
		return styleStr
	}

	value, valueExists := palette[key]
	if !valueExists {
		return styleStr
	}

	return value
}
