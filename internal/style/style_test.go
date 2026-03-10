package style_test

import (
	"strings"
	"testing"

	"github.com/felipeelias/claude-statusline/internal/style"
	"github.com/stretchr/testify/assert"
)

func TestWrapNamed(t *testing.T) {
	tests := []struct{ name, style, text, expected string }{
		{"bold", "bold", "hi", "\033[1mhi\033[0m"},
		{"green fg", "green", "hi", "\033[32mhi\033[0m"},
		{"bold green", "bold green", "hi", "\033[1;32mhi\033[0m"},
		{"dim", "dim", "hi", "\033[2mhi\033[0m"},
		{"italic", "italic", "hi", "\033[3mhi\033[0m"},
		{"underline", "underline", "hi", "\033[4mhi\033[0m"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := style.Parse(tt.style)
			assert.Equal(t, tt.expected, s.Wrap(tt.text))
		})
	}
}

func TestWrapHex(t *testing.T) {
	s := style.Parse("fg:#ff5500")
	assert.Equal(t, "\033[38;2;255;85;0mhi\033[0m", s.Wrap("hi"))
}

func TestWrapBareHex(t *testing.T) {
	s := style.Parse("#769ff0")
	assert.Equal(t, "\033[38;2;118;159;240mhi\033[0m", s.Wrap("hi"))
}

func TestWrapBg(t *testing.T) {
	s := style.Parse("bg:blue")
	assert.Equal(t, "\033[44mhi\033[0m", s.Wrap("hi"))
}

func TestWrapBgHex(t *testing.T) {
	s := style.Parse("bg:#333333")
	assert.Equal(t, "\033[48;2;51;51;51mhi\033[0m", s.Wrap("hi"))
}

func TestWrap256Color(t *testing.T) {
	s := style.Parse("208")
	assert.Equal(t, "\033[38;5;208mhi\033[0m", s.Wrap("hi"))
}

func TestWrapCombined(t *testing.T) {
	s := style.Parse("fg:#aaaaaa bg:#333333 bold")
	result := s.Wrap("hi")
	assert.Contains(t, result, "1")
	assert.Contains(t, result, "38;2;170;170;170")
	assert.Contains(t, result, "48;2;51;51;51")
	assert.True(t, strings.HasSuffix(result, "\033[0m"))
}

func TestWrapEmpty(t *testing.T) {
	s := style.Parse("")
	assert.Equal(t, "hi", s.Wrap("hi"))
}

func TestWrapFgNamedPrefix(t *testing.T) {
	s := style.Parse("fg:red")
	assert.Equal(t, "\033[31mhi\033[0m", s.Wrap("hi"))
}

func TestWrapFg256Color(t *testing.T) {
	s := style.Parse("fg:208")
	assert.Equal(t, "\033[38;5;208mhi\033[0m", s.Wrap("hi"))
}

func TestWrapBg256Color(t *testing.T) {
	s := style.Parse("bg:238")
	assert.Equal(t, "\033[48;5;238mhi\033[0m", s.Wrap("hi"))
}

func TestWrapFgBg256ColorCombined(t *testing.T) {
	s := style.Parse("fg:208 bg:238 bold")
	result := s.Wrap("hi")
	assert.Contains(t, result, "38;5;208")
	assert.Contains(t, result, "48;5;238")
	assert.Contains(t, result, "1")
}

func TestWrapBgNamed(t *testing.T) {
	tests := []struct{ name, style, expected string }{
		{"bg:red", "bg:red", "\033[41mhi\033[0m"},
		{"bg:green", "bg:green", "\033[42mhi\033[0m"},
		{"bg:yellow", "bg:yellow", "\033[43mhi\033[0m"},
		{"bg:blue", "bg:blue", "\033[44mhi\033[0m"},
		{"bg:magenta", "bg:magenta", "\033[45mhi\033[0m"},
		{"bg:cyan", "bg:cyan", "\033[46mhi\033[0m"},
		{"bg:white", "bg:white", "\033[47mhi\033[0m"},
		{"bg:black", "bg:black", "\033[40mhi\033[0m"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := style.Parse(tt.style)
			assert.Equal(t, tt.expected, s.Wrap("hi"))
		})
	}
}
