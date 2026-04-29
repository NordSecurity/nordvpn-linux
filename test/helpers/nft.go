package helpers

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

func ListTable(tableName string) []string {
	return []string{"list", "table", "inet", tableName}
}

func WithNftCommandOutput(t *testing.T, args []string, fn func(out string)) {
	t.Helper()
	out, err := exec.Command("nft", args...).Output()
	require.NoError(t, err)
	fn(string(out))
}
