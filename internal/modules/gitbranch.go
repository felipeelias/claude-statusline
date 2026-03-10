package modules

import (
	"os/exec"
	"strings"

	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
)

// GitBranchModule renders the current git branch name with optional worktree indicator.
type GitBranchModule struct{}

func (GitBranchModule) Name() string { return "git_branch" }

func (GitBranchModule) Render(data input.Data, cfg config.Config) (string, error) {
	branch := gitBranch(data.Cwd)
	if branch == "" {
		return "", nil
	}

	inWorktree := data.Worktree != nil && data.Worktree.Name != ""

	templateData := struct {
		Branch     string
		InWorktree bool
	}{
		Branch:     branch,
		InWorktree: inWorktree,
	}

	result, err := renderTemplate("git_branch", cfg.GitBranch.Format, templateData)
	if err != nil {
		return "", err
	}

	return wrapStyle(result, cfg.GitBranch.Style, cfg), nil
}

// gitBranch runs git rev-parse to get the current branch name.
// Returns empty string if the directory is not a git repo or git is not installed.
func gitBranch(cwd string) string {
	//nolint:noctx // no context available in module interface
	cmd := exec.Command("git", "-C", cwd, "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(out))
}
