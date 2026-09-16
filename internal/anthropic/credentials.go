package anthropic

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Where Claude Code keeps the OAuth credentials it is currently using, and why
// this is not simply "~/.claude/.credentials.json".
//
// Claude Code isolates credentials per config directory. With CLAUDE_CONFIG_DIR
// unset it uses ~/.claude and the macOS Keychain item "Claude Code-credentials";
// with it set, the item is suffixed with the first eight hex digits of the
// SHA-256 of that directory path. So a machine running one Claude Code session
// on an Enterprise seat and another on Pro has two live tokens at once, and
// which one is "current" is a property of the session this process was spawned
// from - Claude Code exports CLAUDE_CONFIG_DIR into the status line child.
//
// Reading the path from the environment is what keeps this independent of
// whatever manages those directories. Account switchers (clauth and friends)
// work by pointing CLAUDE_CONFIG_DIR at a profile and symlinking the files
// underneath, so following the environment variable follows the switch for
// free, with no knowledge of the switcher.
//
// The on-disk file is read as a fallback rather than the primary source
// because it goes stale: Claude Code refreshes the access token into the
// Keychain, and a symlinked or copied .credentials.json can sit expired for
// days while the session it belongs to is working perfectly.

// keychainService is the macOS Keychain item holding the credentials for the
// default config directory. Per-directory items append a hash of the path.
const keychainService = "Claude Code-credentials"

// configDirHashLength is how much of the SHA-256 Claude Code puts in the
// Keychain service name.
const configDirHashLength = 8

// keychainLookup reads a Keychain item by service name. It is a variable so a
// test can stub it: the real one would otherwise return the developer's own
// live credentials in the middle of a test run.
var keychainLookup = readKeychain

// credentials is the part of Claude Code's credential blob this package needs.
// The key names are Claude Code's, not ours: these files are read, never
// written, so camelCase here is the format talking.
type credentials struct {
	OAuth struct {
		AccessToken  string `json:"accessToken"`  //nolint:tagliatelle // Claude Code's format
		RefreshToken string `json:"refreshToken"` //nolint:tagliatelle // Claude Code's format
		// ExpiresAt is milliseconds since the epoch, and zero when absent.
		ExpiresAt int64 `json:"expiresAt"` //nolint:tagliatelle // Claude Code's format
		// SubscriptionType is what the credential itself claims the plan is.
		// It is a useful hint but not authoritative; /oauth/profile is.
		SubscriptionType string `json:"subscriptionType"` //nolint:tagliatelle // Claude Code's format
	} `json:"claudeAiOauth"` //nolint:tagliatelle // Claude Code's format
}

// expired reports whether the access token is past its stated expiry. Missing
// expiry counts as usable: some credential files omit it, and a token that
// might work beats no token at all.
func (c credentials) expired() bool {
	if c.OAuth.ExpiresAt == 0 {
		return false
	}

	return time.Now().After(time.UnixMilli(c.OAuth.ExpiresAt))
}

// configDir is the Claude Code configuration directory this process belongs to.
func configDir() (string, error) {
	dir := os.Getenv("CLAUDE_CONFIG_DIR")
	if dir != "" {
		return dir, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".claude"), nil
}

// keychainServiceFor returns the Keychain service name Claude Code uses for a
// config directory: the bare name for the default one, hash-suffixed otherwise.
func keychainServiceFor(dir string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	if dir == filepath.Join(home, ".claude") {
		return keychainService, nil
	}

	sum := sha256.Sum256([]byte(dir))

	return keychainService + "-" + hex.EncodeToString(sum[:])[:configDirHashLength], nil
}

// keychainTimeout caps the Keychain lookup. It normally answers in
// milliseconds, but a locked or prompting keychain can hang indefinitely, and
// nothing here is worth blocking a status line refresh on.
const keychainTimeout = 2 * time.Second

// readKeychain returns the raw credential blob stored under a service name, or
// empty if there is no such item. A missing item is the normal case on Linux
// and on machines that keep credentials in a file, so it is not an error.
func readKeychain(service string) []byte {
	ctx, cancel := context.WithTimeout(context.Background(), keychainTimeout)
	defer cancel()

	// Fixed argv; the service name is this package's own construction.
	out, err := exec.CommandContext(ctx, "security", "find-generic-password", "-s", service, "-w").Output()
	if err != nil {
		return nil
	}

	return []byte(strings.TrimSpace(string(out)))
}

// parseCredentials decodes a credential blob, rejecting one with no token.
func parseCredentials(raw []byte) (credentials, error) {
	var creds credentials

	if len(raw) == 0 {
		return creds, errors.New("empty credentials")
	}

	err := json.Unmarshal(raw, &creds)
	if err != nil {
		return creds, fmt.Errorf("parsing credentials: %w", err)
	}

	if creds.OAuth.AccessToken == "" {
		return creds, errors.New("no access token in credentials")
	}

	return creds, nil
}

// currentCredentials returns the credentials of the account this process's
// Claude Code session is signed in as.
//
// The Keychain is consulted first because it is where Claude Code writes
// refreshed tokens; the file is the fallback for platforms and setups that do
// not use it. An expired candidate is remembered but kept back, so a live
// token anywhere wins - and if everything on offer is expired we return the
// first one rather than nothing, leaving the caller to surface a stale reading
// instead of an empty status line.
func currentCredentials() (credentials, error) {
	var zero credentials

	dir, err := configDir()
	if err != nil {
		return zero, err
	}

	service, err := keychainServiceFor(dir)
	if err != nil {
		return zero, err
	}

	var (
		firstExpired credentials
		haveExpired  bool
		problems     []string
	)

	var none credentials

	consider := func(raw []byte, source string) (credentials, bool) {
		creds, parseErr := parseCredentials(raw)
		if parseErr != nil {
			problems = append(problems, source+": "+parseErr.Error())

			return none, false
		}

		if creds.expired() {
			if !haveExpired {
				firstExpired, haveExpired = creds, true
			}

			problems = append(problems, source+": token expired")

			return none, false
		}

		return creds, true
	}

	if creds, ok := consider(keychainLookup(service), "keychain "+service); ok {
		return creds, nil
	}

	path := filepath.Join(dir, ".credentials.json")

	raw, readErr := os.ReadFile(path)
	if readErr != nil {
		problems = append(problems, path+": "+readErr.Error())
	} else if creds, ok := consider(raw, path); ok {
		return creds, nil
	}

	if haveExpired {
		return firstExpired, nil
	}

	return zero, fmt.Errorf("no usable credentials (%s)", strings.Join(problems, "; "))
}

// accessToken reads the OAuth token Claude Code already holds. Nothing here
// writes credentials, and the token never leaves this process.
func accessToken() (string, error) {
	creds, err := currentCredentials()
	if err != nil {
		return "", err
	}

	return creds.OAuth.AccessToken, nil
}
