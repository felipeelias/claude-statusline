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

func TestGitBranchModule_DetailedMode_CleanRepo(t *testing.T) {
	cfg := config.Default()
	dir := initGitRepo(t)

	data := input.Data{Cwd: dir}
	result, err := modules.GitBranchModule{}.Render(data, cfg)
	require.NoError(t, err)
	assert.NotContains(t, result, "*")
	assert.NotContains(t, result, "\u2191")
	assert.NotContains(t, result, "\u2193")
}

func TestGitBranchModule_DetailedMode_DirtyIndicator(t *testing.T) {
	cfg := config.Default()
	dir := initGitRepo(t)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "dirty.txt"), []byte("dirty"), 0644))

	data := input.Data{Cwd: dir}
	result, err := modules.GitBranchModule{}.Render(data, cfg)
	require.NoError(t, err)
	assert.Contains(t, result, "*")
}

func TestGitBranchModule_DetailedMode_StagedIsDirty(t *testing.T) {
	cfg := config.Default()
	dir := initGitRepo(t)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "staged.txt"), []byte("new"), 0644))
	cmd := exec.CommandContext(t.Context(), "git", "-C", dir, "add", "staged.txt")
	require.NoError(t, cmd.Run())

	data := input.Data{Cwd: dir}
	result, err := modules.GitBranchModule{}.Render(data, cfg)
	require.NoError(t, err)
	assert.Contains(t, result, "*")
}

func TestGitBranchModule_DetailedMode_AheadBehind(t *testing.T) {
	cfg := config.Default()
	origin := initGitRepo(t)

	dir := t.TempDir()
	cmd := exec.CommandContext(t.Context(), "git", "clone", origin, dir)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null")
	require.NoError(t, cmd.Run())

	cmd = exec.CommandContext(t.Context(), "git", "-C", dir, "config", "user.email", "test@test.com")
	require.NoError(t, cmd.Run())
	cmd = exec.CommandContext(t.Context(), "git", "-C", dir, "config", "user.name", "Test")
	require.NoError(t, cmd.Run())

	cmd = exec.CommandContext(t.Context(), "git", "-C", dir, "commit", "--allow-empty", "-m", "ahead")
	require.NoError(t, cmd.Run())

	data := input.Data{Cwd: dir}
	result, err := modules.GitBranchModule{}.Render(data, cfg)
	require.NoError(t, err)
	assert.Contains(t, result, "\u21911") // ↑1
}

func TestGitBranchModule_SimpleMode_BranchOnly(t *testing.T) {
	cfg := config.Default()
	cfg.GitBranch.Mode = "simple"
	dir := initGitRepo(t)

	// Create a dirty file — should NOT show * in simple mode
	require.NoError(t, os.WriteFile(filepath.Join(dir, "dirty.txt"), []byte("dirty"), 0644))

	cmd := exec.CommandContext(t.Context(), "git", "-C", dir, "checkout", "-b", "my-branch")
	require.NoError(t, cmd.Run())

	data := input.Data{Cwd: dir}
	result, err := modules.GitBranchModule{}.Render(data, cfg)
	require.NoError(t, err)
	assert.Contains(t, result, "my-branch")
	assert.NotContains(t, result, "*")
}

func TestGitBranchModule_SimpleMode_NonGitDir(t *testing.T) {
	cfg := config.Default()
	cfg.GitBranch.Mode = "simple"
	dir := t.TempDir()

	data := input.Data{Cwd: dir}
	result, err := modules.GitBranchModule{}.Render(data, cfg)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestGitBranchModule_DetailedMode_NonGitDir(t *testing.T) {
	cfg := config.Default()
	dir := t.TempDir()

	data := input.Data{Cwd: dir}
	result, err := modules.GitBranchModule{}.Render(data, cfg)
	require.NoError(t, err)
	assert.Empty(t, result)
}
