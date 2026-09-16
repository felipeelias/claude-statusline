package anthropic //nolint:testpackage // the decision under test is unexported by design

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Internal tests, unlike the rest of this package's suite. What is worth
// pinning here is which credential store gets picked and how the Keychain
// service name is derived, and neither is reachable from outside: the type is
// unexported and so is every step of the decision. Exporting them only to
// test them would widen the package's surface for no caller's benefit.

// stubKeychain replaces the Keychain with a map of service name to blob.
func stubKeychain(t *testing.T, items map[string]string) {
	t.Helper()

	previous := keychainLookup
	keychainLookup = func(service string) []byte {
		blob, ok := items[service]
		if !ok {
			return nil
		}

		return []byte(blob)
	}

	t.Cleanup(func() { keychainLookup = previous })
}

// blob builds a credential file body. An expiresAt of zero is left out
// entirely, which is how some real credential files arrive.
func blob(token string, expiresAt int64) string {
	raw := `{"claudeAiOauth":{"accessToken":"` + token + `","refreshToken":"r-` + token + `"`
	if expiresAt != 0 {
		raw += `,"expiresAt":` + strconv.FormatInt(expiresAt, 10)
	}

	return raw + `}}`
}

func future() int64 { return time.Now().Add(time.Hour).UnixMilli() }
func past() int64   { return time.Now().Add(-time.Hour).UnixMilli() }

func TestKeychainServiceFor(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	t.Run("the default config directory uses the bare service name", func(t *testing.T) {
		service, err := keychainServiceFor(filepath.Join(home, ".claude"))

		require.NoError(t, err)
		assert.Equal(t, "Claude Code-credentials", service)
	})

	t.Run("any other directory is hash-suffixed", func(t *testing.T) {
		service, err := keychainServiceFor("/somewhere/else")

		require.NoError(t, err)
		assert.NotEqual(t, "Claude Code-credentials", service)
		assert.Regexp(t, `^Claude Code-credentials-[0-9a-f]{8}$`, service)
	})

	t.Run("the suffix matches what Claude Code actually stores", func(t *testing.T) {
		// Pinned against a real machine: Claude Code keys the Keychain item on
		// the first eight hex digits of the SHA-256 of the config directory
		// path. If this drifts, credential lookup silently finds nothing and
		// every Anthropic-backed segment goes blank.
		service, err := keychainServiceFor("/Users/frankb-mac-laptop/.clauth/profiles/w/runtime-32122-0")

		require.NoError(t, err)
		assert.Equal(t, "Claude Code-credentials-8760b835", service)
	})

	t.Run("two config directories never share an item", func(t *testing.T) {
		first, err := keychainServiceFor("/profiles/enterprise")
		require.NoError(t, err)

		second, err := keychainServiceFor("/profiles/pro")
		require.NoError(t, err)

		assert.NotEqual(t, first, second, "an Enterprise seat and a Pro one must not collide")
	})
}

func TestConfigDir(t *testing.T) {
	t.Run("follows CLAUDE_CONFIG_DIR when Claude Code sets it", func(t *testing.T) {
		t.Setenv("CLAUDE_CONFIG_DIR", "/some/profile")

		dir, err := configDir()

		require.NoError(t, err)
		assert.Equal(t, "/some/profile", dir)
	})

	t.Run("falls back to ~/.claude", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		t.Setenv("CLAUDE_CONFIG_DIR", "")

		dir, err := configDir()

		require.NoError(t, err)
		assert.Equal(t, filepath.Join(home, ".claude"), dir)
	})
}

func TestCredentialsExpired(t *testing.T) {
	t.Run("a missing expiry counts as usable", func(t *testing.T) {
		creds, err := parseCredentials([]byte(blob("t", 0)))

		require.NoError(t, err)
		assert.False(t, creds.expired(), "a file without expiry is still worth trying")
	})

	t.Run("a past expiry is expired", func(t *testing.T) {
		creds, err := parseCredentials([]byte(blob("t", past())))

		require.NoError(t, err)
		assert.True(t, creds.expired())
	})

	t.Run("a future expiry is not", func(t *testing.T) {
		creds, err := parseCredentials([]byte(blob("t", future())))

		require.NoError(t, err)
		assert.False(t, creds.expired())
	})
}

func TestParseCredentials(t *testing.T) {
	t.Run("nothing at all is an error", func(t *testing.T) {
		_, err := parseCredentials(nil)

		require.ErrorContains(t, err, "empty credentials")
	})

	t.Run("a blob with no token is an error", func(t *testing.T) {
		_, err := parseCredentials([]byte(`{"claudeAiOauth":{}}`))

		require.ErrorContains(t, err, "no access token")
	})

	t.Run("junk is an error rather than a panic", func(t *testing.T) {
		_, err := parseCredentials([]byte("{not json"))

		require.ErrorContains(t, err, "parsing credentials")
	})
}

// writeCredentialsFile puts a credential blob where the config directory
// expects one.
func writeCredentialsFile(t *testing.T, dir, body string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(dir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".credentials.json"), []byte(body), 0o600))
}

func TestCurrentCredentials(t *testing.T) {
	t.Run("reads the Keychain item for this session's config directory", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("CLAUDE_CONFIG_DIR", dir)

		service, err := keychainServiceFor(dir)
		require.NoError(t, err)

		stubKeychain(t, map[string]string{service: blob("enterprise-token", future())})

		creds, err := currentCredentials()

		require.NoError(t, err)
		assert.Equal(t, "enterprise-token", creds.OAuth.AccessToken)
	})

	t.Run("switching config directory switches account", func(t *testing.T) {
		enterprise, pro := t.TempDir(), t.TempDir()

		enterpriseService, err := keychainServiceFor(enterprise)
		require.NoError(t, err)

		proService, err := keychainServiceFor(pro)
		require.NoError(t, err)

		stubKeychain(t, map[string]string{
			enterpriseService: blob("enterprise-token", future()),
			proService:        blob("pro-token", future()),
		})

		t.Setenv("CLAUDE_CONFIG_DIR", enterprise)
		creds, err := currentCredentials()
		require.NoError(t, err)
		assert.Equal(t, "enterprise-token", creds.OAuth.AccessToken)

		t.Setenv("CLAUDE_CONFIG_DIR", pro)
		creds, err = currentCredentials()
		require.NoError(t, err)
		assert.Equal(t, "pro-token", creds.OAuth.AccessToken,
			"the account follows the session, with no knowledge of what manages the directories")
	})

	t.Run("a live Keychain token beats a stale file", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("CLAUDE_CONFIG_DIR", dir)
		writeCredentialsFile(t, dir, blob("stale-file-token", past()))

		service, err := keychainServiceFor(dir)
		require.NoError(t, err)

		stubKeychain(t, map[string]string{service: blob("live-token", future())})

		creds, err := currentCredentials()

		require.NoError(t, err)
		assert.Equal(t, "live-token", creds.OAuth.AccessToken,
			"Claude Code refreshes into the Keychain; a symlinked file can sit expired for days")
	})

	t.Run("falls back to the file when the Keychain has nothing", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("CLAUDE_CONFIG_DIR", dir)
		writeCredentialsFile(t, dir, blob("file-token", future()))
		stubKeychain(t, nil)

		creds, err := currentCredentials()

		require.NoError(t, err)
		assert.Equal(t, "file-token", creds.OAuth.AccessToken)
	})

	t.Run("a live file beats an expired Keychain item", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("CLAUDE_CONFIG_DIR", dir)
		writeCredentialsFile(t, dir, blob("file-token", future()))

		service, err := keychainServiceFor(dir)
		require.NoError(t, err)

		stubKeychain(t, map[string]string{service: blob("expired-keychain", past())})

		creds, err := currentCredentials()

		require.NoError(t, err)
		assert.Equal(t, "file-token", creds.OAuth.AccessToken)
	})

	t.Run("everything expired still returns a token rather than nothing", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("CLAUDE_CONFIG_DIR", dir)
		writeCredentialsFile(t, dir, blob("stale-file", past()))

		service, err := keychainServiceFor(dir)
		require.NoError(t, err)

		stubKeychain(t, map[string]string{service: blob("stale-keychain", past())})

		creds, err := currentCredentials()

		require.NoError(t, err)
		assert.Equal(t, "stale-keychain", creds.OAuth.AccessToken,
			"a stale reading with a warning marker beats an empty status line")
	})

	t.Run("no credentials anywhere says where it looked", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("CLAUDE_CONFIG_DIR", dir)
		stubKeychain(t, nil)

		_, err := currentCredentials()

		require.ErrorContains(t, err, "no usable credentials")
		assert.Contains(t, err.Error(), ".credentials.json", "the error names the paths tried")
	})
}

func TestAccountDirIsolatesAccounts(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())

	t.Setenv("CLAUDE_CONFIG_DIR", "/profiles/enterprise")
	enterprise, err := CachePath()
	require.NoError(t, err)

	t.Setenv("CLAUDE_CONFIG_DIR", "/profiles/pro")
	pro, err := CachePath()
	require.NoError(t, err)

	assert.NotEqual(t, enterprise, pro,
		"one shared usage.json lets a Pro session blank an Enterprise session's budget")
}
