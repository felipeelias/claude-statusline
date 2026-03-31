package modules_test

import (
	"os"
	"os/exec"
	"testing"

	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
	"github.com/felipeelias/claude-statusline/internal/modules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrapHyperlink(t *testing.T) {
	t.Run("wraps text with OSC 8 sequence", func(t *testing.T) {
		result := modules.WrapHyperlink("https://github.com/owner/repo", "main")
		expected := "\033]8;;https://github.com/owner/repo\033\\main\033]8;;\033\\"
		assert.Equal(t, expected, result)
	})

	t.Run("empty URL returns text unchanged", func(t *testing.T) {
		result := modules.WrapHyperlink("", "main")
		assert.Equal(t, "main", result)
	})
}

func TestGitRemoteToHTTPS(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "GitHub SSH with .git",
			input:    "git@github.com:owner/repo.git",
			expected: "https://github.com/owner/repo",
		},
		{
			name:     "GitHub SSH without .git",
			input:    "git@github.com:owner/repo",
			expected: "https://github.com/owner/repo",
		},
		{
			name:     "GitLab SSH",
			input:    "git@gitlab.com:group/project.git",
			expected: "https://gitlab.com/group/project",
		},
		{
			name:     "Bitbucket SSH",
			input:    "git@bitbucket.org:team/repo.git",
			expected: "https://bitbucket.org/team/repo",
		},
		{
			name:     "GitHub HTTPS with .git",
			input:    "https://github.com/owner/repo.git",
			expected: "https://github.com/owner/repo",
		},
		{
			name:     "GitHub HTTPS without .git",
			input:    "https://github.com/owner/repo",
			expected: "https://github.com/owner/repo",
		},
		{
			name:     "GitLab nested subgroup SSH",
			input:    "git@gitlab.com:group/subgroup/project.git",
			expected: "https://gitlab.com/group/subgroup/project",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace only",
			input:    "  \n",
			expected: "",
		},
		{
			name:     "invalid URL",
			input:    "not-a-url",
			expected: "",
		},
		{
			name:     "SSH with hyphenated hostname",
			input:    "git@my-gitlab.company.com:org/repo.git",
			expected: "https://my-gitlab.company.com/org/repo",
		},
		{
			name:     "SSH with hyphenated username",
			input:    "my-user@github.com:owner/repo.git",
			expected: "https://github.com/owner/repo",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := modules.GitRemoteToHTTPS(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestBranchURL(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		branch   string
		expected string
	}{
		{
			name:     "GitHub default",
			baseURL:  "https://github.com/owner/repo",
			branch:   "main",
			expected: "https://github.com/owner/repo/tree/main",
		},
		{
			name:     "GitHub trailing slash",
			baseURL:  "https://github.com/owner/repo/",
			branch:   "main",
			expected: "https://github.com/owner/repo/tree/main",
		},
		{
			name:     "GitLab uses /-/tree/",
			baseURL:  "https://gitlab.com/group/project",
			branch:   "main",
			expected: "https://gitlab.com/group/project/-/tree/main",
		},
		{
			name:     "self-hosted GitLab",
			baseURL:  "https://my-gitlab.company.com/org/repo",
			branch:   "develop",
			expected: "https://my-gitlab.company.com/org/repo/-/tree/develop",
		},
		{
			name:     "Bitbucket uses /src/",
			baseURL:  "https://bitbucket.org/team/repo",
			branch:   "main",
			expected: "https://bitbucket.org/team/repo/src/main",
		},
		{
			name:     "branch with hash is encoded",
			baseURL:  "https://github.com/owner/repo",
			branch:   "fix/#123",
			expected: "https://github.com/owner/repo/tree/fix/%23123",
		},
		{
			name:     "branch with spaces is encoded",
			baseURL:  "https://github.com/owner/repo",
			branch:   "feature/my branch",
			expected: "https://github.com/owner/repo/tree/feature/my%20branch",
		},
		{
			name:     "simple branch with slash",
			baseURL:  "https://github.com/owner/repo",
			branch:   "feature/my-branch",
			expected: "https://github.com/owner/repo/tree/feature/my-branch",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := modules.BranchURL(tc.baseURL, tc.branch)
			assert.Equal(t, tc.expected, result)
		})
	}
}

const branchOnlyFmt = "{{.Branch}}"

func TestGitBranchModule_HyperlinkEnabled(t *testing.T) {
	cfg := config.Default()
	cfg.GitBranch.Hyperlink = true
	cfg.GitBranch.HyperlinkBaseURL = "https://github.com/owner/repo"
	// Use a simple format to make assertions easier.
	cfg.GitBranch.Format = branchOnlyFmt
	cfg.GitBranch.Style = ""

	dir := initGitRepo(t)

	cmd := exec.CommandContext(t.Context(), "git", "-C", dir, "checkout", "-b", "my-feature")
	require.NoError(t, cmd.Run())

	data := input.Data{Cwd: dir}
	result, err := modules.GitBranchModule{}.Render(data, cfg)
	require.NoError(t, err)
	assert.Contains(t, result, "\033]8;;https://github.com/owner/repo/tree/my-feature\033\\")
	assert.Contains(t, result, "my-feature")
	assert.Contains(t, result, "\033]8;;\033\\")
}

func TestGitBranchModule_HyperlinkDisabled(t *testing.T) {
	cfg := config.Default()
	cfg.GitBranch.Hyperlink = false
	cfg.GitBranch.Format = branchOnlyFmt
	cfg.GitBranch.Style = ""

	dir := initGitRepo(t)

	data := input.Data{Cwd: dir}
	result, err := modules.GitBranchModule{}.Render(data, cfg)
	require.NoError(t, err)
	assert.NotContains(t, result, "\033]8;;")
}

func TestGitBranchModule_HyperlinkAutoDetect(t *testing.T) {
	origin := initGitRepo(t)

	dir := t.TempDir()
	cmd := exec.CommandContext(t.Context(), "git", "clone", origin, dir)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null")
	require.NoError(t, cmd.Run())

	cmd = exec.CommandContext(t.Context(), "git", "-C", dir, "config", "user.email", "test@test.com")
	require.NoError(t, cmd.Run())
	cmd = exec.CommandContext(t.Context(), "git", "-C", dir, "config", "user.name", "Test")
	require.NoError(t, cmd.Run())

	cfg := config.Default()
	cfg.GitBranch.Hyperlink = true
	// No HyperlinkBaseURL set — auto-detect from remote.
	cfg.GitBranch.Format = branchOnlyFmt
	cfg.GitBranch.Style = ""

	data := input.Data{Cwd: dir}
	result, err := modules.GitBranchModule{}.Render(data, cfg)
	require.NoError(t, err)
	// The remote is a local path, so gitRemoteToHTTPS will return "" and no link is added.
	// This tests graceful degradation.
	assert.NotContains(t, result, "\033]8;;")
	assert.NotEmpty(t, result) // branch name is present
}

func TestGitBranchModule_HyperlinkNoRemoteGraceful(t *testing.T) {
	cfg := config.Default()
	cfg.GitBranch.Hyperlink = true
	// No base URL, and the repo has no remote.
	cfg.GitBranch.Format = branchOnlyFmt
	cfg.GitBranch.Style = ""

	dir := initGitRepo(t)

	data := input.Data{Cwd: dir}
	result, err := modules.GitBranchModule{}.Render(data, cfg)
	require.NoError(t, err)
	// No remote available, no base URL — should render without hyperlink.
	assert.NotContains(t, result, "\033]8;;")
	assert.NotEmpty(t, result) // branch name is present
}

func TestDirectoryModule_HyperlinkEnabled(t *testing.T) {
	cfg := config.Default()
	cfg.Directory.Hyperlink = true
	cfg.Directory.Style = ""

	data := input.Data{Cwd: "/home/user/projects/myapp"}

	result, err := modules.NewDirectoryModuleWithHome("/home/user").Render(data, cfg)
	require.NoError(t, err)
	assert.Contains(t, result, "\033]8;;file:///home/user/projects/myapp\033\\")
	assert.Contains(t, result, "\033]8;;\033\\")
}

func TestDirectoryModule_HyperlinkEncodesSpaces(t *testing.T) {
	cfg := config.Default()
	cfg.Directory.Hyperlink = true
	cfg.Directory.Style = ""

	data := input.Data{Cwd: "/home/user/my projects/app"}

	result, err := modules.NewDirectoryModuleWithHome("/home/user").Render(data, cfg)
	require.NoError(t, err)
	assert.Contains(t, result, "\033]8;;file:///home/user/my%20projects/app\033\\")
}

func TestDirectoryModule_HyperlinkDisabled(t *testing.T) {
	cfg := config.Default()
	cfg.Directory.Hyperlink = false
	cfg.Directory.Style = ""

	data := input.Data{Cwd: "/home/user/projects"}

	result, err := modules.NewDirectoryModuleWithHome("/home/user").Render(data, cfg)
	require.NoError(t, err)
	assert.NotContains(t, result, "\033]8;;")
}

func TestDirectoryModule_HyperlinkCustomTemplate(t *testing.T) {
	cfg := config.Default()
	cfg.Directory.Hyperlink = true
	cfg.Directory.HyperlinkURLTemplate = "vscode://file{{.AbsPath}}"
	cfg.Directory.Style = ""

	data := input.Data{Cwd: "/home/user/projects/myapp"}

	result, err := modules.NewDirectoryModuleWithHome("/home/user").Render(data, cfg)
	require.NoError(t, err)
	assert.Contains(t, result, "\033]8;;vscode://file/home/user/projects/myapp\033\\")
}

func TestDirectoryModule_HyperlinkRawPathTemplate(t *testing.T) {
	cfg := config.Default()
	cfg.Directory.Hyperlink = true
	cfg.Directory.HyperlinkURLTemplate = "vscode://file{{.AbsPath}}"
	cfg.Directory.Style = ""

	data := input.Data{Cwd: "/home/user/my projects/app"}

	result, err := modules.NewDirectoryModuleWithHome("/home/user").Render(data, cfg)
	require.NoError(t, err)
	// AbsPath is raw — spaces are NOT encoded
	assert.Contains(t, result, "\033]8;;vscode://file/home/user/my projects/app\033\\")
}

func TestDirectoryModule_HyperlinkEmptyTemplate(t *testing.T) {
	cfg := config.Default()
	cfg.Directory.Hyperlink = true
	cfg.Directory.HyperlinkURLTemplate = ""
	cfg.Directory.Style = ""

	data := input.Data{Cwd: "/home/user/projects"}

	result, err := modules.NewDirectoryModuleWithHome("/home/user").Render(data, cfg)
	require.NoError(t, err)
	// Empty template means no hyperlink.
	assert.NotContains(t, result, "\033]8;;")
}
