package modules_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
	"github.com/felipeelias/claude-statusline/internal/modules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seedUsage writes a usage cache the credits and windows modules will read,
// aged as the test needs. Nothing here touches the network.
func seedUsage(t *testing.T, body string, age time.Duration) {
	t.Helper()

	state := t.TempDir()
	t.Setenv("XDG_STATE_HOME", state)

	dir := filepath.Join(state, "claude-statusline")
	require.NoError(t, os.MkdirAll(dir, 0o700))

	path := filepath.Join(dir, "usage.json")
	require.NoError(t, os.WriteFile(path, []byte(body), 0o600))

	modTime := time.Now().Add(-age)
	require.NoError(t, os.Chtimes(path, modTime, modTime))
}

const creditsEnabled = `{"spend": {
	"used":  {"amount_minor": 71908, "currency": "USD", "exponent": 2},
	"limit": {"amount_minor": 76900, "currency": "USD", "exponent": 2},
	"percent": 94, "severity": "critical", "enabled": true}}`

const creditsDisabled = `{"spend": {
	"used": {"amount_minor": 0, "currency": "USD", "exponent": 2},
	"limit": null, "percent": 0, "severity": "normal", "enabled": false}}`

func creditsConfig() config.Config {
	cfg := config.Default()
	cfg.Credits.Disabled = false

	return cfg
}

func TestCreditsModule_Name(t *testing.T) {
	assert.Equal(t, "credits", modules.CreditsModule{}.Name())
}

func TestCreditsModule_Render(t *testing.T) {
	t.Run("renders spend against the ceiling", func(t *testing.T) {
		seedUsage(t, creditsEnabled, 0)

		result, err := modules.CreditsModule{}.Render(input.Data{}, creditsConfig())

		require.NoError(t, err)
		assert.Contains(t, result, "$719/$769")
		assert.Contains(t, result, "94%")
	})

	t.Run("renders a progress bar", func(t *testing.T) {
		seedUsage(t, creditsEnabled, 0)

		result, err := modules.CreditsModule{}.Render(input.Data{}, creditsConfig())

		require.NoError(t, err)
		assert.Contains(t, result, "█", "94% should be a nearly full bar")
	})

	t.Run("a plan without a credit pool renders nothing", func(t *testing.T) {
		seedUsage(t, creditsDisabled, 0)

		result, err := modules.CreditsModule{}.Render(input.Data{}, creditsConfig())

		require.NoError(t, err)
		assert.Empty(t, result, "Pro and Max meter windows, not a pool")
	})

	t.Run("no cached reading renders nothing rather than an error", func(t *testing.T) {
		t.Setenv("XDG_STATE_HOME", t.TempDir())

		result, err := modules.CreditsModule{}.Render(input.Data{}, creditsConfig())

		require.NoError(t, err, "a status line has nowhere to report an error to")
		assert.Empty(t, result)
	})

	t.Run("a corrupt cache renders nothing rather than an error", func(t *testing.T) {
		seedUsage(t, "{not json", 0)

		result, err := modules.CreditsModule{}.Render(input.Data{}, creditsConfig())

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("an old reading is marked, not silently frozen", func(t *testing.T) {
		seedUsage(t, creditsEnabled, time.Hour)

		result, err := modules.CreditsModule{}.Render(input.Data{}, creditsConfig())

		require.NoError(t, err)
		assert.Contains(t, result, "⚠︎")
	})

	t.Run("a recent reading carries no marker", func(t *testing.T) {
		seedUsage(t, creditsEnabled, time.Minute)

		result, err := modules.CreditsModule{}.Render(input.Data{}, creditsConfig())

		require.NoError(t, err)
		assert.NotContains(t, result, "⚠︎")
	})

	t.Run("thresholds colour the segment by severity", func(t *testing.T) {
		seedUsage(t, creditsEnabled, 0)

		cfg := creditsConfig()
		cfg.Credits.Style = "green"
		cfg.Credits.Thresholds = []config.Threshold{{Above: 90, Style: "red"}}

		result, err := modules.CreditsModule{}.Render(input.Data{}, cfg)

		require.NoError(t, err)
		assert.Contains(t, result, "\x1b[31m", "94% is above the 90% threshold, so red")
	})

	t.Run("a custom format reaches every field", func(t *testing.T) {
		seedUsage(t, creditsEnabled, 0)

		cfg := creditsConfig()
		cfg.Credits.Style = ""
		cfg.Credits.Thresholds = nil
		cfg.Credits.Format = `{{.Currency}} {{printf "%.2f" .Used}} of {{printf "%.2f" .Limit}}`

		result, err := modules.CreditsModule{}.Render(input.Data{}, cfg)

		require.NoError(t, err)
		assert.Equal(t, "USD 719.08 of 769.00", result)
	})

	t.Run("a broken format is reported", func(t *testing.T) {
		seedUsage(t, creditsEnabled, 0)

		cfg := creditsConfig()
		cfg.Credits.Format = "{{.Nope"

		_, err := modules.CreditsModule{}.Render(input.Data{}, cfg)

		require.Error(t, err)
	})
}
