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
	var s GitStatus

	for _, line := range strings.Split(output, "\n") {
		switch {
		case strings.HasPrefix(line, "# branch.head "):
			s.Branch = strings.TrimPrefix(line, "# branch.head ")

		case strings.HasPrefix(line, "# branch.ab "):
			_, _ = fmt.Sscanf(line, "# branch.ab +%d -%d", &s.Ahead, &s.Behind)

		case strings.HasPrefix(line, "1 ") || strings.HasPrefix(line, "2 "):
			// Ordinary (1) or rename/copy (2) entry: "1 XY ..." or "2 XY ..."
			fields := strings.SplitN(line, " ", 3)
			if len(fields) >= 2 && len(fields[1]) == 2 {
				xy := fields[1]
				if xy[0] != '.' {
					s.Staged++
				}
				if xy[1] != '.' {
					s.Modified++
				}
			}

		case strings.HasPrefix(line, "u "):
			s.Conflicts++

		case strings.HasPrefix(line, "? "):
			s.Untracked++
		}
	}

	return s
}
