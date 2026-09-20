package testutil

import (
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
