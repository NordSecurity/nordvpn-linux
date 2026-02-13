package dns

import (
	"fmt"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func Test_DNSResolveNmCli_PhysicalInterfaceDeduction(t *testing.T) {
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
			expectedConns: []string{"Wired-connection-1", "Wi-Fi", "Docker"},
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
			name:          "with no physical connections, bridge picked up",
			cmdOutput:     "VPN:vpn\nDocker:bridge\nLoopback:loopback",
			cmdError:      nil,
			expectedConns: []string{"Docker"},
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
virbr0:tun
br-f78c0ce0d3eb:tun
mpqemubr0:bridge
vnet3:tun`,
			cmdError:      nil,
			expectedConns: []string{"Wired connection_:  2", "docker0", "mpqemubr0"},
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

func Test_DNSResolveNmCli_RollbackOnReloadFailure(t *testing.T) {
	category.Set(t, category.Integration)

	type connectionState struct {
		ipv4DNS       string
		ignoreAutoDNS string
	}
	// Mock state representing DNS configuration per connection
	type mockState struct {
		connections         map[string]connectionState
		failOnReloadForConn string
	}

	tests := []struct {
		name             string
		initialState     mockState
		nameservers      []string
		expectedError    bool
		expectedFinalDNS map[string]connectionState
	}{
		{
			name: "rollback on modify failure",
			initialState: mockState{
				connections: map[string]connectionState{
					"Wired-connection-1": {ipv4DNS: "1.2.3.4", ignoreAutoDNS: "no"},
					"Wi-Fi":              {ipv4DNS: "5.6.7.8", ignoreAutoDNS: "yes"},
				},
				failOnReloadForConn: "Wi-Fi",
			},
			nameservers:   []string{"1.1.1.1", "8.8.8.8"},
			expectedError: true,
			expectedFinalDNS: map[string]connectionState{
				"Wired-connection-1": {ipv4DNS: "1.2.3.4", ignoreAutoDNS: "no"},
				"Wi-Fi":              {ipv4DNS: "5.6.7.8", ignoreAutoDNS: "yes"},
			},
		},
		{
			name: "happy path scenario, no rollback",
			initialState: mockState{
				connections: map[string]connectionState{
					"Wired-connection-1": {ipv4DNS: "1.2.3.4", ignoreAutoDNS: "yes"},
					"Wi-Fi":              {ipv4DNS: "5.6.7.8", ignoreAutoDNS: "no"},
				},
				failOnReloadForConn: "",
			},
			nameservers:   []string{"1.1.1.1", "8.8.8.8"},
			expectedError: false,
			expectedFinalDNS: map[string]connectionState{
				"Wired-connection-1": {ipv4DNS: "1.1.1.1,8.8.8.8", ignoreAutoDNS: "yes"},
				"Wi-Fi":              {ipv4DNS: "1.1.1.1,8.8.8.8", ignoreAutoDNS: "yes"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currentState := mockState{
				connections:         make(map[string]connectionState),
				failOnReloadForConn: tt.initialState.failOnReloadForConn,
			}
			for k, v := range tt.initialState.connections {
				currentState.connections[k] = v
			}

			nmcli := &NMCli{
				cmdExecutor: func(name string, args ...string) ([]byte, error) {
					// Simulate getConnectionFromPhysicalInterfaces, to get the list of connections
					if len(args) > 2 && args[0] == "-t" && args[2] == "NAME,TYPE" {
						var output string
						for connName := range currentState.connections {
							// this is not quite relevant for this test suite
							output += fmt.Sprintf("%s:802-3-ethernet\n", connName)
						}
						return []byte(output), nil
					}

					// Simulate call to getConnectionState to get ipv4.dns value
					if len(args) > 2 && args[0] == "-t" && args[2] == nmCliIPv4DNSKey {
						connName := args[5]
						if conn, exists := currentState.connections[connName]; exists {
							return []byte(fmt.Sprintf("%s:%s", nmCliIPv4DNSKey, conn.ipv4DNS)), nil
						}
						return nil, fmt.Errorf("connection not found")
					}

					// Simulate call to getConnectionState, to get ignore-auto-dns value
					if len(args) > 2 && args[0] == "-t" && args[2] == nmCliIPIgnoreAutoDnsKey {
						connName := args[5]
						if conn, exists := currentState.connections[connName]; exists {
							return []byte(fmt.Sprintf("%s:%s", nmCliIPIgnoreAutoDnsKey, conn.ignoreAutoDNS)), nil
						}
						return nil, fmt.Errorf("connection not found")
					}

					// Simulate: call to SetDNS
					if len(args) > 2 && args[0] == nmCliConKey && args[1] == "modify" {
						connName := args[2]

						// Fail on specified connection
						if connName == tt.initialState.failOnReloadForConn {
							return nil, fmt.Errorf("failed to modify connection %s", connName)
						}

						// Extract DNS and ignore-auto-dns values from args
						for i := 3; i < len(args); i++ {
							if args[i] == nmCliIPv4DNSKey && i+1 < len(args) {
								conn := currentState.connections[connName]
								conn.ipv4DNS = args[i+1]
								currentState.connections[connName] = conn
							}
							if args[i] == nmCliIPIgnoreAutoDnsKey && i+1 < len(args) {
								conn := currentState.connections[connName]
								conn.ignoreAutoDNS = args[i+1]
								currentState.connections[connName] = conn
							}
						}
						return []byte("ok"), nil
					}

					return []byte("ok"), nil
				},
			}

			err := nmcli.Set("", tt.nameservers)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			for connName, expectedDNS := range tt.expectedFinalDNS {
				assert.Equal(t, expectedDNS, currentState.connections[connName],
					"Connection %s should have DNS restored to %s", connName, expectedDNS)
			}
		})
	}
}
