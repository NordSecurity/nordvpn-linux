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
				State:      "Connected",
				Technology: config.Technology_NORDLYNX,
				Protocol:   config.Protocol_UDP,
				Hostname:   "Verona",
				Ip:         "127.0.0.1",
				Country:    "Lithuania",
				City:       "Vilnius",
				Download:   69,
				Upload:     69,
				Uptime:     13e9,
			},
			expected: `Status: Connected
Hostname: Verona
IP: 127.0.0.1
Country: Lithuania
City: Vilnius
Current technology: NORDLYNX
Current protocol: UDP
Post-quantum VPN: disabled
Transfer: 69 B received, 69 B sent
Uptime: 13 seconds
`,
		},
		{
			name: "disconnected",
			resp: &pb.StatusResponse{
				State:  "Disconnected",
				Uptime: -1,
			},
			expected: `Status: Disconnected
`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, Status(test.resp))
		})
	}
}
