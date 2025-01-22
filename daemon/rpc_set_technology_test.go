package daemon

import (
	"context"
	"fmt"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/events"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
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
