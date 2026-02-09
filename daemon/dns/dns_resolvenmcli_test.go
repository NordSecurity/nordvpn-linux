package dns

import (
	"fmt"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func Test_DSNResolveNmCli_PhysicalInterfaceDeduction(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		name          string
		cmdOutput     string
		cmdError      error
		expectedConns []string
	}{
		{
			name: "multiple physical connections",
			cmdOutput: `Wired-connection-1:802-3-ethernet
	Wi-Fi:802-11-wireless
	VPN:vpn
	Docker:bridge`,
			cmdError:      nil,
			expectedConns: []string{"Wired-connection-1", "Wi-Fi"},
		},
		{
			name: "wireless and gsm connections",
			cmdOutput: `Mobile Broadband:gsm
	Wi-Fi Network:802-11-wireless
	Loopback:loopback`,
			cmdError:      nil,
			expectedConns: []string{"Mobile Broadband", "Wi-Fi Network"},
		},
		{
			name: "ethernet and cdma connections",
			cmdOutput: `eth0:802-3-ethernet
	CDMA Connection:cdma
	tun0:tun`,
			cmdError:      nil,
			expectedConns: []string{"eth0", "CDMA Connection"},
		},
		{
			name:          "no physical connections",
			cmdOutput:     "VPN:vpn\nDocker:bridge\nLoopback:loopback",
			cmdError:      nil,
			expectedConns: []string{},
		},
		{
			name:          "empty output",
			cmdOutput:     "",
			cmdError:      nil,
			expectedConns: []string{},
		},
		{
			name:          "command execution error",
			cmdOutput:     "",
			cmdError:      fmt.Errorf("nmcli command failed"),
			expectedConns: []string{},
		},
		{
			name:          "malformed output - single field",
			cmdOutput:     "InvalidLine\neth0:802-3-ethernet",
			cmdError:      nil,
			expectedConns: []string{"eth0"},
		},
		{
			name: "connections with whitespace in names",
			cmdOutput: `  Wired connection 1  :802-3-ethernet
	  Wi-Fi Network  :802-11-wireless`,
			cmdError:      nil,
			expectedConns: []string{"Wired connection 1", "Wi-Fi Network"},
		},
		{
			name: "connections with semicolon in names",
			cmdOutput: `Wired connection_:  2:802-3-ethernet
docker0:bridge
lo:loopback
virbr0:bridge
br-f78c0ce0d3eb:bridge
mpqemubr0:bridge
vnet3:tun`,
			cmdError:      nil,
			expectedConns: []string{"Wired connection_:  2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nmcli := &NMCli{
				cmdExecutor: func(name string, arg ...string) ([]byte, error) {
					return []byte(tt.cmdOutput), tt.cmdError
				},
			}

			conns, err := nmcli.getConnectionFromPhysicalInterfaces()
			if tt.cmdError != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.cmdError.Error())
			}
			assert.Equal(t, tt.expectedConns, conns)
		})
	}
}
