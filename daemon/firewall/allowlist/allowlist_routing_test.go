package allowlist

import (
	"errors"
	"fmt"
	"os/exec"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

const (
	mark     = "0x123"
	port     = "22"
	protocol = "tcp"
)

func workingCommandFunc(command string, arg ...string) ([]byte, error) {
	arg = append(arg, "-w", internal.SecondsToWaitForIptablesLock)
	return exec.Command(command, arg...).CombinedOutput()
}

func failingCommandFunc(command string, arg ...string) ([]byte, error) {
	return nil, fmt.Errorf("failing command func")
}

type MockCmd struct {
	called bool
}

func (m MockCmd) onlyOnceCommandFunc(string, ...string) ([]byte, error) {
	if m.called {
		return nil, errors.New("already called")
	}
	m.called = true
	return []byte("ok"), nil
}

func TestIPTables_routingPorts(t *testing.T) {
	category.Set(t, category.Route)

	got, err := checkRouting(workingCommandFunc, port, mark)
	assert.NoError(t, err)
	assert.False(t, got)

	err = routePortsToIPTables(workingCommandFunc, port, protocol, mark)

	assert.NoError(t, err)

	got, err = checkRouting(workingCommandFunc, port, mark)

	assert.NoError(t, err)
	assert.True(t, got)

	err = clearRouting(workingCommandFunc)

	assert.NoError(t, err)
}

func Test_getCleanupIPTablesRules(t *testing.T) {
	type args struct {
		commandFunc runCommandFunc
		chain       string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Cleanup of iptables rules",
			args: args{
				commandFunc: workingCommandFunc,
				chain:       "0x123",
			},
			wantErr: true,
		},
		{
			name: "Failing cleanup",
			args: args{
				commandFunc: failingCommandFunc,
				chain:       "0x123",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := getCleanupIPTablesRules(tt.args.commandFunc, tt.args.chain); (err != nil) != tt.wantErr {
				t.Errorf("getCleanupIPTablesRules() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIPTables_EnablePorts(t *testing.T) {
	category.Set(t, category.Route)

	type args struct {
		ports       []int
		protocol    string
		commandFunc runCommandFunc
		mark        string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Add port routing",
			args: args{
				ports:       []int{33},
				protocol:    "tcp",
				mark:        "0x123",
				commandFunc: workingCommandFunc,
			},
			wantErr: false,
		},
		{
			name: "Failing port routing",
			args: args{
				ports:       []int{33},
				protocol:    "tcp",
				mark:        "0x123",
				commandFunc: failingCommandFunc,
			},
			wantErr: true,
		},
		{
			name: "Multiple port routing",
			args: args{
				ports:       []int{33, 34, 35, 36, 37, 38, 39},
				protocol:    "tcp",
				mark:        "0x123",
				commandFunc: MockCmd{}.onlyOnceCommandFunc,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wh := NewAllowlistRouting(tt.args.commandFunc)
			if err := wh.EnablePorts(tt.args.ports, tt.args.protocol, tt.args.mark); (err != nil) != tt.wantErr {
				t.Errorf("Error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIPTables_Disable(t *testing.T) {
	category.Set(t, category.Route)

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "Delete allowlist routing rules",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wh := NewAllowlistRouting(workingCommandFunc)
			if err := wh.Disable(); (err != nil) != tt.wantErr {
				t.Errorf("Error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
