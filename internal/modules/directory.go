package modules

import (
	"bytes"
	"net/url"
	"os"
	"strings"
	"text/template"

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

	sep := pathSeparator(cwd)

	// Tilde substitution.
	dir := cwd
	if home != "" {
		if dir == home {
			dir = "~"
		} else if strings.HasPrefix(dir, home+sep) {
			dir = "~" + dir[len(home):]
		}
	}

	dir = truncatePath(dir, cfg.Directory.TruncationLength, sep)

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

// pathSeparator returns the separator used by path: a backslash for an
// anchored Windows path (drive prefix or leading backslash), otherwise a
// forward slash.
func pathSeparator(path string) string {
	if strings.HasPrefix(path, "\\") {
		return "\\"
	}

	if len(path) >= 3 && isDriveLetter(path[0]) && path[1] == ':' && path[2] == '\\' {
		return "\\"
	}

	return "/"
}

// truncatePath keeps the last maxSegments path segments fully and abbreviates earlier ones
// to their first character. The leading root (Unix "/", Windows drive "C:\", or UNC
// "\\server\share\") or "~"+sep prefix is preserved.
func truncatePath(path string, maxSegments int, sep string) string {
	if maxSegments <= 0 {
		return path
	}

	prefix, segmentStr := splitPathPrefix(path, sep)
	if segmentStr == "" {
		return prefix
	}

	segments := strings.Split(segmentStr, sep)

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

	return prefix + strings.Join(segments, sep)
}

func splitPathPrefix(path, sep string) (string, string) {
	tildePrefix := "~" + sep
	if strings.HasPrefix(path, tildePrefix) {
		return tildePrefix, path[len(tildePrefix):]
	}

	if path == "~" {
		return "~", ""
	}

	// Windows drive prefix, e.g. "C:\" or "C:/".
	if len(path) >= 3 && isDriveLetter(path[0]) && path[1] == ':' && string(path[2]) == sep {
		return path[:3], path[3:]
	}

	if prefix, rest, ok := splitUNCPrefix(path, sep); ok {
		return prefix, rest
	}

	if strings.HasPrefix(path, sep) {
		return sep, path[len(sep):]
	}

	return "", path
}

// uncPathParts is the number of path components in a UNC prefix: an empty
// leading component, the server name, and the share name.
const uncPathParts = 3

// splitUNCPrefix splits a Windows UNC prefix, e.g. "\\server\share\" or
// "\\server\share", from the rest of path. ok is false if path does not
// start with a UNC prefix.
func splitUNCPrefix(path, sep string) (string, string, bool) {
	if sep != "\\" || !strings.HasPrefix(path, sep+sep) {
		return "", "", false
	}

	parts := strings.SplitN(path[2:], sep, uncPathParts)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}

	prefix := sep + sep + parts[0] + sep + parts[1]
	if len(parts) == uncPathParts {
		return prefix + sep, parts[2], true
	}

	return prefix, "", true
}

func isDriveLetter(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}

// resolveDirectoryHyperlink executes the URL template with the absolute path.
// Returns empty string if the template is empty or fails to execute.
func resolveDirectoryHyperlink(urlTemplate, absPath string) string {
	if urlTemplate == "" {
		return ""
	}

	tmpl, err := template.New("hyperlink_url").Parse(urlTemplate)
	if err != nil {
		return ""
	}

	data := struct {
		AbsPath        string
		AbsPathEncoded string
	}{
		AbsPath:        absPath,
		AbsPathEncoded: (&url.URL{Path: absPath}).EscapedPath(),
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return ""
	}

	return buf.String()
}
