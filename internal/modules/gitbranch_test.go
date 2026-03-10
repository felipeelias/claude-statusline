package modules

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitBranchModule_Name(t *testing.T) {
	m := GitBranchModule{}
	assert.Equal(t, "git_branch", m.Name())
}

func TestGitBranchModule_Render(t *testing.T) {
	cfg := config.Default()

	t.Run("returns branch name in a git repo", func(t *testing.T) {
		dir := t.TempDir()

		// Initialize a git repo and make a commit so HEAD exists
		cmd := exec.Command("git", "init", dir)
		cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null")
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "-C", dir, "config", "user.email", "test@test.com")
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "-C", dir, "config", "user.name", "Test")
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "-C", dir, "commit", "--allow-empty", "-m", "init")
		require.NoError(t, cmd.Run())

		// Create a known branch name so the test is deterministic
		cmd = exec.Command("git", "-C", dir, "checkout", "-b", "test-branch")
		require.NoError(t, cmd.Run())

		data := input.Data{Cwd: dir}
		result, err := GitBranchModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Contains(t, result, "test-branch")
	})

	t.Run("non-git directory returns empty string", func(t *testing.T) {
		dir := t.TempDir()

		data := input.Data{Cwd: dir}
		result, err := GitBranchModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("nonexistent directory returns empty string", func(t *testing.T) {
		data := input.Data{Cwd: filepath.Join(os.TempDir(), "nonexistent-dir-12345")}
		result, err := GitBranchModule{}.Render(data, cfg)
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})
}
