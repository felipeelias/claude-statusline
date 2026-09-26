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
var ansiEscapePattern = regexp.MustCompile(`\x1b\[[0-?]*[ -/]*[@-~]`)

// Render parses the format string from cfg, evaluates module references and
// styled text tokens, and returns the concatenated result.
func Render(cfg config.Config, data input.Data) (string, error) {
	format := cfg.Format
	if format == "" {
		return "", nil
	}

	separator := cfg.Separator
	if separator == "" {
		separator = config.DefaultSeparator
	}

	registry := buildRegistry(cfg)
	sections := splitSections(format, separator)
	if len(sections) == 1 {
		return renderSection(format, registry, cfg, data)
	}

	var renderedSections []string

	for _, section := range sections {
		rendered, err := renderSection(section, registry, cfg, data)
		if err != nil {
			return "", err
		}

		visibleText := ansiEscapePattern.ReplaceAllString(rendered, "")
		if strings.TrimSpace(visibleText) != "" {
			renderedSections = append(renderedSections, rendered)
		}
	}

	return strings.Join(renderedSections, separator), nil
}

func renderSection(
	format string, registry map[string]moduleEntry, cfg config.Config, data input.Data,
) (string, error) {
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
func splitSections(format, separator string) []string {
	tokenLocations := tokenPattern.FindAllStringIndex(format, -1)
	sections := make([]string, 0, strings.Count(format, separator)+1)
	sectionStart := 0
	searchStart := 0

	for searchStart < len(format) {
		separatorOffset := strings.Index(format[searchStart:], separator)
		if separatorOffset == -1 {
			break
		}

		separatorStart := searchStart + separatorOffset
		separatorEnd := separatorStart + len(separator)
		if withinToken(separatorStart, separatorEnd, tokenLocations) {
			searchStart = separatorEnd

			continue
		}

		sections = append(sections, format[sectionStart:max(sectionStart, separatorStart)])
		sectionStart = separatorEnd
		searchStart = separatorEnd
		if len(separator) > 1 {
			// Preserve overlapping separators without revisiting a single-byte separator.
			searchStart--
		}
	}

	return append(sections, format[sectionStart:])
}

func withinToken(start, end int, tokenLocations [][]int) bool {
	for _, location := range tokenLocations {
		if start >= location[0] && end <= location[1] {
			return true
		}
	}

	return false
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
		"effort":        {module: modules.EffortModule{}, disabled: cfg.Effort.Disabled},
		"output_style":  {module: modules.OutputStyleModule{}, disabled: cfg.OutputStyle.Disabled},
		"directory":     {module: modules.NewDirectoryModule(), disabled: cfg.Directory.Disabled},
		"cost":          {module: modules.CostModule{}, disabled: cfg.Cost.Disabled},
		"context":       {module: modules.ContextModule{}, disabled: cfg.Context.Disabled},
		"git_branch":    {module: modules.GitBranchModule{}, disabled: cfg.GitBranch.Disabled},
		"session_timer": {module: modules.SessionTimerModule{}, disabled: cfg.SessionTimer.Disabled},
		"lines_changed": {module: modules.LinesChangedModule{}, disabled: cfg.LinesChanged.Disabled},
		"usage":         {module: modules.UsageModule{}, disabled: cfg.Usage.Disabled},
		"version":       {module: modules.VersionModule{}, disabled: cfg.Version.Disabled},
		"vim_mode":      {module: modules.VimModeModule{}, disabled: cfg.VimMode.Disabled},
		"agent_name":    {module: modules.AgentNameModule{}, disabled: cfg.AgentName.Disabled},
	}
}
