package config

import "sort"

const (
	// Powerline / Nerd Font separator glyphs.
	plRight    = "\ue0b0" // Powerline right arrow:
	plLeftCap  = "\ue0b6" // Powerline left half-circle:
	plRightCap = "\ue0b4" // Powerline right half-circle:
	// Git branch icon.
	iconBranch = "\ue0a0"
	// Worktree icon.
	iconWorktree = "\uf0e8"
)

// PresetNames returns a sorted list of built-in preset names.
func PresetNames() []string {
	names := make([]string, 0, len(builtinPresets))
	for name := range builtinPresets {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
}

// ApplyPreset returns a Config for the named preset.
// If the preset is not found, it falls back to Default().
// The second return value indicates whether the preset was found.
func ApplyPreset(name string) (Config, bool) {
	fn, ok := builtinPresets[name]
	if !ok {
		return Default(), false
	}

	return fn(), true
}

var builtinPresets = map[string]func() Config{
	"default":          Default,
	"minimal":          presetMinimal,
	"pastel-powerline": presetPastelPowerline,
	"tokyo-night":      presetTokyoNight,
	"gruvbox-rainbow":  presetGruvboxRainbow,
	"catppuccin":       presetCatppuccin,
}

// Minimal — clean spacing, no separators, no icons, no background colors.
func presetMinimal() Config {
	cfg := Default()
	cfg.Preset = "minimal"
	cfg.Format = "$directory  $git_branch  $model  $cost  $context"
	cfg.Directory.Style = "blue"
	cfg.GitBranch.Format = "{{.Branch}}"
	cfg.GitBranch.Style = "cyan"
	cfg.Model.Style = "bold"
	cfg.Cost.Style = "green"
	cfg.Context.Format = `{{printf "%.0f" .UsedPct}}%`
	cfg.Context.Style = "green"

	return cfg
}

type thresholdColors struct {
	warn string
	high string
}

// capsuleFormat builds a powerline format with a left half-circle cap,
// right-arrow transitions, and the given trailing glyph.
func capsuleFormat(colors [5]string, trailing string) string {
	return "[" + plLeftCap + "](fg:" + colors[0] + ")" +
		"$directory" +
		"[" + plRight + "](fg:" + colors[0] + " bg:" + colors[1] + ")" +
		"$git_branch" +
		"[" + plRight + "](fg:" + colors[1] + " bg:" + colors[2] + ")" +
		"$model" +
		"[" + plRight + "](fg:" + colors[2] + " bg:" + colors[3] + ")" +
		"$cost" +
		"[" + plRight + "](fg:" + colors[3] + " bg:" + colors[4] + ")" +
		"$context" +
		"[" + trailing + " ](fg:" + colors[4] + ")"
}

// segStyle builds a style string with optional foreground and required background.
// When segFg is empty, no fg is set (terminal default foreground).
func segStyle(segFg string, bgColor string) string {
	if segFg == "" {
		return "bg:" + bgColor
	}

	return "fg:" + segFg + " bg:" + bgColor
}

// powerlineConfig builds a powerline-style Config with the given format and colors.
// Pass segFg="" to use terminal default foreground (like Starship's Pastel Powerline).
func powerlineConfig(preset string, format string, segFg string, colors [5]string, thresholds thresholdColors) Config {
	return Config{
		Preset: preset,
		Format: format,
		Directory: DirectoryConfig{
			Format: " {{.Dir}} ", Style: segStyle(segFg, colors[0]),
			TruncationLength: defaultTruncationLength,
		},
		GitBranch: GitBranchConfig{
			Format: " " + iconBranch + " {{.Branch}}{{if .InWorktree}} " + iconWorktree + "{{end}} ",
			Style:  segStyle(segFg, colors[1]),
		},
		Model: ModelConfig{
			Format: " {{.DisplayName}} ", Style: segStyle(segFg, colors[2]) + " bold",
		},
		Cost: CostConfig{
			Format: ` ${{printf "%.2f" .TotalCostUSD}} `,
			Style:  segStyle(segFg, colors[3]),
			Thresholds: []Threshold{
				{Above: 1.0, Style: segStyle(thresholds.warn, colors[3])},
				{Above: costWarnThreshold, Style: segStyle(thresholds.high, colors[3])},
			},
		},
		Context: ContextConfig{
			Format: ` {{.Bar}} {{printf "%.0f" .UsedPct}}% `, Style: segStyle(segFg, colors[4]),
			BarWidth: defaultBarWidth, BarFill: defaultBarFill, BarEmpty: defaultBarEmpty,
			Thresholds: []Threshold{
				{Above: ctxWarnThreshold, Style: segStyle(thresholds.warn, colors[4])},
				{Above: ctxHighThreshold, Style: segStyle(thresholds.high, colors[4])},
			},
		},
		SessionTimer: SessionTimerConfig{Format: " {{.Elapsed}} ", Style: "dim", Disabled: true},
		LinesChanged: LinesChangedConfig{
			Format: " +{{.Added}} -{{.Removed}} ", AddedStyle: "green", RemovedStyle: "red", Disabled: true,
		},
	}
}

// Pastel Powerline — based on Starship's Pastel Powerline preset.
// Left half-circle cap, arrow transitions, arrow trailing.
// Colors: purple → pink → peach → blue → dark blue.
func presetPastelPowerline() Config {
	colors := [5]string{"#DA627D", "#FCA17D", "#86BBD8", "#06969A", "#33658A"}

	return powerlineConfig("pastel-powerline",
		capsuleFormat(colors, plRight),
		"", colors,
		thresholdColors{warn: "#f9e2af", high: "#f38ba8"},
	)
}

// Tokyo Night — based on Starship's Tokyo Night preset.
// Gradient ░▒▓ leading, all rounded half-circle transitions.
// Colors: bright blue → dark blue-gray → darker → darkest → near-black.
func presetTokyoNight() Config {
	colors := [5]string{"#769ff0", "#394260", "#212736", "#1d2230", "#1a1b26"}
	format := "[\u2591\u2592\u2593](fg:#a3aed2)" +
		"[" + plRightCap + "](fg:#a3aed2 bg:" + colors[0] + ")" +
		"$directory" +
		"[" + plRightCap + "](fg:" + colors[0] + " bg:" + colors[1] + ")" +
		"$git_branch" +
		"[" + plRightCap + "](fg:" + colors[1] + " bg:" + colors[2] + ")" +
		"$model" +
		"[" + plRightCap + "](fg:" + colors[2] + " bg:" + colors[3] + ")" +
		"$cost" +
		"[" + plRightCap + "](fg:" + colors[3] + " bg:" + colors[4] + ")" +
		"$context" +
		"[" + plRightCap + " ](fg:" + colors[4] + ")"

	return powerlineConfig("tokyo-night",
		format,
		"#e3e5e5", colors,
		thresholdColors{warn: "#e0af68", high: "#f7768e"},
	)
}

// Gruvbox Rainbow — based on Starship's Gruvbox Rainbow preset.
// Left half-circle cap, arrow transitions, rounded half-circle trailing.
// Colors: yellow → aqua → blue → gray → dark.
func presetGruvboxRainbow() Config {
	colors := [5]string{"#d79921", "#689d6a", "#458588", "#665c54", "#3c3836"}

	return powerlineConfig("gruvbox-rainbow",
		capsuleFormat(colors, plRightCap),
		"#fbf1c7", colors,
		thresholdColors{warn: "#fabd2f", high: "#fb4934"},
	)
}

// Catppuccin — based on Starship's Catppuccin Powerline preset (Mocha).
// Left half-circle cap, arrow transitions, rounded half-circle trailing.
// Colors: peach → yellow → green → sapphire → lavender.
func presetCatppuccin() Config {
	colors := [5]string{"#fab387", "#f9e2af", "#a6e3a1", "#74c7ec", "#b4befe"}

	return powerlineConfig("catppuccin",
		capsuleFormat(colors, plRightCap),
		"#11111b", colors,
		thresholdColors{warn: "#f9e2af", high: "#f38ba8"},
	)
}
