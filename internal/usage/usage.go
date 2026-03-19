package usage

import (
	"errors"
	"time"
)

const (
	tokenExpiredBackoff = 10 * time.Minute
	networkErrorBackoff = 60 * time.Second
)

// Fetcher retrieves usage data, handling caching and credentials transparently.
type Fetcher interface {
	GetUsage(cacheTTL time.Duration) (*UsageData, error)
}

// DefaultFetcher fetches usage data via the OAuth API with file-based caching.
type DefaultFetcher struct{}

const (
	minCacheTTL = 30 * time.Second
	fetchLockMaxAge = 30 * time.Second // lock auto-expires after this (covers crashed processes)
)

// GetUsage retrieves usage data from the cache or API.
// Returns (nil, nil) on any failure so the module can degrade gracefully.
func (DefaultFetcher) GetUsage(cacheTTL time.Duration) (*UsageData, error) {
	if cacheTTL < minCacheTTL {
		cacheTTL = minCacheTTL
	}

	cachePath := defaultCachePath()
	if cachePath == "" {
		return nil, nil
	}

	c := &cache{path: cachePath}
	entry := c.read()

	if c.state(entry, cacheTTL) == cacheFresh {
		return entry.Data, nil
	}

	if c.inErrorBackoff(entry) {
		if entry != nil && entry.Data != nil {
			return entry.Data, nil
		}

		return nil, nil
	}

	// Stale or expired: try to acquire the fetch lock.
	// Only one process fetches at a time; others serve stale data or return empty.
	if !c.tryLock(fetchLockMaxAge) {
		if entry != nil && entry.Data != nil {
			return entry.Data, nil
		}

		return nil, nil
	}
	defer c.unlock()

	// Re-read cache after acquiring lock - another process may have just refreshed it.
	entry = c.read()
	if c.state(entry, cacheTTL) == cacheFresh {
		return entry.Data, nil
	}

	data, err := fetchFromAPI(c, entry)
	if err != nil {
		if entry != nil && entry.Data != nil {
			return entry.Data, nil
		}

		return nil, nil
	}

	return data, nil
}

// fetchFromAPI retrieves credentials, calls the API, and updates the cache.
func fetchFromAPI(c *cache, existing *cacheEntry) (*UsageData, error) {
	credSrc, err := NewCredentialSource()
	if err != nil {
		return nil, err
	}

	token, err := credSrc.AccessToken()
	if err != nil {
		return nil, err
	}

	cl := newClient()

	data, err := cl.fetch(token)
	if err != nil {
		backoff := errorBackoff(err)
		c.writeError(existing, backoff)

		return nil, err
	}

	_ = c.write(&cacheEntry{
		Data:      data,
		FetchedAt: time.Now(),
	})

	return data, nil
}

// errorBackoff returns the appropriate backoff duration for a given error.
func errorBackoff(err error) time.Duration {
	if errors.Is(err, errTokenExpired) {
		return tokenExpiredBackoff
	}

	var rle *rateLimitError
	if errors.As(err, &rle) && rle.retryAfter > 0 {
		return rle.retryAfter
	}

	return networkErrorBackoff
}

// MockFetcher returns fixed usage data for testing and preview modes.
type MockFetcher struct {
	Data *UsageData
}

// GetUsage returns the mock data.
func (m MockFetcher) GetUsage(_ time.Duration) (*UsageData, error) {
	return m.Data, nil
}
