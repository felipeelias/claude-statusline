package config

import "sort"

const (
	// Powerline / Nerd Font separator glyphs.
	plRight    = "\ue0b0" // Powerline right arrow
	plRoundedL = "\ue0b6" // Rounded left (open)
	plRoundedR = "\ue0b4" // Rounded right (close)
)

// ThemeNames returns a sorted list of built-in theme names.
func ThemeNames() []string {
	names := make([]string, 0, len(builtinThemes))
	for name := range builtinThemes {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
}

// ApplyTheme returns a Config with the named theme applied.
// If the theme is not found, it falls back to Default().
// The second return value indicates whether the theme was found.
func ApplyTheme(name string) (Config, bool) {
	fn, ok := builtinThemes[name]
	if !ok {
		return Default(), false
	}

	return fn(), true
}

// builtinThemes maps theme names to constructor functions.
var builtinThemes = map[string]func() Config{
	"default":   themeDefault,
	"powerline": themePowerline,
	"rounded":   themeRounded,
	"minimal":   themeMinimal,
}

func themeDefault() Config {
	return Default()
}

// segmentModules returns module configs shared by powerline and rounded themes.
func segmentModules() (
	ModelConfig, DirectoryConfig, CostConfig, ContextConfig, GitBranchConfig,
	SessionTimerConfig, LinesChangedConfig,
) {
	model := ModelConfig{
		Format: " {{.DisplayName}} ",
		Style:  "fg:palette:seg_fg bg:palette:model_bg bold",
	}

	directory := DirectoryConfig{
		Format:           " {{.Dir}} ",
		Style:            "fg:palette:seg_fg bg:palette:dir_bg",
		TruncationLength: defaultTruncationLength,
	}

	cost := CostConfig{
		Format: ` ${{printf "%.2f" .TotalCostUSD}} `,
		Style:  "fg:palette:cost_ok bg:palette:cost_bg",
		Thresholds: []Threshold{
			{Above: 1.0, Style: "fg:palette:cost_warn bg:palette:cost_bg"},
			{Above: costWarnThreshold, Style: "fg:palette:cost_high bg:palette:cost_bg"},
		},
	}

	context := ContextConfig{
		Format:   ` {{.Bar}} {{printf "%.0f" .UsedPct}}% `,
		Style:    "fg:palette:ctx_ok bg:palette:ctx_bg",
		BarWidth: defaultBarWidth,
		BarFill:  "\u2588",
		BarEmpty: "\u2591",
		Thresholds: []Threshold{
			{Above: ctxWarnThreshold, Style: "fg:palette:ctx_warn bg:palette:ctx_bg"},
			{Above: ctxMedThreshold, Style: "fg:208 bg:palette:ctx_bg"},
			{Above: ctxHighThreshold, Style: "fg:palette:ctx_high bg:palette:ctx_bg"},
		},
	}

	gitBranch := GitBranchConfig{
		Format: " \ue0a0 {{.Branch}}{{if .InWorktree}} \uf0e8{{end}} ",
		Style:  "fg:palette:seg_fg bg:palette:git_bg",
	}

	sessionTimer := SessionTimerConfig{
		Format:   " {{.Elapsed}} ",
		Style:    "dim",
		Disabled: true,
	}

	linesChanged := LinesChangedConfig{
		Format:       " +{{.Added}} -{{.Removed}} ",
		AddedStyle:   "green",
		RemovedStyle: "red",
		Disabled:     true,
	}

	return model, directory, cost, context, gitBranch, sessionTimer, linesChanged
}

func themePowerline() Config {
	model, directory, cost, context, gitBranch, sessionTimer, linesChanged := segmentModules()

	return Config{
		Theme:   "powerline",
		Palette: "default",
		Format: "[" + plRight + "](fg:palette:dir_bg)" +
			"$directory" +
			"[" + plRight + "](fg:palette:dir_bg bg:palette:git_bg)" +
			"$git_branch" +
			"[" + plRight + "](fg:palette:git_bg bg:palette:model_bg)" +
			"$model" +
			"[" + plRight + "](fg:palette:model_bg bg:palette:cost_bg)" +
			"$cost" +
			"[" + plRight + "](fg:palette:cost_bg bg:palette:ctx_bg)" +
			"$context" +
			"[" + plRight + "](fg:palette:ctx_bg)",
		Palettes:     defaultPalettes(),
		Model:        model,
		Directory:    directory,
		Cost:         cost,
		Context:      context,
		GitBranch:    gitBranch,
		SessionTimer: sessionTimer,
		LinesChanged: linesChanged,
	}
}

func themeRounded() Config {
	model, directory, cost, context, gitBranch, sessionTimer, linesChanged := segmentModules()

	return Config{
		Theme:   "rounded",
		Palette: "default",
		Format: "[" + plRoundedL + "](fg:palette:dir_bg)" +
			"$directory" +
			"[" + plRoundedR + "](fg:palette:dir_bg) " +
			"[" + plRoundedL + "](fg:palette:git_bg)" +
			"$git_branch" +
			"[" + plRoundedR + "](fg:palette:git_bg) " +
			"[" + plRoundedL + "](fg:palette:model_bg)" +
			"$model" +
			"[" + plRoundedR + "](fg:palette:model_bg) " +
			"[" + plRoundedL + "](fg:palette:cost_bg)" +
			"$cost" +
			"[" + plRoundedR + "](fg:palette:cost_bg) " +
			"[" + plRoundedL + "](fg:palette:ctx_bg)" +
			"$context" +
			"[" + plRoundedR + "](fg:palette:ctx_bg)",
		Palettes:     defaultPalettes(),
		Model:        model,
		Directory:    directory,
		Cost:         cost,
		Context:      context,
		GitBranch:    gitBranch,
		SessionTimer: sessionTimer,
		LinesChanged: linesChanged,
	}
}

func themeMinimal() Config {
	cfg := Default()
	cfg.Theme = "minimal"
	cfg.Format = "$directory  $git_branch  $model  $cost  $context"
	cfg.Directory.Style = "dim"
	cfg.GitBranch.Format = "{{.Branch}}"
	cfg.GitBranch.Style = "dim"
	cfg.Context.Format = `{{printf "%.0f" .UsedPct}}%`

	return cfg
}
