package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds the full statusline configuration.
type Config struct {
	Preset       string             `toml:"preset"`
	Format       string             `toml:"format"`
	Model        ModelConfig        `toml:"model"`
	Directory    DirectoryConfig    `toml:"directory"`
	Cost         CostConfig         `toml:"cost"`
	Context      ContextConfig      `toml:"context"`
	GitBranch    GitBranchConfig    `toml:"git_branch"`
	SessionTimer SessionTimerConfig `toml:"session_timer"`
	LinesChanged LinesChangedConfig `toml:"lines_changed"`
	Usage        UsageConfig        `toml:"usage"`
	Version      VersionConfig      `toml:"version"`
	VimMode      VimModeConfig      `toml:"vim_mode"`
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
	BarStyle   string      `toml:"bar_style"`
	BarFill    string      `toml:"bar_fill"`
	BarEmpty   string      `toml:"bar_empty"`
	Thresholds []Threshold `toml:"thresholds"`
}

// GitBranchConfig holds git branch module settings.
type GitBranchConfig struct {
	Format   string `toml:"format"`
	Style    string `toml:"style"`
	Disabled bool   `toml:"disabled"`
	Mode     string `toml:"mode"` // "detailed" (default) or "simple"
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

// VersionConfig holds version module settings.
type VersionConfig struct {
	Format   string `toml:"format"`
	Style    string `toml:"style"`
	Disabled bool   `toml:"disabled"`
}

// UsageConfig holds usage module settings.
type UsageConfig struct {
	Format     string      `toml:"format"`
	Style      string      `toml:"style"`
	Disabled   bool        `toml:"disabled"`
	BarWidth   int         `toml:"bar_width"`
	BarStyle   string      `toml:"bar_style"`
	BarFill    string      `toml:"bar_fill"`
	BarEmpty   string      `toml:"bar_empty"`
	Thresholds []Threshold `toml:"thresholds"`
}

// VimModeConfig holds vim mode module settings.
type VimModeConfig struct {
	Format   string `toml:"format"`
	Style    string `toml:"style"`
	Disabled bool   `toml:"disabled"`
}

const (
	defaultTruncationLength = 3
	defaultBarWidth = 5
	costWarnThreshold       = 5.0
	ctxWarnThreshold        = 50
	ctxHighThreshold        = 90
	usageWarnThreshold      = 75
	usageHighThreshold      = 90
)

// Default returns a Config with hardcoded default values.
//
//nolint:funlen // single-struct initializer reads best as one block
func Default() Config {
	return Config{
		Preset: "default",
		Format: "$directory | $git_branch | $model | $cost | $context",
		Model: ModelConfig{
			Format: "{{.DisplayName}}",
			Style:  "bold",
		},
		Directory: DirectoryConfig{
			Format:           "{{.Dir}}",
			Style:            "cyan",
			TruncationLength: defaultTruncationLength,
		},
		Cost: CostConfig{
			Format: `${{printf "%.2f" .TotalCostUSD}}`,
			Style:  "green",
			Thresholds: []Threshold{
				{Above: 1.0, Style: "yellow"},
				{Above: costWarnThreshold, Style: "red"},
			},
		},
		Context: ContextConfig{
			Format:   `{{.Bar}} {{printf "%.0f" .UsedPct}}%`,
			Style:    "green",
			BarWidth: defaultBarWidth,
			Thresholds: []Threshold{
				{Above: ctxWarnThreshold, Style: "yellow"},
				{Above: ctxHighThreshold, Style: "red"},
			},
		},
		GitBranch: GitBranchConfig{
			Format: iconBranch + " {{.Branch}}" +
				"{{if .InWorktree}} " + iconWorktree + "{{end}}" +
				"{{if .IsDirty}} *{{end}}" +
				"{{if .Ahead}} \u2191{{.Ahead}}{{end}}" +
				"{{if .Behind}} \u2193{{.Behind}}{{end}}",
			Style: "cyan",
			Mode:  "detailed",
		},
		SessionTimer: SessionTimerConfig{
			Format:   "{{if .Hours}}{{.Hours}}h{{end}}{{printf \"%02d\" .Minutes}}m{{printf \"%02d\" .Seconds}}s",
			Style:    "dim",
			Disabled: true,
		},
		LinesChanged: LinesChangedConfig{
			Format:       "+{{.Added}} -{{.Removed}}",
			AddedStyle:   "green",
			RemovedStyle: "red",
			Disabled:     true,
		},
		Version: VersionConfig{
			Format:   `v{{.Version}}`,
			Style:    "dim",
			Disabled: true,
		},
		Usage: UsageConfig{
			Format:   `{{.BlockBar}} {{printf "%.0f" .BlockPct}}% W:{{printf "%.0f" .WeeklyPct}}%`,
			Style:    "green",
			Disabled: true,
			BarWidth: defaultBarWidth,
			Thresholds: []Threshold{
				{Above: usageWarnThreshold, Style: "yellow"},
				{Above: usageHighThreshold, Style: "red"},
			},
		},
		VimMode: VimModeConfig{
			Format:   "{{.Mode}}",
			Style:    "bold yellow",
			Disabled: true,
		},
	}
}

// presetHeader is used to extract the preset field from a TOML file
// before applying the full config on top.
type presetHeader struct {
	Preset string `toml:"preset"`
}

// Load reads a TOML config file and merges it with defaults.
// If the file does not exist, Default() is returned with no error.
// If the file exists but has parse errors, an error is returned.
//
// Loading is two-pass: first the preset field is read to select the base
// config, then the full file is decoded on top so user overrides layer cleanly.
func Load(path string) (Config, error) {
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Default(), nil
	}
	if err != nil {
		return Config{}, err
	}

	raw := string(content)

	// Pass 1: read preset field.
	var header presetHeader

	_, err = toml.Decode(raw, &header)
	if err != nil {
		return Config{}, err
	}

	// Pass 2: start from preset base, decode user overrides on top.
	cfg, _ := ApplyPreset(header.Preset)

	_, err = toml.Decode(raw, &cfg)
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

# Preset defines the complete visual style (layout, colors, separators).
# Built-in presets:
#   "default"          - flat with | pipes (no Nerd Font needed)
#   "minimal"          - clean spacing, no separators (no Nerd Font needed)
#   "pastel-powerline" - pastel powerline arrows (Nerd Font)
#   "tokyo-night"      - dark blues rounded powerline (Nerd Font)
#   "gruvbox-rainbow"  - earthy rainbow powerline (Nerd Font)
#   "catppuccin"       - Catppuccin Mocha powerline (Nerd Font)
# Run 'claude-statusline themes' to preview all presets.
preset = "default"

# Format string controls the layout. Modules are referenced with $name.
# Styled text groups use [text](style) syntax.
# When using a preset, you typically don't need to change the format.
format = "$directory | $git_branch | $model | $cost | $context"

# Module configuration. Each module supports format, style, and disabled.
# Styles: "bold", "dim", "italic", "fg:#hex", "bg:#hex", "208"

# [model]
# format = "{{.DisplayName}}"
# style = "bold"
# Template fields: DisplayName, ID, Short (e.g. "Sonnet 4.6")

# [directory]
# format = "{{.Dir}}"
# style = "cyan"
# truncation_length = 3

# [cost]
# format = '${{printf "%.2f" .TotalCostUSD}}'
# style = "green"
# thresholds = [
#   { above = 1.0, style = "yellow" },
#   { above = 5.0, style = "red" },
# ]

# [context]
# format = '{{.Bar}} {{printf "%.0f" .UsedPct}}%'
# style = "green"
# bar_width = 5
# bar_style = "classic"  # "classic", "blocks", "dots", "line", "squares"
# bar_fill = "█"         # overrides bar_style fill character
# bar_empty = "░"        # overrides bar_style empty character
# thresholds = [
#   { above = 50, style = "yellow" },
#   { above = 90, style = "red" },
# ]

# [git_branch]
# mode = "detailed"  # "detailed" (default) or "simple" (fast, branch only)
# style = "cyan"
# Template fields: Branch, InWorktree, IsDirty, IsClean,
#   Staged, Modified, Untracked, Ahead, Behind, Conflicts

# Disabled by default. Set disabled = false and add the module to format string to enable.

# [version]
# disabled = false
# format = "v{{.Version}}"
# style = "dim"

# [session_timer]
# disabled = false
# format = "{{if .Hours}}{{.Hours}}h{{end}}{{printf \"%02d\" .Minutes}}m{{printf \"%02d\" .Seconds}}s"
# style = "dim"

# [lines_changed]
# disabled = false
# format = "+{{.Added}} -{{.Removed}}"
# added_style = "green"
# removed_style = "red"

# Requires Claude Code 2.1.80+ which provides rate_limits in the status line payload.
# Add $usage to your format string to display it.
# [usage]
# disabled = false
# format = '{{.BlockBar}} {{printf "%.0f" .BlockPct}}% W:{{printf "%.0f" .WeeklyPct}}%'
# style = "green"
# bar_width = 5
# bar_style = "classic"  # "classic", "blocks", "dots", "line", "squares"
# bar_fill = "█"         # overrides bar_style fill character
# bar_empty = "░"        # overrides bar_style empty character
# thresholds = [
#   { above = 75, style = "yellow" },
#   { above = 90, style = "red" },
# ]
# Template fields: BlockPct, WeeklyPct, BlockBar, WeeklyBar, BlockResets, WeeklyResets

# [vim_mode]
# disabled = false
# format = "{{.Mode}}"
# style = "bold yellow"
# Template fields: Mode (e.g. "NORMAL", "INSERT")
`

// SampleConfig returns a commented TOML config template for the init command.
func SampleConfig() string {
	return sampleConfigTemplate
}
