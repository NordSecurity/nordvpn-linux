package daemon

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/events"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	core_test "github.com/NordSecurity/nordvpn-linux/test/mock/core"
	"github.com/NordSecurity/nordvpn-linux/test/mock/networker"
	"github.com/stretchr/testify/assert"
)

func TestSetTechnology_NordWhisper(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name                  string
		nordwhisperEnabled    bool
		nordWhisperEnabledErr error
		currentTech           config.Technology
		currentProtocol       config.Protocol
		targetTech            config.Technology
		expectedTech          config.Technology
		expectedProtocol      config.Protocol
		expectedResponseType  int64
	}{
		{
			name:                 "NordWhisper disabled",
			currentTech:          config.Technology_NORDLYNX,
			currentProtocol:      config.Protocol_UDP,
			targetTech:           config.Technology_NORDWHISPER,
			expectedTech:         config.Technology_NORDLYNX,
			expectedProtocol:     config.Protocol_UDP,
			expectedResponseType: internal.CodeFeatureHidden,
		},
		{
			name:                  "failed to get NordWhisper status",
			nordWhisperEnabledErr: fmt.Errorf("failed to get NordWhisper status"),
			currentTech:           config.Technology_NORDLYNX,
			currentProtocol:       config.Protocol_TCP,
			targetTech:            config.Technology_NORDWHISPER,
			expectedTech:          config.Technology_NORDLYNX,
			expectedProtocol:      config.Protocol_TCP,
			expectedResponseType:  internal.CodeFeatureHidden,
		},
		{
			name:                 "switch from NordWhisper to OpenVPN",
			nordwhisperEnabled:   true,
			currentTech:          config.Technology_NORDWHISPER,
			currentProtocol:      config.Protocol_Webtunnel,
			targetTech:           config.Technology_OPENVPN,
			expectedTech:         config.Technology_OPENVPN,
			expectedProtocol:     config.Protocol_UDP,
			expectedResponseType: internal.CodeSuccess,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			remoteConfigGetter := mock.NewRemoteConfigMock()
			remoteConfigGetter.GetNordWhisperErr = test.nordWhisperEnabledErr

			configManager := mock.NewMockConfigManager()
			configManager.Cfg = &config.Config{
				Technology: test.currentTech,
				AutoConnectData: config.AutoConnectData{
					Protocol: test.currentProtocol,
				},
			}

			networker := networker.Mock{}

			r := RPC{
				remoteConfigGetter: remoteConfigGetter,
				cm:                 configManager,
				netw:               &networker,
				factory:            func(t config.Technology) (vpn.VPN, error) { return nil, nil },
				events:             events.NewEventsEmpty(),
			}

			resp, err := r.SetTechnology(context.Background(),
				&pb.SetTechnologyRequest{Technology: test.targetTech})
			assert.Nil(t, err, "Unexpected error returned by IsNordWhisperEnabled rpc.")
			assert.Equal(t, test.expectedResponseType, resp.Type, "Unexpected response type received.")
			assert.Equal(t, test.expectedTech, configManager.Cfg.Technology, "Unexpected technology saved in config.")
			assert.Equal(t, test.expectedProtocol, configManager.Cfg.AutoConnectData.Protocol,
				"Unexpected protocol saved in config.")
		})
	}
}

func TestSetTechnology_ECHReset(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name        string
		currentTech config.Technology
		targetTech  config.Technology
		storedECH   bool
		expectedECH bool
	}{
		{
			name:        "leaving NordWhisper for NordLynx resets ECH to default",
			currentTech: config.Technology_NORDWHISPER,
			targetTech:  config.Technology_NORDLYNX,
			storedECH:   false,
			expectedECH: true,
		},
		{
			name:        "leaving NordWhisper for OpenVPN resets ECH to default",
			currentTech: config.Technology_NORDWHISPER,
			targetTech:  config.Technology_OPENVPN,
			storedECH:   false,
			expectedECH: true,
		},
		{
			name:        "already-default ECH stays enabled when leaving NordWhisper",
			currentTech: config.Technology_NORDWHISPER,
			targetTech:  config.Technology_NORDLYNX,
			storedECH:   true,
			expectedECH: true,
		},
		{
			name:        "switching between non-NordWhisper technologies also forces default",
			currentTech: config.Technology_NORDLYNX,
			targetTech:  config.Technology_OPENVPN,
			storedECH:   false,
			expectedECH: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			remoteConfigGetter := mock.NewRemoteConfigMock()
			remoteConfigGetter.NordWhisperEnabled = true

			configManager := mock.NewMockConfigManager()
			cfg := config.Config{
				Technology: test.currentTech,
				AutoConnectData: config.AutoConnectData{
					Protocol: config.Protocol_UDP,
				},
			}
			cfg.AutoConnectData.ECH.Set(test.storedECH)
			configManager.Cfg = &cfg

			networker := networker.Mock{}

			r := RPC{
				remoteConfigGetter: remoteConfigGetter,
				cm:                 configManager,
				netw:               &networker,
				factory:            func(t config.Technology) (vpn.VPN, error) { return nil, nil },
				events:             events.NewEventsEmpty(),
			}

			resp, err := r.SetTechnology(context.Background(),
				&pb.SetTechnologyRequest{Technology: test.targetTech})
			assert.Nil(t, err, "Unexpected error returned by SetTechnology rpc.")
			assert.Equal(t, internal.CodeSuccess, resp.Type, "Expected the switch to succeed.")
			assert.Equal(t, test.expectedECH, configManager.Cfg.AutoConnectData.ECH.Get(),
				"Unexpected ECH value saved in config.")
		})
	}
}

func TestSetTechnology_DedicatedServer(t *testing.T) {
	category.Set(t, category.Unit)

	configManager := mock.NewMockConfigManager()
	configManager.Cfg = &config.Config{
		Technology:  config.Technology_NORDLYNX,
		AutoConnect: true,
		AutoConnectData: config.AutoConnectData{
			Group: config.ServerGroup_DEDICATED_SERVER,
		},
	}

	networker := networker.Mock{}

	remoteConfigGetter := mock.NewRemoteConfigMock()
	remoteConfigGetter.NordWhisperEnabled = true

	r := RPC{
		remoteConfigGetter: remoteConfigGetter,
		cm:                 configManager,
		netw:               &networker,
		factory:            func(t config.Technology) (vpn.VPN, error) { return nil, nil },
		events:             events.NewEventsEmpty(),
	}

	resp, _ := r.SetTechnology(context.Background(), &pb.SetTechnologyRequest{
		Technology: config.Technology_OPENVPN,
	})

	assert.Equal(t, internal.CodeDedicatedServersNoNordlynx, resp.Type, "Invalid response code by the daemon.")
	assert.Equal(t, config.Technology_NORDLYNX, configManager.Cfg.Technology,
		"Technology should not be changed when daemon responds with a non-success code.")
}

func TestSetTechnologyCheckAutoConnectData(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name                  string
		isNordWhisperDisabled bool
		cfg                   config.Config
		toTechnology          config.Technology
		expectedResponse      *pb.Payload
	}{
		{
			name: "set technology works when autoconnect is off",
			cfg: config.Config{
				Technology:  config.Technology_NORDLYNX,
				AutoConnect: false,
				AutoConnectData: config.AutoConnectData{
					Group: config.ServerGroup_DOUBLE_VPN,
				},
			},
			toTechnology: config.Technology_NORDWHISPER,
			expectedResponse: &pb.Payload{
				Type: internal.CodeSuccess,
				Data: []string{"false", config.TechNameToUpperCamelCase(config.Technology_NORDWHISPER)},
			},
		},
		{
			name: "works when autoconnect is on and autoconnect has compatible settings",
			cfg: config.Config{
				Technology:  config.Technology_NORDLYNX,
				AutoConnect: true,
				AutoConnectData: config.AutoConnectData{
					Group: config.ServerGroup_DOUBLE_VPN,
				},
			},
			toTechnology: config.Technology_OPENVPN,
			expectedResponse: &pb.Payload{
				Type: internal.CodeSuccess,
				Data: []string{"false", config.TechNameToUpperCamelCase(config.Technology_OPENVPN)},
			},
		},
		{
			name: "fails when autoconnect is on and autoconnect is with incompatible settings",
			cfg: config.Config{
				Technology:  config.Technology_NORDLYNX,
				AutoConnect: true,
				AutoConnectData: config.AutoConnectData{
					Group: config.ServerGroup_DOUBLE_VPN,
				},
			},
			toTechnology:          config.Technology_NORDWHISPER,
			isNordWhisperDisabled: false,
			expectedResponse: &pb.Payload{
				Type: internal.CodeTechnologyIncompatibleWithAutoconnect,
			},
		},
		{
			name: "fails always for NordWhisper when feature is disabled",
			cfg: config.Config{
				Technology:  config.Technology_NORDLYNX,
				AutoConnect: true,
				AutoConnectData: config.AutoConnectData{
					Group: config.ServerGroup_DOUBLE_VPN,
				},
			},
			toTechnology:          config.Technology_NORDWHISPER,
			isNordWhisperDisabled: true,
			expectedResponse: &pb.Payload{
				Type: internal.CodeFeatureHidden,
			},
		},
	}

	networker := networker.Mock{}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			configManager := mock.NewMockConfigManager()
			configManager.Cfg = &test.cfg

			remoteConfigGetter := mock.NewRemoteConfigMock()
			remoteConfigGetter.NordWhisperEnabled = !test.isNordWhisperDisabled
			dm := DataManager{serversData: ServersData{Servers: core_test.ServersList()}}

			r := RPC{
				remoteConfigGetter: remoteConfigGetter,
				cm:                 configManager,
				netw:               &networker,
				factory:            func(t config.Technology) (vpn.VPN, error) { return nil, nil },
				events:             events.NewEventsEmpty(),
				dm:                 &dm,
				serversAPI:         core_test.NewMockFailingServersAPI(errors.New("500")),
			}

			resp, err := r.setTechnologyImpl(&pb.SetTechnologyRequest{
				Technology: test.toTechnology,
			}, !test.isNordWhisperDisabled)
			assert.NoError(t, err)
			assert.Equal(t, test.expectedResponse, resp)
			if resp.Type == internal.CodeSuccess {
				assert.Equal(t, test.toTechnology, configManager.Cfg.Technology)
			} else {
				assert.Equal(t, test.cfg.Technology, configManager.Cfg.Technology)
			}
		})
	}
}
