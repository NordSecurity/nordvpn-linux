package helpers

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ListChain(chainName string) []string {
	return []string{"list", "chain", "inet", "nordvpn", chainName}
}

func AssertRulesOrder(t *testing.T, content, firstSubstr, secondSubstr string) {
	t.Helper()
	i := strings.Index(content, firstSubstr)
	j := strings.Index(content, secondSubstr)
	if i == -1 {
		t.Errorf("rule not found in output: %s", firstSubstr)
		return
	}
	if j == -1 {
		t.Errorf("rule not found in output: %s", secondSubstr)
		return
	}
	if i >= j {
		t.Errorf("expected rule:\n  %s\nto appear before:\n  %s", firstSubstr, secondSubstr)
	}
}

func RunNftCommand(t *testing.T, args ...string) string {
	t.Helper()
	out, err := exec.Command("nft", args...).Output()

	assert.NoError(t, err)
	return string(out)
}

func WithNftCommandOutput(t *testing.T, args []string, fn func(out string)) {
	t.Helper()
	out := RunNftCommand(t, args...)
	fn(out)
}
