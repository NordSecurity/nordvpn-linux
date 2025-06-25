package sysinfo

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func Test_HostInfo(t *testing.T) {
	category.Set(t, category.Integration)

	out, _ := exec.Command("uname", "-a").Output()
	result := strings.TrimSpace(string(out))
	assert.Equal(t, result, GetHostInfo())
}
