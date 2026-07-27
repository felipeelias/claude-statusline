package modules

import (
	"regexp"
	"strings"

	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
)

var modelIDPattern = regexp.MustCompile(`^claude-(opus|sonnet|haiku)-(\d+)-(\d+)(?:-\d+)?$`)

// parenSuffix splits a display name such as "Opus 5 (1M context)" into its base
// name and the text inside the trailing parentheses.
var parenSuffix = regexp.MustCompile(`^(.*?)\s*\(([^)]*)\)\s*$`)

// contextSuffix matches a context-window suffix such as "1M context" so it can
// be abbreviated to "1m" and keep the statusline short.
var contextSuffix = regexp.MustCompile(`(?i)^(\d+(?:\.\d+)?)\s*([mk])\s+context$`)

// ShortName extracts a compact name from a model ID (e.g. "Sonnet 4.6").
// Falls back to displayName if the ID doesn't match the expected pattern.
func ShortName(id, displayName string) string {
	match := modelIDPattern.FindStringSubmatch(id)
	if match == nil {
		return displayName
	}

	family := strings.ToUpper(match[1][:1]) + match[1][1:]

	return family + " " + match[2] + "." + match[3]
}

// ModelModule renders the AI model name.
type ModelModule struct{}

func (ModelModule) Name() string { return "model" }

// modelTemplateData is the data available to the model module's format template.
// DisplayName is the raw value from Claude Code; Name, Context and Effort are
// its parts, and Details joins the parts that vary per session.
type modelTemplateData struct {
	ID          string
	DisplayName string
	Short       string
	Name        string
	Context     string
	Effort      string
	Details     string
}

func (ModelModule) Render(data input.Data, cfg config.Config) (string, error) {
	displayName := data.Model.DisplayName
	if displayName == "" && data.Model.ID == "" {
		return "", nil
	}

	name, contextWindow := splitDisplayName(displayName)
	if name == "" {
		name = ShortName(data.Model.ID, data.Model.ID)
	}

	templateData := modelTemplateData{
		ID:          data.Model.ID,
		DisplayName: displayName,
		Short:       ShortName(data.Model.ID, displayName),
		Name:        name,
		Context:     contextWindow,
		Effort:      data.Effort.Level,
		Details:     joinDetails(contextWindow, data.Effort.Level),
	}

	result, err := renderTemplate("model", cfg.Model.Format, templateData)
	if err != nil {
		return "", err
	}

	return wrapStyle(result, cfg.Model.Style), nil
}

// splitDisplayName separates a recognised context-window suffix from a model
// display name and abbreviates it: "Opus 5 (1M context)" becomes "Opus 5", "1m".
// Any other name is returned unchanged with an empty window, so an unrecognised
// suffix stays part of the name rather than being mistaken for a window size.
func splitDisplayName(displayName string) (string, string) {
	match := parenSuffix.FindStringSubmatch(displayName)
	if match == nil {
		return displayName, ""
	}

	window := contextSuffix.FindStringSubmatch(match[2])
	if window == nil {
		return displayName, ""
	}

	return match[1], window[1] + strings.ToLower(window[2])
}

// joinDetails joins the non-empty parts with ", " for the .Details field.
func joinDetails(parts ...string) string {
	kept := make([]string, 0, len(parts))

	for _, part := range parts {
		if part != "" {
			kept = append(kept, part)
		}
	}

	return strings.Join(kept, ", ")
}
