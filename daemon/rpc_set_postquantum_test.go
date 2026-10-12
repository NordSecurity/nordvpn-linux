package daemon

import (
	"context"
	"strconv"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/events"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/mock/networker"
	"github.com/stretchr/testify/assert"
)

type mockPostquantumVpnConfigManager struct {
	c config.Config
}

func (*mockPostquantumVpnConfigManager) SaveWith(f config.SaveFunc) error {
	return nil
}

func (m *mockPostquantumVpnConfigManager) Load(c *config.Config) error {
	c.Mesh = m.c.Mesh
	c.AutoConnect = m.c.AutoConnect
	c.AutoConnectData = m.c.AutoConnectData
	c.VPNProtocol = m.c.VPNProtocol
	return nil
}

func (*mockPostquantumVpnConfigManager) Reset(bool, bool) error {
	return nil
}

func TestSetPostquantumVpn(t *testing.T) {
	mockConfigManager := mockPostquantumVpnConfigManager{c: config.Config{}}

	mockPublisherSubscriber := events.MockPublisherSubscriber[bool]{}
	mockEvents := events.Events{Settings: &events.SettingsEvents{PostquantumVPN: &mockPublisherSubscriber}}
	mockNetworker := networker.Mock{}

	r := RPC{
		cm:     &mockConfigManager,
		events: &mockEvents,
		netw:   &mockNetworker,
	}

	successPayload := pb.Payload{
		Type: internal.CodeSuccess,
		Data: []string{strconv.FormatBool(false)},
	}

	successWithVPNPayload := pb.Payload{
		Type: internal.CodeSuccess,
		Data: []string{strconv.FormatBool(true)},
	}

	conflictMeshPayload := pb.Payload{
		Type: internal.CodePqAndMeshnetSimultaneously,
	}

	conflictUnknownProtocolPayload := pb.Payload{
		Type: internal.CodePqWithoutNordlynx,
		Data: []string{"VPN_PROTOCOL_UNSPECIFIED"},
	}

	conflictOpenVPNProtocolPayload := pb.Payload{
		Type: internal.CodePqWithoutNordlynx,
		Data: []string{"OpenVPN (UDP)"},
	}

	conflictNordWhisperPayload := pb.Payload{
		Type: internal.CodePqWithoutNordlynx,
		Data: []string{"NordWhisper"},
	}

	conflictDedicatedServerPayload := pb.Payload{
		Type: internal.CodeDedicatedServersPq,
	}

	tests := []struct {
		testName               string
		pq                     bool
		meshnet                bool
		vpnActive              bool
		vpnProtocol            config.VPNProtocol
		autoconnect            bool
		autoconnectTargetGroup config.ServerGroup
		payload                *pb.Payload
		eventPublished         bool
	}{
		{
			testName:       "pq off mesh is off protocol unknown",
			pq:             false,
			meshnet:        false,
			vpnActive:      false,
			vpnProtocol:    config.VPNProtocol_VPN_PROTOCOL_UNSPECIFIED,
			payload:        &conflictUnknownProtocolPayload,
			eventPublished: false,
		},
		{
			testName:       "pq off mesh is off protocol nlx",
			pq:             false,
			meshnet:        false,
			vpnActive:      false,
			vpnProtocol:    config.VPNProtocol_VPN_PROTOCOL_NORDLYNX,
			payload:        &successPayload,
			eventPublished: true,
		},
		{
			testName:       "pq on mesh is off protocol unknown",
			pq:             true,
			meshnet:        false,
			vpnActive:      false,
			vpnProtocol:    config.VPNProtocol_VPN_PROTOCOL_UNSPECIFIED,
			payload:        &conflictUnknownProtocolPayload,
			eventPublished: false,
		},
		{
			testName:       "pq on mesh is off protocol nlx",
			pq:             true,
			meshnet:        false,
			vpnActive:      false,
			vpnProtocol:    config.VPNProtocol_VPN_PROTOCOL_NORDLYNX,
			payload:        &successPayload,
			eventPublished: true,
		},
		{
			testName:       "pq on mesh is on",
			pq:             true,
			meshnet:        true,
			vpnActive:      false,
			payload:        &conflictMeshPayload,
			eventPublished: false,
		},
		{
			testName:       "pq off mesh is off protocol nlx vpn on",
			pq:             false,
			meshnet:        false,
			vpnActive:      true,
			vpnProtocol:    config.VPNProtocol_VPN_PROTOCOL_NORDLYNX,
			payload:        &successWithVPNPayload,
			eventPublished: true,
		},
		{
			testName:       "pq on mesh is off protocol openvpn",
			pq:             true,
			meshnet:        false,
			vpnActive:      false,
			vpnProtocol:    config.VPNProtocol_VPN_PROTOCOL_OPENVPN_UDP,
			payload:        &conflictOpenVPNProtocolPayload,
			eventPublished: false,
		},
		{
			testName:       "pq on mesh is off protocol nordwhisper",
			pq:             true,
			meshnet:        false,
			vpnActive:      false,
			vpnProtocol:    config.VPNProtocol_VPN_PROTOCOL_NORDWHISPER,
			payload:        &conflictNordWhisperPayload,
			eventPublished: false,
		},
		{
			testName:               "pq is off autoconnect target is a dedicated server",
			pq:                     true,
			meshnet:                false,
			vpnActive:              false,
			autoconnect:            true,
			autoconnectTargetGroup: config.ServerGroup_DEDICATED_SERVER,
			vpnProtocol:            config.VPNProtocol_VPN_PROTOCOL_NORDLYNX,
			payload:                &conflictDedicatedServerPayload,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockConfigManager.c.Mesh = test.meshnet
			mockConfigManager.c.VPNProtocol = test.vpnProtocol
			mockConfigManager.c.AutoConnect = test.autoconnect
			mockConfigManager.c.AutoConnectData.Group = test.autoconnectTargetGroup
			mockConfigManager.c.AutoConnectData.PostquantumVpn = !test.pq

			mockNetworker.ConnectRetries = 0
			mockNetworker.VpnActive = test.vpnActive

			req := pb.SetGenericRequest{Enabled: test.pq}
			resp, err := r.SetPostQuantum(context.Background(), &req)

			assert.NoError(t, err)
			assert.Equal(t, test.payload, resp)
			assert.Equal(t, test.eventPublished, mockPublisherSubscriber.EventPublished)
			mockPublisherSubscriber.EventPublished = false
		})
	}
}
