package modules

import (
	"os/exec"
	"strings"

	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
)

// GitBranchModule renders the current git branch name with optional status indicators.
type GitBranchModule struct{}

func (GitBranchModule) Name() string { return "git_branch" }

func (GitBranchModule) Render(data input.Data, cfg config.Config) (string, error) {
	inWorktree := data.Worktree != nil && data.Worktree.Name != ""

	var templateData gitBranchTemplateData

	if cfg.GitBranch.Mode == "simple" {
		branch := gitBranchSimple(data.Cwd)
		if branch == "" {
			return "", nil
		}
		templateData = gitBranchTemplateData{
			Branch:     branch,
			InWorktree: inWorktree,
		}
	} else {
		status := gitStatusDetailed(data.Cwd)
		if status.Branch == "" {
			return "", nil
		}
		dirty := status.Staged > 0 || status.Modified > 0 || status.Untracked > 0 || status.Conflicts > 0
		templateData = gitBranchTemplateData{
			Branch:     status.Branch,
			InWorktree: inWorktree,
			Staged:     status.Staged,
			Modified:   status.Modified,
			Untracked:  status.Untracked,
			Ahead:      status.Ahead,
			Behind:     status.Behind,
			Conflicts:  status.Conflicts,
			IsDirty:    dirty,
			IsClean:    !dirty,
		}
	}

	result, err := renderTemplate("git_branch", cfg.GitBranch.Format, templateData)
	if err != nil {
		return "", err
	}

	result = gitBranchHyperlink(result, templateData.Branch, data.Cwd, cfg.GitBranch)

	return wrapStyle(result, cfg.GitBranch.Style), nil
}

// gitBranchHyperlink wraps text in an OSC 8 hyperlink to the branch on the remote.
// Returns text unchanged if hyperlink is disabled or no base URL can be determined.
func gitBranchHyperlink(text, branch, cwd string, cfg config.GitBranchConfig) string {
	if !cfg.Hyperlink {
		return text
	}

	baseURL := cfg.HyperlinkBaseURL
	if baseURL == "" {
		baseURL = GitRemoteToHTTPS(gitRemoteURL(cwd))
	}

	if baseURL == "" {
		return text
	}

	return WrapHyperlink(BranchURL(baseURL, branch), text)
}

type gitBranchTemplateData struct {
	Branch     string
	InWorktree bool
	Staged     int
	Modified   int
	Untracked  int
	Ahead      int
	Behind     int
	Conflicts  int
	IsDirty    bool
	IsClean    bool
}

// gitBranchSimple runs git rev-parse to get the current branch name.
// Returns empty string if the directory is not a git repo or git is not installed.
func gitBranchSimple(cwd string) string {
	//nolint:noctx // no context available in module interface
	cmd := exec.Command("git", "-C", cwd, "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(out))
}

// gitStatusDetailed runs git status --porcelain=v2 --branch and parses the output.
// Returns zero-value GitStatus if the directory is not a git repo or git is not installed.
func gitStatusDetailed(cwd string) GitStatus {
	//nolint:noctx // no context available in module interface
	cmd := exec.Command("git", "-C", cwd, "status", "--porcelain=v2", "--branch")
	out, err := cmd.Output()
	if err != nil {
		return GitStatus{}
	}

	return ParsePorcelainV2(string(out))
}
