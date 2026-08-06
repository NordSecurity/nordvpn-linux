package internal

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestIsSystemShutdown(t *testing.T) {
	category.Set(t, category.Unit)

	assert.False(t, IsSystemShutdown())
}

func TestIsSystemd(t *testing.T) {
	category.Set(t, category.Unit)

	// alternative way of detecting
	out, err := exec.Command("ps", "--no-headers", "-o", "comm", "1").CombinedOutput()
	str := strings.Trim(strings.Trim(string(out), "\n"), " ")
	isSystemd := err == nil && str == "systemd"

	assert.Equal(t, isSystemd, IsSystemd())
}
