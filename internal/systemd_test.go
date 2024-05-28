package internal

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestIsServiceActive(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name    string
		service string
		status  bool
	}{
		{
			name:    "Snapd service check",
			service: "snapd",
			status:  IsSystemd(), // only if test is run under systemd
		},
		{
			name:    "Non-existing service check",
			service: "blablabla",
			status:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rez := IsServiceActive(tt.service)
			assert.Equal(t, tt.status, rez)
		})
	}
}

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
