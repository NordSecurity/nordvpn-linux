package daemon

import (
	"context"
	"fmt"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockProtocolPublisherSubcriber struct {
	eventPublished bool
}

func (mp *mockProtocolPublisherSubcriber) Publish(message config.Protocol) {
	mp.eventPublished = true
}
func (*mockProtocolPublisherSubcriber) Subscribe(handler events.Handler[config.Protocol]) {}

func TestSetProtocol_Success(t *testing.T) {
	tests := []struct {
		name            string
		vpnActive       bool
		currentProtocol config.Protocol
		desiredProtocol config.Protocol
		expectedStatus  pb.SetProtocolStatus
	}{
		{
			name:            "set protocol tcp success",
			currentProtocol: config.Protocol_UDP,
			desiredProtocol: config.Protocol_TCP,
			expectedStatus:  pb.SetProtocolStatus_PROTOCOL_CONFIGURED,
		},
		{
			name:            "set protocol udp success",
			currentProtocol: config.Protocol_TCP,
			desiredProtocol: config.Protocol_UDP,
			expectedStatus:  pb.SetProtocolStatus_PROTOCOL_CONFIGURED,
		},
		{
			name:            "set protocol tcp success vpn on",
			vpnActive:       true,
			currentProtocol: config.Protocol_UDP,
			desiredProtocol: config.Protocol_TCP,
			expectedStatus:  pb.SetProtocolStatus_PROTOCOL_CONFIGURED_VPN_ON,
		},
		{
			name:            "set protocol udp success vpn on",
			vpnActive:       true,
			currentProtocol: config.Protocol_TCP,
			desiredProtocol: config.Protocol_UDP,
			expectedStatus:  pb.SetProtocolStatus_PROTOCOL_CONFIGURED_VPN_ON,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			uuid, _ := uuid.NewUUID()
			filesystem := newFilesystemMock(t)
			configManager := config.NewFilesystemConfigManager(
				"/location", "/vault", "",
				&machineIDGetterMock{machineID: uuid},
				&filesystem)

			configManager.SaveWith(func(c config.Config) config.Config {
				c.AutoConnectData.Protocol = test.currentProtocol
				c.Technology = config.Technology_OPENVPN
				return c
			})

			protocolPublisher := &mockProtocolPublisherSubcriber{}
			publisher := SettingsEvents{Protocol: protocolPublisher}
			networker := mockNetworker{
				vpnActive: test.vpnActive,
			}

			rpc := RPC{
				cm:     configManager,
				events: &Events{Settings: &publisher},
				netw:   &networker,
			}

			resp, err := rpc.SetProtocol(context.Background(), &pb.SetProtocolRequest{
				Protocol: test.desiredProtocol,
			})

			assert.Nil(t, err, "RPC ended with error.")
			assert.IsType(t, resp.Response, &pb.SetProtocolResponse_SetProtocolStatus{})
			assert.Equal(t, test.expectedStatus, resp.GetSetProtocolStatus(),
				"Invalid status received in RPC response.")

			var cfg config.Config
			configManager.Load(&cfg)

			assert.Equal(t, test.desiredProtocol, cfg.AutoConnectData.Protocol,
				"Invalid status saved in configuration.")
			assert.Equal(t, true, protocolPublisher.eventPublished,
				"Protocol event was not published after success.")
		})
	}
}

func TestSetProtocol_Error(t *testing.T) {
	tests := []struct {
		name              string
		currentTechnology config.Technology
		currentProtocol   config.Protocol
		writeConfigErr    error
		desiredProtocol   config.Protocol
		expectedResponse  *pb.SetProtocolResponse
	}{
		{
			name:              "set protocol already set",
			currentTechnology: config.Technology_OPENVPN,
			currentProtocol:   config.Protocol_TCP,
			desiredProtocol:   config.Protocol_TCP,
			expectedResponse: &pb.SetProtocolResponse{
				Response: &pb.SetProtocolResponse_ErrorCode{
					ErrorCode: pb.SetErrorCode_ALREADY_SET,
				},
			},
		},
		{
			name:              "set protocol invalid technology",
			currentTechnology: config.Technology_NORDLYNX,
			currentProtocol:   config.Protocol_UDP,
			desiredProtocol:   config.Protocol_TCP,
			expectedResponse: &pb.SetProtocolResponse{
				Response: &pb.SetProtocolResponse_SetProtocolStatus{
					SetProtocolStatus: pb.SetProtocolStatus_INVALID_TECHNOLOGY,
				},
			},
		},
		{
			name:              "set protocol config error",
			currentTechnology: config.Technology_OPENVPN,
			currentProtocol:   config.Protocol_UDP,
			writeConfigErr:    fmt.Errorf("Failed to save config"),
			desiredProtocol:   config.Protocol_TCP,
			expectedResponse: &pb.SetProtocolResponse{
				Response: &pb.SetProtocolResponse_ErrorCode{
					ErrorCode: pb.SetErrorCode_CONFIG_ERROR,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			uuid, _ := uuid.NewUUID()
			filesystem := newFilesystemMock(t)
			configManager := config.NewFilesystemConfigManager(
				"/location", "/vault", "",
				&machineIDGetterMock{machineID: uuid},
				&filesystem)

			configManager.SaveWith(func(c config.Config) config.Config {
				c.AutoConnectData.Protocol = test.currentProtocol
				c.Technology = test.currentTechnology
				return c
			})

			filesystem.WriteErr = test.writeConfigErr

			protocolPublisher := &mockProtocolPublisherSubcriber{}
			publisher := SettingsEvents{Protocol: protocolPublisher}
			networker := mockNetworker{}

			rpc := RPC{
				cm:     configManager,
				events: &Events{Settings: &publisher},
				netw:   &networker,
			}

			resp, err := rpc.SetProtocol(context.Background(), &pb.SetProtocolRequest{
				Protocol: test.desiredProtocol,
			})

			assert.Nil(t, err, "RPC ended with error.")
			assert.Equal(t, test.expectedResponse, resp,
				"Invalid RPC response.")

			var cfg config.Config
			configManager.Load(&cfg)

			assert.Equal(t, test.currentProtocol, cfg.AutoConnectData.Protocol,
				"Invalid status saved in configuration.")
			assert.Equal(t, false, protocolPublisher.eventPublished,
				"Protocol event was published after failure.")
		})
	}
}
