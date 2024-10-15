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
	c.AutoConnectData = m.c.AutoConnectData
	c.Technology = m.c.Technology
	return nil
}

func (*mockPostquantumVpnConfigManager) Reset() error {
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

	conflictTechPayload := pb.Payload{
		Type: internal.CodePqWithoutNordlynx,
	}

	tests := []struct {
		testName       string
		pq             bool
		meshnet        bool
		vpnActive      bool
		tech           config.Technology
		payload        *pb.Payload
		eventPublished bool
	}{
		{
			testName:       "pq off mesh is off tech unknown",
			pq:             false,
			meshnet:        false,
			vpnActive:      false,
			tech:           config.Technology_UNKNOWN_TECHNOLOGY,
			payload:        &conflictTechPayload,
			eventPublished: false,
		},
		{
			testName:       "pq off mesh is off tech nlx",
			pq:             false,
			meshnet:        false,
			vpnActive:      false,
			tech:           config.Technology_NORDLYNX,
			payload:        &successPayload,
			eventPublished: true,
		},
		{
			testName:       "pq on mesh is off tech unknown",
			pq:             true,
			meshnet:        false,
			vpnActive:      false,
			tech:           config.Technology_UNKNOWN_TECHNOLOGY,
			payload:        &conflictTechPayload,
			eventPublished: false,
		},
		{
			testName:       "pq on mesh is off tech nlx",
			pq:             true,
			meshnet:        false,
			vpnActive:      false,
			tech:           config.Technology_NORDLYNX,
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
			testName:       "pq off mesh is off tech nlx vpn on",
			pq:             false,
			meshnet:        false,
			vpnActive:      true,
			tech:           config.Technology_NORDLYNX,
			payload:        &successWithVPNPayload,
			eventPublished: true,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockConfigManager.c.Mesh = test.meshnet
			mockConfigManager.c.Technology = test.tech
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
