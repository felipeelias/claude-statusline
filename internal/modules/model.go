package modules

import (
	"regexp"
	"strings"

	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
)

var modelIDPattern = regexp.MustCompile(`^claude-(opus|sonnet|haiku)-(\d+)-(\d+)(?:-\d+)?$`)

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

func (ModelModule) Render(data input.Data, cfg config.Config) (string, error) {
	displayName := data.Model.DisplayName
	if displayName == "" {
		return "", nil
	}

	templateData := struct {
		ID          string
		DisplayName string
		Short       string
	}{
		ID:          data.Model.ID,
		DisplayName: displayName,
		Short:       ShortName(data.Model.ID, displayName),
	}

	result, err := renderTemplate("model", cfg.Model.Format, templateData)
	if err != nil {
		return "", err
	}

	return wrapStyle(result, cfg.Model.Style), nil
}
