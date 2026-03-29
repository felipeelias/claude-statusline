package modules

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveBarChars(t *testing.T) {
	t.Run("no style or explicit chars uses classic defaults", func(t *testing.T) {
		fill, empty := resolveBarChars("", "", "")
		assert.Equal(t, "\u2588", fill)
		assert.Equal(t, "\u2591", empty)
	})

	t.Run("bar_style selects preset", func(t *testing.T) {
		fill, empty := resolveBarChars("dots", "", "")
		assert.Equal(t, "\u28ff", fill)
		assert.Equal(t, "\u28c0", empty)
	})

	t.Run("bar_style blocks", func(t *testing.T) {
		fill, empty := resolveBarChars("blocks", "", "")
		assert.Equal(t, "\u2588", fill)
		assert.Equal(t, "\u2592", empty)
	})

	t.Run("bar_style line", func(t *testing.T) {
		fill, empty := resolveBarChars("line", "", "")
		assert.Equal(t, "\u2501", fill)
		assert.Equal(t, "\u2500", empty)
	})

	t.Run("bar_style squares", func(t *testing.T) {
		fill, empty := resolveBarChars("squares", "", "")
		assert.Equal(t, "\u25fc", fill)
		assert.Equal(t, "\u25fb", empty)
	})

	t.Run("explicit bar_fill overrides bar_style", func(t *testing.T) {
		fill, empty := resolveBarChars("dots", "X", "")
		assert.Equal(t, "X", fill)
		assert.Equal(t, "\u28c0", empty)
	})

	t.Run("explicit bar_empty overrides bar_style", func(t *testing.T) {
		fill, empty := resolveBarChars("dots", "", "O")
		assert.Equal(t, "\u28ff", fill)
		assert.Equal(t, "O", empty)
	})

	t.Run("explicit both override bar_style completely", func(t *testing.T) {
		fill, empty := resolveBarChars("dots", "X", "O")
		assert.Equal(t, "X", fill)
		assert.Equal(t, "O", empty)
	})

	t.Run("unknown bar_style falls back to classic", func(t *testing.T) {
		fill, empty := resolveBarChars("unknown", "", "")
		assert.Equal(t, "\u2588", fill)
		assert.Equal(t, "\u2591", empty)
	})

	t.Run("explicit chars without bar_style", func(t *testing.T) {
		fill, empty := resolveBarChars("", "#", "-")
		assert.Equal(t, "#", fill)
		assert.Equal(t, "-", empty)
	})
}
