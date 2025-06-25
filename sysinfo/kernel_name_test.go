package sysinfo

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func Test_KerneVersion(t *testing.T) {
	category.Set(t, category.Integration)

	out, _ := exec.Command("uname", "-sr").Output()
	result := strings.TrimSpace(string(out))
	assert.Equal(t, result, KernelVersion())
}

func mockCommandRunner(name string, args ...string) (string, error) {
	mockResults := map[string]string{
		"uname -s": "Linux",
		"uname -r": "6.11.0",
		"uname -a": "Linux hostname 6.11.0-26-generic #26~24.04.1-Ubuntu SMP PREEMPT_DYNAMIC",
	}
	cmd := fmt.Sprintf("%s %s", name, strings.Join(args, " "))
	if result, exists := mockResults[cmd]; exists {
		return result, nil
	}

	return "", fmt.Errorf("unsupported command: %s", cmd)
}

func Test_uname(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		flags    string
		expected string
	}{
		{"Kernel Name", "-s", "Linux"},
		{"Kernel Release", "-r", "6.11.0"},
		{"Full System Info", "-a", "Linux hostname 6.11.0-26-generic #26~24.04.1-Ubuntu SMP PREEMPT_DYNAMIC"},
		{"Unknown Flag", "-x", defaultKernelName},
		{"Empty Flag", "", defaultKernelName},
		{"Whitespace Flag", " ", defaultKernelName},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := uname(mockCommandRunner, tt.flags)
			if got != tt.expected {
				t.Errorf("uname(%q) = %q, want %q", tt.flags, got, tt.expected)
			}
		})
	}
}

func Test_defaultCmdRunner(t *testing.T) {
	category.Set(t, category.Integration)

	type args struct {
		name string
		args []string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"Valid Command - Uname Kernel Name", args{"uname", []string{"-s"}}, "Linux", false},
		{"Valid Command - Uname Operating System", args{"uname", []string{"-o"}}, "GNU/Linux", false},
		{"Invalid Command", args{"fakecmd", []string{}}, "", true},
		{"Invalid Flag", args{"uname", []string{"-x"}}, "", true},
		{"Empty Command", args{"", []string{}}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := defaultCmdRunner(tt.args.name, tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("defaultCmdRunner() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("defaultCmdRunner(%q, %q) = %q, want %q", tt.args.name, tt.args.args, got, tt.want)
			}
		})
	}
}
