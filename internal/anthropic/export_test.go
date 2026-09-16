package anthropic

// Test hooks. Keeping them in an _test.go file means the seams exist only
// while testing: the shipped binary always talks to Anthropic and always
// refreshes through a detached process.

// SetBaseURL points the client at a test server for the duration of a test.
func SetBaseURL(t interface{ Cleanup(f func()) }, url string) {
	previous := baseURL
	baseURL = url

	t.Cleanup(func() { baseURL = previous })
}

// SuppressRefresh replaces the background refresh with a counter, so a test
// can assert that a stale cache triggers one without spawning a process.
func SuppressRefresh(t interface{ Cleanup(f func()) }) *int {
	calls := 0
	previous := refresh
	refresh = func() { calls++ }

	t.Cleanup(func() { refresh = previous })

	return &calls
}

// SuppressKeychain stops credential resolution reaching the real Keychain.
// Without it a test that sets HOME to a temporary directory still resolves the
// default Keychain item - the developer's own live credentials - and a test
// asserting "no credentials here" passes or fails depending on whose machine
// it runs on.
func SuppressKeychain(t interface{ Cleanup(f func()) }) {
	previous := keychainLookup
	keychainLookup = func(string) []byte { return nil }

	t.Cleanup(func() { keychainLookup = previous })
}
