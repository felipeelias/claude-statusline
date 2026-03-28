package modules_test

import (
	"testing"

	"github.com/felipeelias/claude-statusline/internal/modules"
	"github.com/stretchr/testify/assert"
)

func TestParsePorcelainV2_CleanBranch(t *testing.T) {
	output := "# branch.oid abc123\n# branch.head main\n"
	s := modules.ParsePorcelainV2(output)

	assert.Equal(t, "main", s.Branch)
	assert.Equal(t, 0, s.Staged)
	assert.Equal(t, 0, s.Modified)
	assert.Equal(t, 0, s.Untracked)
	assert.Equal(t, 0, s.Ahead)
	assert.Equal(t, 0, s.Behind)
	assert.Equal(t, 0, s.Conflicts)
	assert.Equal(t, 0, s.Staged+s.Modified+s.Untracked+s.Conflicts)
}

func TestParsePorcelainV2_StagedFiles(t *testing.T) {
	output := "# branch.oid abc123\n# branch.head main\n1 A. N... 000000 100644 100644 0000 abcd file.txt\n"
	s := modules.ParsePorcelainV2(output)

	assert.Equal(t, 1, s.Staged)
	assert.Equal(t, 0, s.Modified)
}

func TestParsePorcelainV2_ModifiedFiles(t *testing.T) {
	output := "# branch.oid abc123\n# branch.head main\n1 .M N... 100644 100644 100644 abcd abcd file.txt\n"
	s := modules.ParsePorcelainV2(output)

	assert.Equal(t, 0, s.Staged)
	assert.Equal(t, 1, s.Modified)
}

func TestParsePorcelainV2_StagedAndModified(t *testing.T) {
	output := "# branch.oid abc123\n# branch.head main\n1 MM N... 100644 100644 100644 abcd abcd file.txt\n"
	s := modules.ParsePorcelainV2(output)

	assert.Equal(t, 1, s.Staged)
	assert.Equal(t, 1, s.Modified)
}

func TestParsePorcelainV2_UntrackedFiles(t *testing.T) {
	output := "# branch.oid abc123\n# branch.head main\n? untracked.txt\n"
	s := modules.ParsePorcelainV2(output)

	assert.Equal(t, 1, s.Untracked)
}

func TestParsePorcelainV2_AheadBehind(t *testing.T) {
	output := "# branch.oid abc123\n# branch.head main\n# branch.upstream origin/main\n# branch.ab +3 -1\n"
	s := modules.ParsePorcelainV2(output)

	assert.Equal(t, 3, s.Ahead)
	assert.Equal(t, 1, s.Behind)
}

func TestParsePorcelainV2_NoUpstream(t *testing.T) {
	output := "# branch.oid abc123\n# branch.head feature\n"
	s := modules.ParsePorcelainV2(output)

	assert.Equal(t, "feature", s.Branch)
	assert.Equal(t, 0, s.Ahead)
	assert.Equal(t, 0, s.Behind)
}

func TestParsePorcelainV2_DetachedHead(t *testing.T) {
	output := "# branch.oid abc123def456\n# branch.head (detached)\n"
	s := modules.ParsePorcelainV2(output)

	assert.Equal(t, "(detached)", s.Branch)
}

func TestParsePorcelainV2_Conflicts(t *testing.T) {
	output := "# branch.oid abc123\n# branch.head main\nu UU N... 100644 100644 100644 100644 abcd efgh ijkl conflict.txt\n"
	s := modules.ParsePorcelainV2(output)

	assert.Equal(t, 1, s.Conflicts)
}

func TestParsePorcelainV2_RenamedFile(t *testing.T) {
	output := "# branch.oid abc123\n# branch.head main\n2 R. N... 100644 100644 100644 abcd abcd R100 new.txt\told.txt\n"
	s := modules.ParsePorcelainV2(output)

	assert.Equal(t, 1, s.Staged)
	assert.Equal(t, 0, s.Modified)
}

func TestParsePorcelainV2_MixedState(t *testing.T) {
	output := "# branch.oid abc123\n# branch.head main\n# branch.upstream origin/main\n# branch.ab +1 -0\n" +
		"1 A. N... 000000 100644 100644 0000 abcd staged.txt\n" +
		"1 .M N... 100644 100644 100644 abcd abcd modified.txt\n" +
		"? untracked.txt\n" +
		"u UU N... 100644 100644 100644 100644 abcd efgh ijkl conflict.txt\n"
	s := modules.ParsePorcelainV2(output)

	assert.Equal(t, "main", s.Branch)
	assert.Equal(t, 1, s.Staged)
	assert.Equal(t, 1, s.Modified)
	assert.Equal(t, 1, s.Untracked)
	assert.Equal(t, 1, s.Conflicts)
	assert.Equal(t, 1, s.Ahead)
	assert.Equal(t, 0, s.Behind)
}

func TestParsePorcelainV2_EmptyOutput(t *testing.T) {
	s := modules.ParsePorcelainV2("")

	assert.Equal(t, "", s.Branch)
	assert.Equal(t, 0, s.Staged+s.Modified+s.Untracked+s.Conflicts)
}

func TestParsePorcelainV2_InitialCommit(t *testing.T) {
	output := "# branch.oid (initial)\n# branch.head main\n? README.md\n"
	s := modules.ParsePorcelainV2(output)

	assert.Equal(t, "main", s.Branch)
	assert.Equal(t, 1, s.Untracked)
}
