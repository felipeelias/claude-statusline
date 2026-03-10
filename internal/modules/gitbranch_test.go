package modules_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
	"github.com/felipeelias/claude-statusline/internal/modules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitBranchModule_Name(t *testing.T) {
	m := modules.GitBranchModule{}
	assert.Equal(t, "git_branch", m.Name())
}

func initGitRepo(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	cmd := exec.CommandContext(t.Context(), "git", "init", dir)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null")
	require.NoError(t, cmd.Run())

	cmd = exec.CommandContext(t.Context(), "git", "-C", dir, "config", "user.email", "test@test.com")
	require.NoError(t, cmd.Run())

	cmd = exec.CommandContext(t.Context(), "git", "-C", dir, "config", "user.name", "Test")
	require.NoError(t, cmd.Run())

	cmd = exec.CommandContext(t.Context(), "git", "-C", dir, "commit", "--allow-empty", "-m", "init")
	require.NoError(t, cmd.Run())

	return dir
}

func TestGitBranchModule_ReturnsBranchName(t *testing.T) {
	cfg := config.Default()
	dir := initGitRepo(t)

	cmd := exec.CommandContext(t.Context(), "git", "-C", dir, "checkout", "-b", "test-branch")
	require.NoError(t, cmd.Run())

	data := input.Data{Cwd: dir}
	result, err := modules.GitBranchModule{}.Render(data, cfg)
	require.NoError(t, err)
	assert.Contains(t, result, "test-branch")
}

func TestGitBranchModule_NonGitDir(t *testing.T) {
	cfg := config.Default()
	dir := t.TempDir()

	data := input.Data{Cwd: dir}
	result, err := modules.GitBranchModule{}.Render(data, cfg)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestGitBranchModule_NonexistentDir(t *testing.T) {
	cfg := config.Default()

	data := input.Data{Cwd: filepath.Join(os.TempDir(), "nonexistent-dir-12345")}
	result, err := modules.GitBranchModule{}.Render(data, cfg)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestGitBranchModule_WorktreeIcon(t *testing.T) {
	cfg := config.Default()
	dir := initGitRepo(t)

	data := input.Data{
		Cwd:      dir,
		Worktree: &input.Worktree{Name: "feature-branch"},
	}
	result, err := modules.GitBranchModule{}.Render(data, cfg)
	require.NoError(t, err)
	assert.Contains(t, result, string('\uf0e8'))
}

func TestGitBranchModule_NoWorktreeIcon(t *testing.T) {
	cfg := config.Default()
	dir := initGitRepo(t)

	data := input.Data{Cwd: dir}
	result, err := modules.GitBranchModule{}.Render(data, cfg)
	require.NoError(t, err)
	assert.NotContains(t, result, string('\uf0e8'))
}
