package modules

import (
	"net/url"
	"os"
	"strings"

	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
)

// DirectoryModule renders the current working directory with tilde substitution and truncation.
type DirectoryModule struct {
	homeDir string
}

// NewDirectoryModule creates a DirectoryModule that uses the real home directory.
func NewDirectoryModule() DirectoryModule {
	home, _ := os.UserHomeDir()

	return DirectoryModule{homeDir: home}
}

// NewDirectoryModuleWithHome creates a DirectoryModule with a custom home directory for testing.
func NewDirectoryModuleWithHome(home string) DirectoryModule {
	return DirectoryModule{homeDir: home}
}

func (DirectoryModule) Name() string { return "directory" }

func (m DirectoryModule) Render(data input.Data, cfg config.Config) (string, error) {
	cwd := data.Cwd
	if cwd == "" {
		return "", nil
	}

	home := m.homeDir
	if home == "" {
		home, _ = os.UserHomeDir()
	}

	// Tilde substitution.
	dir := cwd
	if home != "" {
		if dir == home {
			dir = "~"
		} else if strings.HasPrefix(dir, home+"/") {
			dir = "~" + dir[len(home):]
		}
	}

	dir = truncatePath(dir, cfg.Directory.TruncationLength)

	templateData := struct{ Dir string }{Dir: dir}

	result, err := renderTemplate("directory", cfg.Directory.Format, templateData)
	if err != nil {
		return "", err
	}

	if cfg.Directory.Hyperlink {
		linkURL := resolveDirectoryHyperlink(cfg.Directory.HyperlinkURLTemplate, cwd)
		result = WrapHyperlink(linkURL, result)
	}

	return wrapStyle(result, cfg.Directory.Style), nil
}

// truncatePath keeps the last maxSegments path segments fully and abbreviates earlier ones
// to their first character. The leading "/" or "~/" prefix is preserved.
func truncatePath(path string, maxSegments int) string {
	if maxSegments <= 0 {
		return path
	}

	prefix, segmentStr := splitPathPrefix(path)
	if segmentStr == "" {
		return prefix
	}

	segments := strings.Split(segmentStr, "/")

	if len(segments) <= maxSegments {
		return path
	}

	cutoff := len(segments) - maxSegments
	for i := range cutoff {
		if len(segments[i]) > 0 {
			runes := []rune(segments[i])
			segments[i] = string(runes[0])
		}
	}

	return prefix + strings.Join(segments, "/")
}

func splitPathPrefix(path string) (string, string) {
	if strings.HasPrefix(path, "~/") {
		return "~/", path[2:]
	}

	if path == "~" {
		return "~", ""
	}

	if strings.HasPrefix(path, "/") {
		return "/", path[1:]
	}

	return "", path
}

// resolveDirectoryHyperlink executes the URL template with the absolute path.
// Returns empty string if the template is empty or fails to execute.
func resolveDirectoryHyperlink(urlTemplate, absPath string) string {
	if urlTemplate == "" {
		return ""
	}

	data := struct {
		AbsPath        string
		AbsPathEncoded string
	}{
		AbsPath:        absPath,
		AbsPathEncoded: (&url.URL{Path: absPath}).EscapedPath(),
	}

	result, err := renderTemplate("hyperlink_url", urlTemplate, data)
	if err != nil {
		return ""
	}

	return result
}
