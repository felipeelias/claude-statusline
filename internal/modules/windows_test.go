package modules_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
	"github.com/felipeelias/claude-statusline/internal/modules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const twoWindows = `{"limits": [
	{"kind": "session",    "percent": 9, "severity": "normal",
	 "resets_at": "2026-09-14T23:30:00+00:00"},
	{"kind": "weekly_all", "percent": 2, "severity": "normal",
	 "resets_at": "2026-09-20T21:00:00+00:00"}
]}`

func windowsConfig() config.Config {
	cfg := config.Default()
	cfg.Windows.Disabled = false

	return cfg
}

func TestWindowsModule_Name(t *testing.T) {
	assert.Equal(t, "windows", modules.WindowsModule{}.Name())
}

func TestWindowsModule_Render(t *testing.T) {
	t.Run("renders every window with its short label", func(t *testing.T) {
		seedUsage(t, twoWindows, 0)

		result, err := modules.WindowsModule{}.Render(input.Data{}, windowsConfig())

		require.NoError(t, err)
		assert.Contains(t, result, "5h 9%")
		assert.Contains(t, result, "wk 2%")
	})

	t.Run("joins them with the configured separator", func(t *testing.T) {
		seedUsage(t, twoWindows, 0)

		cfg := windowsConfig()
		cfg.Windows.Separator = " // "

		result, err := modules.WindowsModule{}.Render(input.Data{}, cfg)

		require.NoError(t, err)
		assert.Contains(t, result, " // ")
	})

	t.Run("an unknown window kind still renders", func(t *testing.T) {
		seedUsage(t, `{"limits": [{"kind": "weekly_opus", "percent": 30}]}`, 0)

		result, err := modules.WindowsModule{}.Render(input.Data{}, windowsConfig())

		require.NoError(t, err)
		assert.Contains(t, result, "opus 30%",
			"a window Anthropic adds later must not disappear from the line")
	})

	t.Run("no windows renders nothing", func(t *testing.T) {
		seedUsage(t, `{"limits": []}`, 0)

		result, err := modules.WindowsModule{}.Render(input.Data{}, windowsConfig())

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("no cached reading renders nothing rather than an error", func(t *testing.T) {
		t.Setenv("XDG_STATE_HOME", t.TempDir())

		result, err := modules.WindowsModule{}.Render(input.Data{}, windowsConfig())

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("the worst window drives the colour", func(t *testing.T) {
		seedUsage(t, `{"limits": [
			{"kind": "session", "percent": 5},
			{"kind": "weekly_all", "percent": 95}
		]}`, 0)

		cfg := windowsConfig()
		cfg.Windows.Style = "green"
		cfg.Windows.Thresholds = []config.Threshold{{Above: 90, Style: "red"}}

		result, err := modules.WindowsModule{}.Render(input.Data{}, cfg)

		require.NoError(t, err)
		assert.Contains(t, result, "\x1b[31m",
			"a quiet 5h window must not mask a nearly exhausted weekly one")
	})

	t.Run("an old reading is marked once, not per window", func(t *testing.T) {
		seedUsage(t, twoWindows, time.Hour)

		result, err := modules.WindowsModule{}.Render(input.Data{}, windowsConfig())

		require.NoError(t, err)
		assert.Equal(t, 1, countMarkers(result), "one marker for the reading, not one per window")
	})

	t.Run("a reset later today shows a clock time", func(t *testing.T) {
		resetsAt := time.Now().Add(3 * time.Hour)
		if resetsAt.YearDay() != time.Now().YearDay() {
			t.Skip("run crosses midnight; the weekday branch is covered below")
		}

		seedUsage(t, fmt.Sprintf(`{"limits": [{"kind": "session", "percent": 9, "resets_at": %q}]}`,
			resetsAt.Format(time.RFC3339)), 0)

		cfg := windowsConfig()
		cfg.Windows.Format = "{{.Resets}}"

		result, err := modules.WindowsModule{}.Render(input.Data{}, cfg)

		require.NoError(t, err)
		assert.Contains(t, result, resetsAt.Format("15:04"))
	})

	t.Run("a reset on another day shows a weekday", func(t *testing.T) {
		resetsAt := time.Now().Add(72 * time.Hour)

		seedUsage(t, fmt.Sprintf(`{"limits": [{"kind": "weekly_all", "percent": 2, "resets_at": %q}]}`,
			resetsAt.Format(time.RFC3339)), 0)

		cfg := windowsConfig()
		cfg.Windows.Format = "{{.Resets}}"

		result, err := modules.WindowsModule{}.Render(input.Data{}, cfg)

		require.NoError(t, err)
		assert.Contains(t, result, resetsAt.Format("Mon"))
	})

	t.Run("a missing reset timestamp renders empty, not a zero date", func(t *testing.T) {
		seedUsage(t, `{"limits": [{"kind": "session", "percent": 9}]}`, 0)

		cfg := windowsConfig()
		cfg.Windows.Style = ""
		cfg.Windows.Thresholds = nil
		cfg.Windows.Format = "[{{.Resets}}]"

		result, err := modules.WindowsModule{}.Render(input.Data{}, cfg)

		require.NoError(t, err)
		assert.Equal(t, "[]", result)
	})

	t.Run("a broken format is reported", func(t *testing.T) {
		seedUsage(t, twoWindows, 0)

		cfg := windowsConfig()
		cfg.Windows.Format = "{{.Nope"

		_, err := modules.WindowsModule{}.Render(input.Data{}, cfg)

		require.Error(t, err)
	})
}

func countMarkers(s string) int {
	count := 0
	for _, r := range s {
		if r == '⚠' {
			count++
		}
	}

	return count
}
