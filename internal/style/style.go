package style

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	ansiBlack   = 30
	ansiRed     = 31
	ansiGreen   = 32
	ansiYellow  = 33
	ansiBlue    = 34
	ansiMagenta = 35
	ansiCyan    = 36
	ansiWhite   = 37

	ansiBold      = 1
	ansiDim       = 2
	ansiItalic    = 3
	ansiUnderline = 4

	ansiFgExtended   = 38
	ansiBgExtended   = 48
	ansiFgToBgOffset = 10

	hexColorLen = 7
	hexBase     = 16
	hexBitSize  = 8
)

var namedColors = map[string]int{
	"black":   ansiBlack,
	"red":     ansiRed,
	"green":   ansiGreen,
	"yellow":  ansiYellow,
	"blue":    ansiBlue,
	"magenta": ansiMagenta,
	"cyan":    ansiCyan,
	"white":   ansiWhite,
}

var attributes = map[string]int{
	"bold":      ansiBold,
	"dim":       ansiDim,
	"italic":    ansiItalic,
	"underline": ansiUnderline,
}

// Style holds parsed ANSI codes and can wrap text with them.
type Style struct {
	codes []string
}

// Parse parses a starship-like style string into a Style.
func Parse(spec string) Style {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return Style{}
	}

	var codes []string

	for token := range strings.FieldsSeq(spec) {
		if parsed, ok := parseToken(token); ok {
			codes = append(codes, parsed...)
		}
	}

	return Style{codes: codes}
}

func parseToken(token string) ([]string, bool) {
	if code, ok := attributes[token]; ok {
		return []string{strconv.Itoa(code)}, true
	}

	if code, ok := namedColors[token]; ok {
		return []string{strconv.Itoa(code)}, true
	}

	if value, found := strings.CutPrefix(token, "fg:"); found {
		return parseFg(value)
	}

	if value, found := strings.CutPrefix(token, "bg:"); found {
		return parseBg(value)
	}

	n, err := strconv.Atoi(token)
	if err == nil && n >= 0 && n <= 255 {
		return []string{fmt.Sprintf("38;5;%d", n)}, true
	}

	return nil, false
}

func parseFg(value string) ([]string, bool) {
	if strings.HasPrefix(value, "#") {
		return parseHexColor(value, ansiFgExtended)
	}

	if code, ok := namedColors[value]; ok {
		return []string{strconv.Itoa(code)}, true
	}

	return nil, false
}

func parseBg(value string) ([]string, bool) {
	if strings.HasPrefix(value, "#") {
		return parseHexColor(value, ansiBgExtended)
	}

	if code, ok := namedColors[value]; ok {
		return []string{strconv.Itoa(code + ansiFgToBgOffset)}, true
	}

	return nil, false
}

func parseHexColor(hex string, base int) ([]string, bool) {
	if len(hex) != hexColorLen || hex[0] != '#' {
		return nil, false
	}

	red, err := strconv.ParseUint(hex[1:3], hexBase, hexBitSize)
	if err != nil {
		return nil, false
	}

	green, err := strconv.ParseUint(hex[3:5], hexBase, hexBitSize)
	if err != nil {
		return nil, false
	}

	blue, err := strconv.ParseUint(hex[5:7], hexBase, hexBitSize)
	if err != nil {
		return nil, false
	}

	return []string{fmt.Sprintf("%d;2;%d;%d;%d", base, red, green, blue)}, true
}

// Wrap wraps text with ANSI escape codes. If no codes are set, returns text unchanged.
func (s Style) Wrap(text string) string {
	if len(s.codes) == 0 {
		return text
	}

	return fmt.Sprintf("\033[%sm%s\033[0m", strings.Join(s.codes, ";"), text)
}
