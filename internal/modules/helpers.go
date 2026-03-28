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
