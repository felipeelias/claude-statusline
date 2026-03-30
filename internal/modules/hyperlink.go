package modules

import (
	"net/url"
	"os/exec"
	"regexp"
	"strings"
)

// WrapHyperlink wraps text in an OSC 8 terminal hyperlink sequence.
// If linkURL is empty, the text is returned unmodified.
func WrapHyperlink(linkURL, text string) string {
	if linkURL == "" {
		return text
	}

	return "\033]8;;" + linkURL + "\033\\" + text + "\033]8;;\033\\"
}

// sshURLPattern matches SSH-style git remote URLs like git@github.com:owner/repo.git.
var sshURLPattern = regexp.MustCompile(`^[\w.-]+@([\w.-]+):([\w./-]+?)(?:\.git)?$`)

// GitRemoteToHTTPS converts a git remote URL (SSH or HTTPS) to an HTTPS base URL.
// Returns empty string if the URL cannot be parsed.
func GitRemoteToHTTPS(remoteURL string) string {
	remoteURL = strings.TrimSpace(remoteURL)
	if remoteURL == "" {
		return ""
	}

	// Handle SSH URLs: git@github.com:owner/repo.git
	if m := sshURLPattern.FindStringSubmatch(remoteURL); m != nil {
		return "https://" + m[1] + "/" + m[2]
	}

	// Handle HTTPS URLs: https://github.com/owner/repo.git
	parsed, err := url.Parse(remoteURL)
	if err != nil || parsed.Host == "" {
		return ""
	}

	path := strings.TrimSuffix(parsed.Path, ".git")

	return "https://" + parsed.Host + path
}

// gitRemoteURL runs git remote get-url origin in the given directory.
// Returns empty string if the command fails or git is not available.
func gitRemoteURL(cwd string) string {
	//nolint:noctx // no context available in module interface
	cmd := exec.Command("git", "-C", cwd, "remote", "get-url", "origin")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(out))
}
