package sysinfo

import (
	"errors"
	"os/exec"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func Test_TestGetHostArchitecture(t *testing.T) {
	category.Set(t, category.Integration)

	out, _ := exec.Command("uname", "-m").Output()
	result := strings.TrimSpace(string(out))
	assert.Equal(t, result, GetHostArchitecture())
}

func Test_getHostArchitecture(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		runner   cmdRunner
		expected string
	}{
		{
			name:     "architecture returned",
			runner:   func(string, ...string) (string, error) { return "x86_64", nil },
			expected: "x86_64",
		},
		{
			name:     "command fails",
			runner:   func(string, ...string) (string, error) { return "", errors.New("fail") },
			expected: "",
		},
		{
			name:     "empty output",
			runner:   func(string, ...string) (string, error) { return "", nil },
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, getHostArchitecture(tt.runner))
		})
	}
}
