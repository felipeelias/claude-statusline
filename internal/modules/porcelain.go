package modules

import (
	"fmt"
	"strings"
)

// GitStatus holds parsed output from git status --porcelain=v2 --branch.
type GitStatus struct {
	Branch    string
	Staged    int
	Modified  int
	Untracked int
	Ahead     int
	Behind    int
	Conflicts int
}

// ParsePorcelainV2 parses the output of git status --porcelain=v2 --branch
// into a GitStatus struct.
func ParsePorcelainV2(output string) GitStatus {
	var status GitStatus

	for line := range strings.SplitSeq(output, "\n") {
		switch {
		case strings.HasPrefix(line, "# branch.head "):
			status.Branch = strings.TrimPrefix(line, "# branch.head ")

		case strings.HasPrefix(line, "# branch.ab "):
			_, _ = fmt.Sscanf(line, "# branch.ab +%d -%d", &status.Ahead, &status.Behind)

		case strings.HasPrefix(line, "1 ") || strings.HasPrefix(line, "2 "):
			parseChangedEntry(line, &status)

		case strings.HasPrefix(line, "u "):
			status.Conflicts++

		case strings.HasPrefix(line, "? "):
			status.Untracked++
		}
	}

	return status
}

// parseChangedEntry parses an ordinary (1) or rename/copy (2) porcelain v2 entry.
func parseChangedEntry(line string, status *GitStatus) {
	// "1 XY sub ..." or "2 XY sub ..." — XY is at field index 1.
	statusXY, _, found := strings.Cut(line[2:], " ")
	if !found || len(statusXY) != 2 {
		return
	}

	if statusXY[0] != '.' {
		status.Staged++
	}

	if statusXY[1] != '.' {
		status.Modified++
	}
}
