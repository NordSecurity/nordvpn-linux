package golden

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/stretchr/testify/require"
)

const UpdateGoldenEnvVar = "UPDATE_GOLDEN"

var update = os.Getenv(UpdateGoldenEnvVar) != ""

func AssertMatchesGolden(t *testing.T, got string) {
	t.Helper()
	path := goldenFilePath(t)
	dir := filepath.Dir(path)

	if update {
		require.NoError(t, os.MkdirAll(dir, 0o750))
		require.NoError(t, os.WriteFile(path, []byte(got), 0o600))
		return
	}

	golden, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		t.Fatalf(
			"golden file %s not found — run with %s=1 to create it",
			UpdateGoldenEnvVar,
			path,
		)
	}
	require.NoError(t, err)

	if string(golden) == got {
		return
	}

	diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(string(golden)),
		B:        difflib.SplitLines(got),
		FromFile: "golden",
		ToFile:   "actual",
		Context:  3,
	})
	t.Fatalf("ruleset does not match golden file %s:\n%s", path, diff)
}

// goldenFilePath derives the path from t.Name().
// For a subtest "TestVPNRuleset/kill_switch_only" the result is
// "testdata/TestVPNRuleset/kill_switch_only.nft".
func goldenFilePath(t *testing.T) string {
	t.Helper()
	parts := strings.SplitN(t.Name(), "/", 2)
	if len(parts) != 2 {
		t.Fatalf("goldenFilePath: %s must be a subtest (contain '/')", t.Name())
	}
	return filepath.Join("testdata", parts[0], parts[1]+".nft")
}
