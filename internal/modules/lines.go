package modules

import (
	"fmt"

	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
)

// LinesChangedModule renders lines added and removed with independent styles.
type LinesChangedModule struct{}

func (LinesChangedModule) Name() string { return "lines_changed" }

func (LinesChangedModule) Render(data input.Data, cfg config.Config) (string, error) {
	added := data.Cost.TotalLinesAdded
	removed := data.Cost.TotalLinesRemoved

	if added == 0 && removed == 0 {
		return "", nil
	}

	addedStr := wrapStyle(fmt.Sprintf("+%d", added), cfg.LinesChanged.AddedStyle)
	removedStr := wrapStyle(fmt.Sprintf("-%d", removed), cfg.LinesChanged.RemovedStyle)

	return addedStr + " " + removedStr, nil
}
