package daemon

import (
	"context"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/events"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
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

	r := RPC{
		cm:     &mockConfigManager,
		events: &mockEvents,
	}

	successPayload := pb.Payload{
		Type: internal.CodeSuccess,
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
		tech           config.Technology
		payload        *pb.Payload
		eventPublished bool
	}{
		{
			testName:       "pq off mesh is off tech unknown",
			pq:             false,
			meshnet:        false,
			tech:           config.Technology_UNKNOWN_TECHNOLOGY,
			payload:        &conflictTechPayload,
			eventPublished: false,
		},
		{
			testName:       "pq off mesh is off tech nlx",
			pq:             false,
			meshnet:        false,
			tech:           config.Technology_NORDLYNX,
			payload:        &successPayload,
			eventPublished: true,
		},
		{
			testName:       "pq on mesh is off tech unknown",
			pq:             true,
			meshnet:        false,
			tech:           config.Technology_UNKNOWN_TECHNOLOGY,
			payload:        &conflictTechPayload,
			eventPublished: false,
		},
		{
			testName:       "pq on mesh is off tech nlx",
			pq:             true,
			meshnet:        false,
			tech:           config.Technology_NORDLYNX,
			payload:        &successPayload,
			eventPublished: true,
		},
		{
			testName:       "pq on mesh is on",
			pq:             true,
			meshnet:        true,
			payload:        &conflictMeshPayload,
			eventPublished: false,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockConfigManager.c.Mesh = test.meshnet
			mockConfigManager.c.Technology = test.tech
			mockConfigManager.c.AutoConnectData.PostquantumVpn = !test.pq

			req := pb.SetGenericRequest{Enabled: test.pq}
			resp, err := r.SetPostQuantum(context.Background(), &req)

			assert.NoError(t, err)
			assert.Equal(t, test.payload, resp)
			assert.Equal(t, test.eventPublished, mockPublisherSubscriber.EventPublished)
			mockPublisherSubscriber.EventPublished = false
		})
	}
}
