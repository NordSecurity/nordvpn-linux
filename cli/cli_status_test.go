package cli

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestStatus(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		resp     *pb.StatusResponse
		expected string
	}{
		{
			name: "connected",
			resp: &pb.StatusResponse{
				State:       pb.ConnectionState_CONNECTED,
				VpnProtocol: config.VPNProtocol_VPN_PROTOCOL_NORDLYNX,
				Hostname:    "Verona",
				Ip:          "127.0.0.1",
				Country:     "Lithuania",
				City:        "Vilnius",
				Download:    69,
				Upload:      69,
				Uptime:      13e9,
			},
			expected: `Status: Connected
Hostname: Verona
IP: 127.0.0.1
Country: Lithuania
City: Vilnius
Protocol: NordLynx
Post-quantum VPN: Disabled
Transfer: 69 B received, 69 B sent
Uptime: 13 seconds
`,
		},
		{
			name: "nordwhisper with ech enabled",
			resp: &pb.StatusResponse{
				State:       pb.ConnectionState_CONNECTED,
				VpnProtocol: config.VPNProtocol_VPN_PROTOCOL_NORDWHISPER,
				Hostname:    "Verona",
				Ip:          "127.0.0.1",
				Country:     "Lithuania",
				City:        "Vilnius",
				Uptime:      13e9,
				Ech:         true,
			},
			expected: `Status: Connected
Hostname: Verona
IP: 127.0.0.1
Country: Lithuania
City: Vilnius
Protocol: NordWhisper
ECH: Enabled
Uptime: 13 seconds
`,
		},
		{
			name: "nordwhisper with ech disabled",
			resp: &pb.StatusResponse{
				State:       pb.ConnectionState_CONNECTED,
				VpnProtocol: config.VPNProtocol_VPN_PROTOCOL_NORDWHISPER,
				Hostname:    "Verona",
				Ip:          "127.0.0.1",
				Country:     "Lithuania",
				City:        "Vilnius",
				Uptime:      13e9,
				Ech:         false,
			},
			expected: `Status: Connected
Hostname: Verona
IP: 127.0.0.1
Country: Lithuania
City: Vilnius
Protocol: NordWhisper
ECH: Disabled
Uptime: 13 seconds
`,
		},
		{
			name: "non-nordwhisper hides ech line even when ech is set",
			resp: &pb.StatusResponse{
				State:       pb.ConnectionState_CONNECTED,
				VpnProtocol: config.VPNProtocol_VPN_PROTOCOL_NORDLYNX,
				Hostname:    "Verona",
				Ip:          "127.0.0.1",
				Country:     "Lithuania",
				City:        "Vilnius",
				Uptime:      13e9,
				Ech:         true,
			},
			expected: `Status: Connected
Hostname: Verona
IP: 127.0.0.1
Country: Lithuania
City: Vilnius
Protocol: NordLynx
Post-quantum VPN: Disabled
Uptime: 13 seconds
`,
		},
		{
			name: "connected to specialty group",
			resp: &pb.StatusResponse{
				State:       pb.ConnectionState_CONNECTED,
				VpnProtocol: config.VPNProtocol_VPN_PROTOCOL_NORDLYNX,
				Hostname:    "Verona",
				Ip:          "127.0.0.1",
				Country:     "Lithuania",
				City:        "Vilnius",
				Uptime:      13e9,
				Parameters:  &pb.ConnectionParameters{Group: config.ServerGroup_DOUBLE_VPN},
			},
			expected: `Status: Connected
Hostname: Verona
IP: 127.0.0.1
Country: Lithuania
City: Vilnius
Group: Double VPN
Protocol: NordLynx
Post-quantum VPN: Disabled
Uptime: 13 seconds
`,
		},
		{
			name: "connected to non-specialty group hides group line",
			resp: &pb.StatusResponse{
				State:       pb.ConnectionState_CONNECTED,
				VpnProtocol: config.VPNProtocol_VPN_PROTOCOL_NORDLYNX,
				Hostname:    "Verona",
				Uptime:      13e9,
				Parameters:  &pb.ConnectionParameters{Group: config.ServerGroup_STANDARD_VPN_SERVERS},
			},
			expected: `Status: Connected
Hostname: Verona
Protocol: NordLynx
Post-quantum VPN: Disabled
Uptime: 13 seconds
`,
		},
		{
			name: "connected to obfuscated server",
			resp: &pb.StatusResponse{
				State:       pb.ConnectionState_CONNECTED,
				VpnProtocol: config.VPNProtocol_VPN_PROTOCOL_NORDWHISPER,
				Hostname:    "Verona",
				Uptime:      13e9,
				Obfuscated:  true,
			},
			expected: `Status: Connected
Hostname: Verona
Group: Obfuscated
Protocol: NordWhisper
ECH: Disabled
Uptime: 13 seconds
`,
		},
		{
			name: "disconnected",
			resp: &pb.StatusResponse{
				State:  pb.ConnectionState_DISCONNECTED,
				Uptime: -1,
			},
			expected: `Status: Disconnected
`,
		},
		{
			name: "paused hours",
			resp: &pb.StatusResponse{
				State:                     pb.ConnectionState_PAUSED,
				PauseRemainingDurationSec: 5055,
				Uptime:                    -1,
			},
			expected: `Status: Paused
Pause time left: 01:24:15
`,
		},
		{
			name: "paused minutes",
			resp: &pb.StatusResponse{
				State:                     pb.ConnectionState_PAUSED,
				PauseRemainingDurationSec: 1964,
				Uptime:                    -1,
			},
			expected: `Status: Paused
Pause time left: 32:44
`,
		},
		{
			name: "paused seconds",
			resp: &pb.StatusResponse{
				State:                     pb.ConnectionState_PAUSED,
				PauseRemainingDurationSec: 23,
				Uptime:                    -1,
			},
			expected: `Status: Paused
Pause time left: 00:23
`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, Status(test.resp))
		})
	}
}
