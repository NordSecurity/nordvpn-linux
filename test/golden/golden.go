package golden

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/stretchr/testify/require"
)

const UpdateGoldenFilesEnvVar = "UPDATE_GOLDEN_FILES"

var update = os.Getenv(UpdateGoldenFilesEnvVar) != ""

// AssertMatchesGolden compares passed string with file located based on
// convention. Has to be called from subtest.
// For a subtest "TestVPNRuleset/kill switch only" the result is
// "testdata/TestVPNRuleset/kill_switch_only.nft" (gotest replaces spaces
// with "_").
func AssertMatchesGolden(t *testing.T, got string) {
	t.Helper()
	dir, name := goldenFileParts(t)

	if update {
		root, err := os.OpenRoot(".")
		require.NoError(t, err)
		defer root.Close()
		require.NoError(t, root.MkdirAll(filepath.Join("testdata", dir), 0o750))

		f, err := root.OpenFile(
			filepath.Join("testdata", dir, name),
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
			0o600,
		)
		require.NoError(t, err)
		defer f.Close()

		_, err = f.WriteString(got)
		require.NoError(t, err)

		return
	}

	root, err := os.OpenRoot("testdata")
	require.NoError(t, err)
	defer root.Close()

	f, err := root.Open(filepath.Join(dir, name))
	if errors.Is(err, fs.ErrNotExist) {
		t.Fatalf(
			"golden file testdata/%s/%s not found — run with %s=1 to create it",
			dir, name, UpdateGoldenFilesEnvVar,
		)
	}
	require.NoError(t, err)
	defer f.Close()

	data, err := io.ReadAll(f)
	require.NoError(t, err)

	if string(data) == got {
		return
	}

	diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(string(data)),
		B:        difflib.SplitLines(got),
		FromFile: "golden",
		ToFile:   "actual",
		Context:  3,
	})
	t.Fatalf("ruleset does not match golden file testdata/%s/%s:\n%s", dir, name, diff)
}

func goldenFileParts(t *testing.T) (dir, name string) {
	t.Helper()
	parts := strings.SplitN(t.Name(), "/", 2)
	if len(parts) != 2 {
		t.Fatalf("goldenFileParts: %s must be a subtest (contain '/')", t.Name())
	}
	return parts[0], parts[1] + ".txt"
}
