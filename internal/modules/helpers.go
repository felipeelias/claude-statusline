package modules

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/style"
)

// renderTemplate executes a Go text/template with the given data and returns the result.
func renderTemplate(name, format string, data any) (string, error) {
	tmpl, err := template.New(name).Parse(format)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer

	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// wrapStyle parses a style string and wraps text with ANSI codes.
func wrapStyle(text, styleStr string) string {
	return style.Parse(styleStr).Wrap(text)
}

const pctMax = 100

// barPreset defines a named pair of fill/empty characters for progress bars.
type barPreset struct {
	Fill  string
	Empty string
}

// barStyles maps bar_style names to their fill/empty character presets.
var barStyles = map[string]barPreset{
	"classic": {Fill: "\u2588", Empty: "\u2591"}, // █ ░
	"blocks":  {Fill: "\u2588", Empty: "\u2592"}, // █ ▒
	"dots":    {Fill: "\u28ff", Empty: "\u28c0"}, // ⣿ ⣀
	"line":    {Fill: "\u2501", Empty: "\u2500"}, // ━ ─
	"squares": {Fill: "\u25fc", Empty: "\u25fb"}, // ◼ ◻
}

// resolveBarChars determines the fill and empty characters for a progress bar.
// Priority: explicit bar_fill/bar_empty > bar_style preset > classic defaults.
func resolveBarChars(barStyle, barFill, barEmpty string) (string, string) {
	classic := barStyles["classic"]
	fill := classic.Fill
	empty := classic.Empty

	if preset, ok := barStyles[barStyle]; ok {
		fill = preset.Fill
		empty = preset.Empty
	}

	if barFill != "" {
		fill = barFill
	}

	if barEmpty != "" {
		empty = barEmpty
	}

	return fill, empty
}

// buildBar creates a progress bar string from a percentage value.
func buildBar(pct float64, width int, fill, empty string) string {
	filled := min(max(int(pct/pctMax*float64(width)), 0), width)
	emptyCount := width - filled

	return strings.Repeat(fill, filled) + strings.Repeat(empty, emptyCount)
}

// resolveThresholdStyle evaluates thresholds in order. The last threshold whose
// Above value is less than the given value wins. If none match, the base style is used.
func resolveThresholdStyle(value float64, thresholds []config.Threshold, baseStyle string) string {
	winner := baseStyle
	for _, threshold := range thresholds {
		if value > threshold.Above {
			winner = threshold.Style
		}
	}

	return winner
}
