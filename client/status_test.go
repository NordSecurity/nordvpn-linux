package client

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestSpecialtyGroupLabel(t *testing.T) {
	category.Set(t, category.Unit)

	connectedTo := func(group config.ServerGroup) *pb.StatusResponse {
		return &pb.StatusResponse{
			State:      pb.ConnectionState_CONNECTED,
			Parameters: &pb.ConnectionParameters{Group: group},
		}
	}

	tests := []struct {
		name     string
		status   *pb.StatusResponse
		expected string
	}{
		{name: "double vpn", status: connectedTo(config.ServerGroup_DOUBLE_VPN), expected: "Double VPN"},
		{name: "onion over vpn", status: connectedTo(config.ServerGroup_ONION_OVER_VPN), expected: "Onion Over VPN"},
		{name: "dedicated ip", status: connectedTo(config.ServerGroup_DEDICATED_IP), expected: "Dedicated IP"},
		{name: "obfuscated group", status: connectedTo(config.ServerGroup_OBFUSCATED), expected: "Obfuscated Servers"},
		{name: "standard vpn servers", status: connectedTo(config.ServerGroup_STANDARD_VPN_SERVERS), expected: ""},
		{name: "p2p", status: connectedTo(config.ServerGroup_P2P), expected: ""},
		{name: "no group", status: connectedTo(config.ServerGroup_UNDEFINED), expected: ""},
		{
			name:     "missing parameters",
			status:   &pb.StatusResponse{State: pb.ConnectionState_CONNECTED},
			expected: "",
		},
		{
			name:     "obfuscated flag without group",
			status:   &pb.StatusResponse{State: pb.ConnectionState_CONNECTED, Obfuscated: true},
			expected: "Obfuscated Servers",
		},
		{
			name: "obfuscated flag wins over group",
			status: &pb.StatusResponse{
				State:      pb.ConnectionState_CONNECTED,
				Obfuscated: true,
				Parameters: &pb.ConnectionParameters{Group: config.ServerGroup_DOUBLE_VPN},
			},
			expected: "Obfuscated Servers",
		},
		{
			name: "meshnet peer ignores stale group",
			status: &pb.StatusResponse{
				State:      pb.ConnectionState_CONNECTED,
				IsMeshPeer: true,
				Parameters: &pb.ConnectionParameters{Group: config.ServerGroup_DOUBLE_VPN},
			},
			expected: "",
		},
		{
			name: "connecting",
			status: &pb.StatusResponse{
				State:      pb.ConnectionState_CONNECTING,
				Parameters: &pb.ConnectionParameters{Group: config.ServerGroup_DOUBLE_VPN},
			},
			expected: "",
		},
		{name: "nil status", status: nil, expected: ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, SpecialtyGroupLabel(test.status))
		})
	}
}
