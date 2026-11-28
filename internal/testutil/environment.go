package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

func IsolateHome(t *testing.T) {
	t.Helper()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(
		"CLAUDE_STATUSLINE_CONFIG",
		filepath.Join(home, ".config", "claude-statusline", "config.toml"),
	)
}

func UnsetEnv(t *testing.T, key string) {
	t.Helper()

	t.Setenv(key, "")

	err := os.Unsetenv(key)
	if err != nil {
		t.Fatal(err)
	}
}
