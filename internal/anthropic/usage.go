// Package anthropic reads what a Claude subscription actually meters, straight
// from Anthropic, rather than estimating it from token counts in a transcript.
//
//	GET /api/oauth/profile  what kind of subscription this is
//	GET /api/oauth/usage    what that subscription meters
//
// A status line renders on every keystroke-ish event, so the HTTP call never
// happens on the render path: renders read a cache file, and a stale cache is
// refreshed by a detached background process that outlives this one.
package anthropic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

const (
	fetchTimeout  = 8 * time.Second
	dirPerms      = 0o700
	filePerms     = 0o600
	lockStaleness = time.Minute
)

// baseURL is where the readings come from, and refresh is what a stale cache
// triggers. Both are variables so a test can point them somewhere harmless:
// the defaults talk to Anthropic and spawn a process, neither of which belongs
// in a test run.
var (
	baseURL = "https://api.anthropic.com/api/oauth"
	refresh = triggerRefresh
)

// TTL is how long a cached usage response is served before a refresh is
// triggered. A refresh is asynchronous, so this is not a render-path cost.
const TTL = 5 * time.Minute

// Money is an amount as the API reports it: minor units plus the exponent
// needed to scale them, so no currency is assumed to have two decimals.
type Money struct {
	AmountMinor int64  `json:"amount_minor"`
	Currency    string `json:"currency"`
	Exponent    int    `json:"exponent"`
}

// Major converts the amount to major units (cents to dollars, say).
func (m Money) Major() float64 {
	scale := 1.0
	for range m.Exponent {
		scale *= 10
	}

	return float64(m.AmountMinor) / scale
}

// Spend is the credit pool metered on usage-based seats. On plans without one
// (Pro, Max) Enabled is false and the module renders nothing.
type Spend struct {
	Used     Money   `json:"used"`
	Limit    *Money  `json:"limit"`
	Percent  float64 `json:"percent"`
	Severity string  `json:"severity"`
	Enabled  bool    `json:"enabled"`
}

// Limit is one rate-limit window (a 5-hour session, a weekly pool).
type Limit struct {
	Kind     string  `json:"kind"`
	Group    string  `json:"group"`
	Percent  float64 `json:"percent"`
	Severity string  `json:"severity"`
	ResetsAt string  `json:"resets_at"`
}

// Usage is the part of the response this status line renders.
type Usage struct {
	Spend  Spend   `json:"spend"`
	Limits []Limit `json:"limits"`

	// Age is how long ago the cache this came from was written. It is not part
	// of the API response; the module uses it to mark a reading as stale.
	Age time.Duration `json:"-"`
}

// Load returns the cached usage, refreshing it in the background when stale.
// A missing or unreadable cache is not an error the status line should print:
// the caller renders nothing and the next render picks up the refresh.
func Load() (*Usage, error) {
	path, err := cachePath()
	if err != nil {
		return nil, err
	}

	info, statErr := os.Stat(path)
	if statErr != nil || time.Since(info.ModTime()) > TTL {
		refresh()
	}

	if statErr != nil {
		return nil, fmt.Errorf("no cached usage: %w", statErr)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading usage cache: %w", err)
	}

	var usage Usage

	err = json.Unmarshal(raw, &usage)
	if err != nil {
		return nil, fmt.Errorf("parsing usage cache: %w", err)
	}

	usage.Age = time.Since(info.ModTime())

	return &usage, nil
}

// Refresh fetches usage from Anthropic and writes the cache. It is what the
// detached background process runs; nothing on the render path calls it.
func Refresh() error {
	lock, err := takeLock()
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(lock) }()

	token, err := accessToken()
	if err != nil {
		return err
	}

	body, err := get(token, "usage")
	if err != nil {
		return err
	}

	path, err := cachePath()
	if err != nil {
		return err
	}

	// Write-then-rename, so a render never sees a half-written cache.
	tmp := path + ".tmp"

	err = os.WriteFile(tmp, body, filePerms)
	if err != nil {
		return fmt.Errorf("writing usage cache: %w", err)
	}

	return os.Rename(tmp, path)
}

func get(token, endpoint string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/"+endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", endpoint, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching %s: %s", endpoint, resp.Status)
	}

	return io.ReadAll(resp.Body)
}

// accessToken reads the OAuth token Claude Code already holds. Nothing here
// writes credentials, and the token never leaves this process.
func accessToken() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	raw, err := os.ReadFile(filepath.Join(home, ".claude", ".credentials.json"))
	if err != nil {
		return "", fmt.Errorf("reading credentials: %w", err)
	}

	// The key names are Claude Code's, not ours: this file is read, never
	// written, so camelCase here is the format talking.
	var creds struct {
		OAuth struct {
			AccessToken string `json:"accessToken"` //nolint:tagliatelle // Claude Code's format
		} `json:"claudeAiOauth"` //nolint:tagliatelle // Claude Code's format
	}

	err = json.Unmarshal(raw, &creds)
	if err != nil {
		return "", fmt.Errorf("parsing credentials: %w", err)
	}

	if creds.OAuth.AccessToken == "" {
		return "", errors.New("no access token in credentials")
	}

	return creds.OAuth.AccessToken, nil
}

func stateDir() (string, error) {
	dir := os.Getenv("XDG_STATE_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		dir = filepath.Join(home, ".local", "state")
	}

	dir = filepath.Join(dir, "claude-statusline")

	// The path is the caller's own XDG_STATE_HOME or home directory. A local
	// CLI reading and writing where its user points it is the feature.
	return dir, os.MkdirAll(dir, dirPerms) //nolint:gosec // user-owned path
}

func cachePath() (string, error) {
	dir, err := stateDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "usage.json"), nil
}

// takeLock stops a burst of renders from firing parallel refreshes. A lock left
// behind by a killed process goes stale rather than wedging refreshes forever.
func takeLock() (string, error) {
	dir, err := stateDir()
	if err != nil {
		return "", err
	}

	lock := filepath.Join(dir, "refresh.lock")

	info, statErr := os.Stat(lock)
	if statErr == nil && time.Since(info.ModTime()) > lockStaleness {
		_ = os.Remove(lock)
	}

	file, err := os.OpenFile(lock, os.O_CREATE|os.O_EXCL|os.O_WRONLY, filePerms)
	if err != nil {
		return "", fmt.Errorf("refresh already running: %w", err)
	}

	_ = file.Close()

	return lock, nil
}

// triggerRefresh re-executes this binary detached, so the status line returns
// immediately and the fetch survives the render process exiting.
//
// Under `go test` it does nothing. The executable there is the package test
// binary, which ignores the unknown "refresh-usage" argument and runs the whole
// suite again - so any test that reads a stale cache would fork another full
// test run, detached from the runner, and that recursion is exponential.
// Tests in this package replace refresh outright; this guard covers the ones
// that reach Load through another package and cannot see that seam.
func triggerRefresh() {
	if testing.Testing() {
		return
	}

	exe, err := os.Executable()
	if err != nil {
		return
	}

	// context.Background() on purpose: this child has to outlive the process
	// starting it, so there is no deadline or cancellation to inherit.
	cmd := exec.CommandContext(context.Background(), exe, "refresh-usage")
	cmd.SysProcAttr = detachAttr()
	cmd.Stdin, cmd.Stdout, cmd.Stderr = nil, nil, nil

	startErr := cmd.Start()
	if startErr == nil {
		_ = cmd.Process.Release()
	}
}
