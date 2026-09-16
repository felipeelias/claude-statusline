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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	fetchTimeout  = 8 * time.Second
	dirPerms      = 0o700
	filePerms     = 0o600
	lockStaleness = time.Minute

	// maxBackoff caps what a Retry-After is allowed to buy. A server asking for
	// a day off should not silence the reading for a day.
	maxBackoff = time.Hour
	// defaultBackoff applies when a refresh fails without saying when to return.
	defaultBackoff = 5 * time.Minute
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
	path, err := CachePath()
	if err != nil {
		return nil, err
	}

	info, statErr := os.Stat(path)
	if (statErr != nil || time.Since(info.ModTime()) > TTL) && !backoffActive() {
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

	body, retryAfter, err := get(token, "usage")
	if err != nil {
		setBackoff(retryAfter)

		return err
	}

	clearBackoff()

	path, err := CachePath()
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

// get returns the body, and how long the server asked us to wait before trying
// again. The wait is meaningful even on success paths that fail later, so it is
// returned rather than folded into the error.
func get(token, endpoint string) ([]byte, time.Duration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/"+endpoint, nil)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("fetching %s: %w", endpoint, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, retryAfter(resp), fmt.Errorf("fetching %s: %s", endpoint, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)

	return body, 0, err
}

// retryAfter reads the header of that name, which Anthropic sends as seconds.
func retryAfter(resp *http.Response) time.Duration {
	seconds, err := strconv.Atoi(strings.TrimSpace(resp.Header.Get("Retry-After")))
	if err != nil || seconds <= 0 {
		return 0
	}

	return time.Duration(seconds) * time.Second
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

// accountDir is the per-account corner of the state directory. Claude Code
// separates credentials by config directory, so two sessions signed in as two
// different accounts - an Enterprise seat and a Pro one, say - have different
// config directories and therefore different readings to cache. Sharing one
// usage.json between them means whichever refreshed last wins, and a Pro
// session (no credit pool, spend.enabled false) silently blanks the Enterprise
// session's budget.
//
// The key is the config directory rather than the account the credentials
// resolve to, because this is on the render path: hashing an environment
// variable is free, while identifying the account means reading the Keychain,
// which is a process spawn on every render.
func accountDir() (string, error) {
	state, err := stateDir()
	if err != nil {
		return "", err
	}

	config, err := configDir()
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256([]byte(config))
	dir := filepath.Join(state, "accounts", hex.EncodeToString(sum[:])[:configDirHashLength])

	// The path is under the caller's own state directory.
	return dir, os.MkdirAll(dir, dirPerms)
}

// CachePath is where the usage reading for the current account is cached. It
// is exported so a test can seed a reading without rebuilding the per-account
// path by hand, and so the layout has one definition rather than four.
func CachePath() (string, error) {
	dir, err := accountDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "usage.json"), nil
}

// backoffPath holds the time before which no refresh should be attempted.
// Without it, a rate-limited endpoint is retried on every render whose cache
// has expired - which is what provokes the rate limit in the first place, and
// leaves the reading frozen for as long as the limit lasts.
func backoffPath() (string, error) {
	dir, err := accountDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "backoff"), nil
}

// backoffActive reports whether a previous failure asked us to stay away.
func backoffActive() bool {
	path, err := backoffPath()
	if err != nil {
		return false
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	until, err := time.Parse(time.RFC3339, strings.TrimSpace(string(raw)))
	if err != nil {
		return false
	}

	return time.Now().Before(until)
}

// setBackoff records when a refresh may next be attempted. A failure with no
// usable Retry-After still backs off, so a broken endpoint is not hammered.
func setBackoff(wait time.Duration) {
	if wait <= 0 {
		wait = defaultBackoff
	}

	if wait > maxBackoff {
		wait = maxBackoff
	}

	path, err := backoffPath()
	if err != nil {
		return
	}

	_ = os.WriteFile(path, []byte(time.Now().Add(wait).Format(time.RFC3339)), filePerms)
}

// clearBackoff is called after a reading arrives, so one success ends the wait.
func clearBackoff() {
	path, err := backoffPath()
	if err != nil {
		return
	}

	_ = os.Remove(path)
}

// takeLock stops a burst of renders from firing parallel refreshes. A lock left
// behind by a killed process goes stale rather than wedging refreshes forever.
func takeLock() (string, error) {
	dir, err := accountDir()
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
