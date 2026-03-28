package modules_test

import (
	"testing"

	"github.com/felipeelias/claude-statusline/internal/modules"
	"github.com/stretchr/testify/assert"
)

func TestParsePorcelainV2_CleanBranch(t *testing.T) {
	output := "# branch.oid abc123\n# branch.head main\n"
	status := modules.ParsePorcelainV2(output)

	assert.Equal(t, "main", status.Branch)
	assert.Equal(t, 0, status.Staged)
	assert.Equal(t, 0, status.Modified)
	assert.Equal(t, 0, status.Untracked)
	assert.Equal(t, 0, status.Ahead)
	assert.Equal(t, 0, status.Behind)
	assert.Equal(t, 0, status.Conflicts)
}

func TestParsePorcelainV2_StagedFiles(t *testing.T) {
	output := "# branch.oid abc123\n# branch.head main\n" +
		"1 A. N... 000000 100644 100644 0000 abcd file.txt\n"
	status := modules.ParsePorcelainV2(output)

	assert.Equal(t, 1, status.Staged)
	assert.Equal(t, 0, status.Modified)
}

func TestParsePorcelainV2_ModifiedFiles(t *testing.T) {
	output := "# branch.oid abc123\n# branch.head main\n" +
		"1 .M N... 100644 100644 100644 abcd1234 efgh5678 file.txt\n"
	status := modules.ParsePorcelainV2(output)

	assert.Equal(t, 0, status.Staged)
	assert.Equal(t, 1, status.Modified)
}

func TestParsePorcelainV2_StagedAndModified(t *testing.T) {
	output := "# branch.oid abc123\n# branch.head main\n" +
		"1 MM N... 100644 100644 100644 abcd1234 efgh5678 file.txt\n"
	status := modules.ParsePorcelainV2(output)

	assert.Equal(t, 1, status.Staged)
	assert.Equal(t, 1, status.Modified)
}

func TestParsePorcelainV2_UntrackedFiles(t *testing.T) {
	output := "# branch.oid abc123\n# branch.head main\n? untracked.txt\n"
	status := modules.ParsePorcelainV2(output)

	assert.Equal(t, 1, status.Untracked)
}

func TestParsePorcelainV2_AheadBehind(t *testing.T) {
	output := "# branch.oid abc123\n# branch.head main\n" +
		"# branch.upstream origin/main\n# branch.ab +3 -1\n"
	status := modules.ParsePorcelainV2(output)

	assert.Equal(t, 3, status.Ahead)
	assert.Equal(t, 1, status.Behind)
}

func TestParsePorcelainV2_NoUpstream(t *testing.T) {
	output := "# branch.oid abc123\n# branch.head feature\n"
	status := modules.ParsePorcelainV2(output)

	assert.Equal(t, "feature", status.Branch)
	assert.Equal(t, 0, status.Ahead)
	assert.Equal(t, 0, status.Behind)
}

func TestParsePorcelainV2_DetachedHead(t *testing.T) {
	output := "# branch.oid abc123def456\n# branch.head (detached)\n"
	status := modules.ParsePorcelainV2(output)

	assert.Equal(t, "(detached)", status.Branch)
}

func TestParsePorcelainV2_Conflicts(t *testing.T) {
	output := "# branch.oid abc123\n# branch.head main\n" +
		"u UU N... 100644 100644 100644 100644 abcd efgh ijkl conflict.txt\n"
	status := modules.ParsePorcelainV2(output)

	assert.Equal(t, 1, status.Conflicts)
}

func TestParsePorcelainV2_RenamedFile(t *testing.T) {
	output := "# branch.oid abc123\n# branch.head main\n" +
		"2 R. N... 100644 100644 100644 abcd1234 efgh5678 R100 new.txt\told.txt\n"
	status := modules.ParsePorcelainV2(output)

	assert.Equal(t, 1, status.Staged)
	assert.Equal(t, 0, status.Modified)
}

func TestParsePorcelainV2_MixedState(t *testing.T) {
	output := "# branch.oid abc123\n# branch.head main\n" +
		"# branch.upstream origin/main\n# branch.ab +1 -0\n" +
		"1 A. N... 000000 100644 100644 0000 abcd staged.txt\n" +
		"1 .M N... 100644 100644 100644 abcd1234 efgh5678 modified.txt\n" +
		"? untracked.txt\n" +
		"u UU N... 100644 100644 100644 100644 abcd efgh ijkl conflict.txt\n"
	status := modules.ParsePorcelainV2(output)

	assert.Equal(t, "main", status.Branch)
	assert.Equal(t, 1, status.Staged)
	assert.Equal(t, 1, status.Modified)
	assert.Equal(t, 1, status.Untracked)
	assert.Equal(t, 1, status.Conflicts)
	assert.Equal(t, 1, status.Ahead)
	assert.Equal(t, 0, status.Behind)
}

func TestParsePorcelainV2_EmptyOutput(t *testing.T) {
	status := modules.ParsePorcelainV2("")

	assert.Empty(t, status.Branch)
	assert.Equal(t, 0, status.Staged+status.Modified+status.Untracked+status.Conflicts)
}

func TestParsePorcelainV2_InitialCommit(t *testing.T) {
	output := "# branch.oid (initial)\n# branch.head main\n? README.md\n"
	status := modules.ParsePorcelainV2(output)

	assert.Equal(t, "main", status.Branch)
	assert.Equal(t, 1, status.Untracked)
}
