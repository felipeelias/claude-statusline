package usage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const cacheDirPerms = 0750

// cacheEntry is the on-disk format for cached usage data.
type cacheEntry struct {
	Data       *UsageData `json:"data"`
	FetchedAt  time.Time  `json:"fetched_at"`
	ErrorUntil *time.Time `json:"error_until,omitempty"`
}

// cacheState describes how fresh the cached data is.
type cacheState int

const (
	cacheFresh   cacheState = iota // within TTL
	cacheStale                     // within 2x TTL (grace period to avoid latency spikes)
	cacheExpired                   // beyond 2x TTL or missing
)

// cache reads and writes usage data to a JSON file on disk.
type cache struct {
	path string
}

// defaultCachePath returns ~/.cache/claude-statusline/usage.json.
func defaultCachePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return filepath.Join(home, ".cache", "claude-statusline", "usage.json")
}

// read loads the cache entry from disk. Returns nil if the file is missing or corrupt.
func (c *cache) read() *cacheEntry {
	data, err := os.ReadFile(c.path)
	if err != nil {
		return nil
	}

	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil
	}

	return &entry
}

// write atomically persists a cache entry to disk.
func (c *cache) write(entry *cacheEntry) error {
	if err := os.MkdirAll(filepath.Dir(c.path), cacheDirPerms); err != nil {
		return err
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	tmp := c.path + ".tmp"

	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}

	return os.Rename(tmp, c.path)
}

// state returns how fresh the cached data is relative to the given TTL.
func (c *cache) state(entry *cacheEntry, ttl time.Duration) cacheState {
	if entry == nil || entry.Data == nil {
		return cacheExpired
	}

	age := time.Since(entry.FetchedAt)

	if age <= ttl {
		return cacheFresh
	}

	if age <= ttl*2 {
		return cacheStale
	}

	return cacheExpired
}

// inErrorBackoff returns true if the cache is in an error backoff period.
func (c *cache) inErrorBackoff(entry *cacheEntry) bool {
	if entry == nil || entry.ErrorUntil == nil {
		return false
	}

	return time.Now().Before(*entry.ErrorUntil)
}

// tryLock attempts to acquire an exclusive fetch lock using atomic mkdir.
// Returns true if the lock was acquired, false if another process holds it.
// Stale locks (older than maxAge) from crashed processes are cleaned up
// so the next render cycle can acquire cleanly.
func (c *cache) tryLock(maxAge time.Duration) bool {
	lockDir := c.path + ".lock"

	if os.Mkdir(lockDir, cacheDirPerms) == nil {
		return true
	}

	// Lock exists - clean up if stale, but don't retry.
	// The next render cycle will acquire it via the mkdir above.
	info, err := os.Stat(lockDir)
	if err == nil && time.Since(info.ModTime()) >= maxAge {
		_ = os.Remove(lockDir)
	}

	return false
}

// unlock releases the fetch lock.
func (c *cache) unlock() {
	_ = os.Remove(c.path + ".lock")
}

// writeError records an error backoff period in the cache.
// Preserves existing data so stale results can still be served.
func (c *cache) writeError(existing *cacheEntry, backoff time.Duration) {
	until := time.Now().Add(backoff)

	entry := &cacheEntry{ErrorUntil: &until}
	if existing != nil {
		entry.Data = existing.Data
		entry.FetchedAt = existing.FetchedAt
	}

	_ = c.write(entry)
}
