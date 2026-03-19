package render

import (
	"regexp"
	"strings"

	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
	"github.com/felipeelias/claude-statusline/internal/modules"
	"github.com/felipeelias/claude-statusline/internal/style"
)

// moduleEntry pairs a module with its disabled flag from config.
type moduleEntry struct {
	module   modules.Module
	disabled bool
}

// tokenPattern matches module references ($word) and styled text ([text](style)).
// The order matters: styled text is matched first to avoid $-matching inside it.
var tokenPattern = regexp.MustCompile(`\[([^\]]*)\]\(([^)]*)\)|\$([a-z_]+)`)

// Render parses the format string from cfg, evaluates module references and
// styled text tokens, and returns the concatenated result.
func Render(cfg config.Config, data input.Data) (string, error) {
	format := cfg.Format
	if format == "" {
		return "", nil
	}

	registry := buildRegistry(cfg)

	var result strings.Builder

	lastIndex := 0
	matches := tokenPattern.FindAllStringSubmatchIndex(format, -1)

	for _, loc := range matches {
		if loc[0] > lastIndex {
			result.WriteString(format[lastIndex:loc[0]])
		}

		rendered, err := renderMatch(format, loc, registry, cfg, data)
		if err != nil {
			return "", err
		}

		result.WriteString(rendered)
		lastIndex = loc[1]
	}

	if lastIndex < len(format) {
		result.WriteString(format[lastIndex:])
	}

	return result.String(), nil
}

func renderMatch(
	format string, loc []int, registry map[string]moduleEntry, cfg config.Config, data input.Data,
) (string, error) {
	if loc[2] != -1 && loc[4] != -1 {
		text := format[loc[2]:loc[3]]
		styleStr := format[loc[4]:loc[5]]

		return style.Parse(styleStr).Wrap(text), nil
	}

	if loc[6] != -1 {
		name := format[loc[6]:loc[7]]
		entry, ok := registry[name]
		if ok && !entry.disabled {
			return entry.module.Render(data, cfg)
		}
	}

	return "", nil
}

// buildRegistry creates a map from module name to moduleEntry, pairing each
// module with its disabled flag from config.
func buildRegistry(cfg config.Config) map[string]moduleEntry {
	return map[string]moduleEntry{
		"model":         {module: modules.ModelModule{}, disabled: cfg.Model.Disabled},
		"directory":     {module: modules.NewDirectoryModule(), disabled: cfg.Directory.Disabled},
		"cost":          {module: modules.CostModule{}, disabled: cfg.Cost.Disabled},
		"context":       {module: modules.ContextModule{}, disabled: cfg.Context.Disabled},
		"git_branch":    {module: modules.GitBranchModule{}, disabled: cfg.GitBranch.Disabled},
		"session_timer": {module: modules.SessionTimerModule{}, disabled: cfg.SessionTimer.Disabled},
		"lines_changed": {module: modules.LinesChangedModule{}, disabled: cfg.LinesChanged.Disabled},
		"usage":         {module: modules.NewUsageModule(), disabled: cfg.Usage.Disabled},
	}
}
