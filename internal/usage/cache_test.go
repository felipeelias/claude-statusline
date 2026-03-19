package usage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testCache(t *testing.T) *cache {
	t.Helper()

	return &cache{path: filepath.Join(t.TempDir(), "usage.json")}
}

func testEntry(age time.Duration) *cacheEntry {
	return &cacheEntry{
		Data: &UsageData{
			FiveHour: Window{Utilisation: 42, ResetsAt: "2025-01-15T10:00:00Z"},
			SevenDay: Window{Utilisation: 15, ResetsAt: "2025-01-20T00:00:00Z"},
		},
		FetchedAt: time.Now().Add(-age),
	}
}

func TestCache_ReadWrite(t *testing.T) {
	t.Run("read missing file returns nil", func(t *testing.T) {
		c := testCache(t)
		assert.Nil(t, c.read())
	})

	t.Run("round-trip", func(t *testing.T) {
		c := testCache(t)
		entry := testEntry(0)

		require.NoError(t, c.write(entry))

		got := c.read()
		require.NotNil(t, got)
		assert.InDelta(t, 42, got.Data.FiveHour.Utilisation, 0.1)
		assert.InDelta(t, 15, got.Data.SevenDay.Utilisation, 0.1)
	})

	t.Run("corrupt file returns nil", func(t *testing.T) {
		c := testCache(t)
		require.NoError(t, os.WriteFile(c.path, []byte("not json"), 0600))

		assert.Nil(t, c.read())
	})

	t.Run("creates parent directory", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "nested", "dir")
		c := &cache{path: filepath.Join(dir, "usage.json")}

		require.NoError(t, c.write(testEntry(0)))

		got := c.read()
		require.NotNil(t, got)
	})
}

func TestCache_State(t *testing.T) {
	ttl := 120 * time.Second

	tests := []struct {
		name     string
		entry    *cacheEntry
		expected cacheState
	}{
		{name: "nil entry", entry: nil, expected: cacheExpired},
		{name: "nil data", entry: &cacheEntry{FetchedAt: time.Now()}, expected: cacheExpired},
		{name: "fresh", entry: testEntry(30 * time.Second), expected: cacheFresh},
		{name: "just under TTL", entry: testEntry(ttl - time.Second), expected: cacheFresh},
		{name: "stale", entry: testEntry(150 * time.Second), expected: cacheStale},
		{name: "just under 2x TTL", entry: testEntry(ttl*2 - time.Second), expected: cacheStale},
		{name: "expired", entry: testEntry(300 * time.Second), expected: cacheExpired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := testCache(t)
			assert.Equal(t, tt.expected, c.state(tt.entry, ttl))
		})
	}
}

func TestCache_ErrorBackoff(t *testing.T) {
	t.Run("no backoff when nil", func(t *testing.T) {
		c := testCache(t)
		assert.False(t, c.inErrorBackoff(nil))
	})

	t.Run("no backoff when no error_until", func(t *testing.T) {
		c := testCache(t)
		entry := testEntry(0)
		assert.False(t, c.inErrorBackoff(entry))
	})

	t.Run("in backoff when error_until is future", func(t *testing.T) {
		c := testCache(t)
		future := time.Now().Add(5 * time.Minute)
		entry := &cacheEntry{ErrorUntil: &future}
		assert.True(t, c.inErrorBackoff(entry))
	})

	t.Run("not in backoff when error_until is past", func(t *testing.T) {
		c := testCache(t)
		past := time.Now().Add(-5 * time.Minute)
		entry := &cacheEntry{ErrorUntil: &past}
		assert.False(t, c.inErrorBackoff(entry))
	})

	t.Run("writeError preserves existing data", func(t *testing.T) {
		c := testCache(t)
		existing := testEntry(0)
		c.writeError(existing, 5*time.Minute)

		got := c.read()
		require.NotNil(t, got)
		require.NotNil(t, got.Data)
		assert.InDelta(t, 42, got.Data.FiveHour.Utilisation, 0.1)
		require.NotNil(t, got.ErrorUntil)
		assert.True(t, got.ErrorUntil.After(time.Now()))
	})

	t.Run("writeError without existing data", func(t *testing.T) {
		c := testCache(t)
		c.writeError(nil, 5*time.Minute)

		got := c.read()
		require.NotNil(t, got)
		assert.Nil(t, got.Data)
		require.NotNil(t, got.ErrorUntil)
	})
}

func TestCache_Lock(t *testing.T) {
	t.Run("first lock succeeds", func(t *testing.T) {
		c := testCache(t)
		assert.True(t, c.tryLock(30*time.Second))
		c.unlock()
	})

	t.Run("second lock fails while held", func(t *testing.T) {
		c := testCache(t)
		assert.True(t, c.tryLock(30*time.Second))
		assert.False(t, c.tryLock(30*time.Second))
		c.unlock()
	})

	t.Run("lock available after unlock", func(t *testing.T) {
		c := testCache(t)
		assert.True(t, c.tryLock(30*time.Second))
		c.unlock()
		assert.True(t, c.tryLock(30*time.Second))
		c.unlock()
	})

	t.Run("stale lock is cleaned up", func(t *testing.T) {
		c := testCache(t)
		// Create a lock dir manually with old mtime.
		lockDir := c.path + ".lock"
		require.NoError(t, os.Mkdir(lockDir, 0750))
		past := time.Now().Add(-1 * time.Minute)
		require.NoError(t, os.Chtimes(lockDir, past, past))

		// First call cleans up the stale lock but returns false.
		assert.False(t, c.tryLock(30*time.Second))

		// Next call succeeds because the stale lock was removed.
		assert.True(t, c.tryLock(30*time.Second))
		c.unlock()
	})
}
