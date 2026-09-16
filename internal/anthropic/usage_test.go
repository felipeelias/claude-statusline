package anthropic_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/frank-bee/claude-statusline/internal/anthropic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleResponse = `{
  "spend": {
    "used":  {"amount_minor": 71908, "currency": "USD", "exponent": 2},
    "limit": {"amount_minor": 76900, "currency": "USD", "exponent": 2},
    "percent": 94, "severity": "critical", "enabled": true
  },
  "limits": [
    {"kind": "session",    "percent": 9, "severity": "normal",
     "resets_at": "2026-09-14T23:30:00+00:00"},
    {"kind": "weekly_all", "percent": 2, "severity": "normal",
     "resets_at": "2026-09-20T21:00:00+00:00"}
  ]
}`

// seedCache points the package at a temporary state directory and writes a
// cache file into it, aged as the test needs.
func seedCache(t *testing.T, body string, age time.Duration) string {
	t.Helper()

	state := t.TempDir()
	t.Setenv("XDG_STATE_HOME", state)
	t.Setenv("CLAUDE_CONFIG_DIR", "")

	path := cachePath(t)
	require.NoError(t, os.WriteFile(path, []byte(body), 0o600))

	modTime := time.Now().Add(-age)
	require.NoError(t, os.Chtimes(path, modTime, modTime))

	return path
}

// cachePath asks the package where the reading for the current account lives.
// The layout is per-account, so a test that builds the path itself would be
// asserting against a directory the code no longer uses.
func cachePath(t *testing.T) string {
	t.Helper()

	path, err := anthropic.CachePath()
	require.NoError(t, err)

	return path
}

// writeCache seeds a cache of a given age into an existing state directory,
// where seedCache would create a fresh one and lose the backoff written beside
// it.
func writeCache(t *testing.T, _, body string, age time.Duration) {
	t.Helper()

	path := cachePath(t)
	require.NoError(t, os.WriteFile(path, []byte(body), 0o600))

	modTime := time.Now().Add(-age)
	require.NoError(t, os.Chtimes(path, modTime, modTime))
}

func TestMoney_Major(t *testing.T) {
	tests := map[string]struct {
		money    anthropic.Money
		expected float64
	}{
		"cents":          {anthropic.Money{AmountMinor: 71908, Exponent: 2}, 719.08},
		"whole units":    {anthropic.Money{AmountMinor: 42, Exponent: 0}, 42},
		"three decimals": {anthropic.Money{AmountMinor: 1500, Exponent: 3}, 1.5},
		"zero":           {anthropic.Money{}, 0},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.InDelta(t, tc.expected, tc.money.Major(), 0.0001)
		})
	}
}

func TestLoad(t *testing.T) {
	t.Run("parses a fresh cache", func(t *testing.T) {
		seedCache(t, sampleResponse, 0)
		anthropic.SuppressRefresh(t)

		usage, err := anthropic.Load()

		require.NoError(t, err)
		assert.True(t, usage.Spend.Enabled)
		assert.InDelta(t, 94.0, usage.Spend.Percent, 0.001)
		assert.InDelta(t, 719.08, usage.Spend.Used.Major(), 0.001)
		assert.InDelta(t, 769.0, usage.Spend.Limit.Major(), 0.001)
		assert.Len(t, usage.Limits, 2)
		assert.Equal(t, "session", usage.Limits[0].Kind)
	})

	t.Run("reports the age of the reading", func(t *testing.T) {
		seedCache(t, sampleResponse, 2*time.Hour)
		anthropic.SuppressRefresh(t)

		usage, err := anthropic.Load()

		require.NoError(t, err)
		assert.Greater(t, usage.Age, time.Hour)
	})

	t.Run("a fresh cache triggers no refresh", func(t *testing.T) {
		seedCache(t, sampleResponse, 0)
		calls := anthropic.SuppressRefresh(t)

		_, err := anthropic.Load()

		require.NoError(t, err)
		assert.Equal(t, 0, *calls)
	})

	t.Run("a stale cache triggers a refresh but still renders", func(t *testing.T) {
		seedCache(t, sampleResponse, anthropic.TTL+time.Minute)
		calls := anthropic.SuppressRefresh(t)

		usage, err := anthropic.Load()

		require.NoError(t, err)
		assert.Equal(t, 1, *calls, "stale cache should trigger exactly one refresh")
		assert.InDelta(t, 94.0, usage.Spend.Percent, 0.001,
			"the stale reading is still served while the refresh runs")
	})

	t.Run("a missing cache errors and triggers a refresh", func(t *testing.T) {
		t.Setenv("XDG_STATE_HOME", t.TempDir())
		calls := anthropic.SuppressRefresh(t)

		_, err := anthropic.Load()

		require.Error(t, err)
		assert.Equal(t, 1, *calls)
	})

	t.Run("a corrupt cache errors rather than panicking", func(t *testing.T) {
		seedCache(t, "{not json", 0)
		anthropic.SuppressRefresh(t)

		_, err := anthropic.Load()

		require.ErrorContains(t, err, "parsing usage cache")
	})
}

func TestRefresh(t *testing.T) {
	// Refresh reads the credentials of whichever account Claude Code is signed
	// in as, so these tests give it a home and a Keychain of their own. Without
	// the Keychain suppressed, a temporary HOME still resolves the default
	// Keychain item, which on a developer machine is a real live token.
	withCredentials := func(t *testing.T) {
		t.Helper()

		anthropic.SuppressKeychain(t)

		home := t.TempDir()
		t.Setenv("HOME", home)
		t.Setenv("CLAUDE_CONFIG_DIR", "")
		require.NoError(t, os.MkdirAll(filepath.Join(home, ".claude"), 0o700))
		require.NoError(t, os.WriteFile(
			filepath.Join(home, ".claude", ".credentials.json"),
			[]byte(`{"claudeAiOauth":{"accessToken":"test-token"}}`), 0o600))
	}

	t.Run("writes what the API returns", func(t *testing.T) {
		var gotAuth, gotPath string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotAuth = r.Header.Get("Authorization")
			gotPath = r.URL.Path

			_, _ = w.Write([]byte(sampleResponse))
		}))
		defer server.Close()

		state := t.TempDir()
		t.Setenv("XDG_STATE_HOME", state)
		withCredentials(t)
		anthropic.SetBaseURL(t, server.URL)

		require.NoError(t, anthropic.Refresh())

		assert.Equal(t, "Bearer test-token", gotAuth, "the token goes in the header, nowhere else")
		assert.Equal(t, "/usage", gotPath)

		written, err := os.ReadFile(cachePath(t))
		require.NoError(t, err)
		assert.JSONEq(t, sampleResponse, string(written))
	})

	t.Run("a rate-limited refresh backs off for as long as the server asks", func(t *testing.T) {
		attempts := 0

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			attempts++

			writer.Header().Set("Retry-After", "3368")
			writer.WriteHeader(http.StatusTooManyRequests)
		}))
		defer server.Close()

		state := t.TempDir()
		t.Setenv("XDG_STATE_HOME", state)
		withCredentials(t)
		anthropic.SetBaseURL(t, server.URL)

		require.Error(t, anthropic.Refresh())
		assert.Equal(t, 1, attempts)

		// A stale cache would normally trigger a refresh on the next render.
		// While the server is asking us to stay away, it must not: retrying on
		// every render is what sustains the rate limit and freezes the reading.
		writeCache(t, state, sampleResponse, time.Hour)

		calls := anthropic.SuppressRefresh(t)

		_, err := anthropic.Load()
		require.NoError(t, err)
		assert.Equal(t, 0, *calls, "no refresh may be attempted while backing off")
	})

	t.Run("a failure with no Retry-After still backs off", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			writer.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		state := t.TempDir()
		t.Setenv("XDG_STATE_HOME", state)
		withCredentials(t)
		anthropic.SetBaseURL(t, server.URL)

		require.Error(t, anthropic.Refresh())

		writeCache(t, state, sampleResponse, time.Hour)

		calls := anthropic.SuppressRefresh(t)

		_, err := anthropic.Load()
		require.NoError(t, err)
		assert.Equal(t, 0, *calls)
	})

	t.Run("one success ends the backoff", func(t *testing.T) {
		fail := true

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			if fail {
				writer.Header().Set("Retry-After", "1")
				writer.WriteHeader(http.StatusTooManyRequests)

				return
			}

			_, _ = writer.Write([]byte(sampleResponse))
		}))
		defer server.Close()

		state := t.TempDir()
		t.Setenv("XDG_STATE_HOME", state)
		withCredentials(t)
		anthropic.SetBaseURL(t, server.URL)

		require.Error(t, anthropic.Refresh())

		fail = false
		require.NoError(t, anthropic.Refresh())

		writeCache(t, state, sampleResponse, time.Hour)

		calls := anthropic.SuppressRefresh(t)

		_, err := anthropic.Load()
		require.NoError(t, err)
		assert.Equal(t, 1, *calls, "a successful reading clears the wait")
	})

	t.Run("the cache file is not world-readable", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(sampleResponse))
		}))
		defer server.Close()

		state := t.TempDir()
		t.Setenv("XDG_STATE_HOME", state)
		withCredentials(t)
		anthropic.SetBaseURL(t, server.URL)

		require.NoError(t, anthropic.Refresh())

		info, err := os.Stat(cachePath(t))
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	})

	t.Run("an API error leaves the previous reading intact", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		path := seedCache(t, sampleResponse, 0)
		withCredentials(t)
		anthropic.SetBaseURL(t, server.URL)

		err := anthropic.Refresh()

		require.ErrorContains(t, err, "401")

		kept, readErr := os.ReadFile(path)
		require.NoError(t, readErr)
		assert.JSONEq(t, sampleResponse, string(kept),
			"a failed refresh must not destroy the last good reading")
	})

	t.Run("missing credentials are an error, not a panic", func(t *testing.T) {
		anthropic.SuppressKeychain(t)
		t.Setenv("XDG_STATE_HOME", t.TempDir())
		t.Setenv("HOME", t.TempDir())
		t.Setenv("CLAUDE_CONFIG_DIR", "")

		require.ErrorContains(t, anthropic.Refresh(), "no usable credentials")
	})
}
