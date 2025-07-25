package daemon

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/events"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	configMock "github.com/NordSecurity/nordvpn-linux/test/mock/config"
	"github.com/NordSecurity/nordvpn-linux/test/mock/networker"
)

func TestSetThreatProtectionLite_Success(t *testing.T) {
	category.Set(t, category.Unit)

	dns := []string{"0.0.0.0", "8.8.8.8", "1.1.1.1"}

	tests := []struct {
		testName       string
		ipv6           bool
		desiredTpl     bool
		currentTpl     bool
		currentDNS     []string
		expectedDNS    []string
		expectedStatus pb.SetThreatProtectionLiteStatus
	}{
		{
			testName:       "set tpl ipv4",
			desiredTpl:     true,
			expectedDNS:    mock.TplNameserversV4,
			expectedStatus: pb.SetThreatProtectionLiteStatus_TPL_CONFIGURED,
		},
		{
			testName:       "set tpl ipv6",
			ipv6:           true,
			desiredTpl:     true,
			expectedDNS:    append(mock.TplNameserversV4, mock.TplNameserversV6...),
			expectedStatus: pb.SetThreatProtectionLiteStatus_TPL_CONFIGURED,
		},
		{
			testName:       "set tpl reset dns ipv4",
			desiredTpl:     true,
			currentDNS:     dns,
			expectedDNS:    mock.TplNameserversV4,
			expectedStatus: pb.SetThreatProtectionLiteStatus_TPL_CONFIGURED_DNS_RESET,
		},
		{
			testName:       "set tpl reset dns ipv6",
			ipv6:           true,
			desiredTpl:     true,
			currentDNS:     dns,
			expectedDNS:    append(mock.TplNameserversV4, mock.TplNameserversV6...),
			expectedStatus: pb.SetThreatProtectionLiteStatus_TPL_CONFIGURED_DNS_RESET,
		},
		{
			testName:       "set tpl off ipv4",
			desiredTpl:     false,
			currentTpl:     true,
			currentDNS:     mock.TplNameserversV4,
			expectedDNS:    mock.DefaultNameserversV4,
			expectedStatus: pb.SetThreatProtectionLiteStatus_TPL_CONFIGURED,
		},
		{
			testName:       "set tpl on ipv6",
			ipv6:           true,
			desiredTpl:     false,
			currentTpl:     true,
			currentDNS:     mock.TplNameserversV4,
			expectedDNS:    append(mock.DefaultNameserversV4, mock.DefaultNameserversV6...),
			expectedStatus: pb.SetThreatProtectionLiteStatus_TPL_CONFIGURED,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			uuid, _ := uuid.NewUUID()
			filesystem := configMock.NewFilesystemMock(t)
			configManager := config.NewFilesystemConfigManager(
				"/location", "/vault", "",
				&machineIDGetterMock{machineID: uuid},
				&filesystem,
				nil)

			configManager.SaveWith(func(c config.Config) config.Config {
				c.AutoConnectData = config.AutoConnectData{
					ThreatProtectionLite: test.currentTpl,
					DNS:                  test.currentDNS,
				}
				c.IPv6 = test.ipv6

				return c
			})

			networker := networker.Mock{}
			dnsGetter := mock.DNSGetter{}
			tplPublisher := &events.MockPublisherSubscriber[bool]{}
			publisher := events.SettingsEvents{ThreatProtectionLite: tplPublisher}

			rpc := RPC{
				cm:          configManager,
				netw:        &networker,
				nameservers: &dnsGetter,
				events:      &events.Events{Settings: &publisher},
			}

			resp, err := rpc.SetThreatProtectionLite(context.Background(),
				&pb.SetThreatProtectionLiteRequest{ThreatProtectionLite: test.desiredTpl})

			assert.Nil(t, err, "RPC ended with error.")
			assert.IsType(t,
				resp.Response,
				&pb.SetThreatProtectionLiteResponse_SetThreatProtectionLiteStatus{},
				"RPC response is of invalid type.")
			assert.Equal(t,
				resp.GetSetThreatProtectionLiteStatus(),
				test.expectedStatus,
				"Invalid response from RPC.")
			assert.Equal(t, test.expectedDNS, networker.Dns, "Invalid nameservers were configured.")

			var config config.Config
			configManager.Load(&config)

			assert.Equal(t, test.desiredTpl, config.AutoConnectData.ThreatProtectionLite,
				"Threat protection lite was not saved in the config.")
			assert.Equal(t, true, tplPublisher.EventPublished, "TPL set event was not published.")
		})
	}
}

func TestSetThreatProtectionLite_Error(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		testName         string
		desiredTpl       bool
		currentTpl       bool
		setDnsErr        error
		writeConfigErr   error
		expectedResponse *pb.SetThreatProtectionLiteResponse
	}{
		{
			testName:   "already set on",
			desiredTpl: true,
			currentTpl: true,
			expectedResponse: &pb.SetThreatProtectionLiteResponse{
				Response: &pb.SetThreatProtectionLiteResponse_ErrorCode{ErrorCode: pb.SetErrorCode_ALREADY_SET},
			},
		},
		{
			testName:   "already set off",
			desiredTpl: false,
			currentTpl: false,
			expectedResponse: &pb.SetThreatProtectionLiteResponse{
				Response: &pb.SetThreatProtectionLiteResponse_ErrorCode{ErrorCode: pb.SetErrorCode_ALREADY_SET},
			},
		},
		{
			testName:   "set dns error",
			desiredTpl: true,
			setDnsErr:  fmt.Errorf("Failed to set dns."),
			expectedResponse: &pb.SetThreatProtectionLiteResponse{
				Response: &pb.SetThreatProtectionLiteResponse_ErrorCode{ErrorCode: pb.SetErrorCode_CONFIG_ERROR},
			},
		},
		{
			testName:       "save config error",
			desiredTpl:     true,
			writeConfigErr: fmt.Errorf("Failed to save config"),
			expectedResponse: &pb.SetThreatProtectionLiteResponse{
				Response: &pb.SetThreatProtectionLiteResponse_ErrorCode{ErrorCode: pb.SetErrorCode_CONFIG_ERROR},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			uuid, _ := uuid.NewUUID()
			filesystem := configMock.NewFilesystemMock(t)
			filesystem.WriteErr = test.writeConfigErr
			configManager := config.NewFilesystemConfigManager(
				"/location", "/vault", "",
				&machineIDGetterMock{machineID: uuid},
				&filesystem,
				nil)

			configManager.SaveWith(func(c config.Config) config.Config {
				c.AutoConnectData = config.AutoConnectData{
					ThreatProtectionLite: test.currentTpl,
					DNS:                  mock.DefaultNameserversV4,
				}

				return c
			})

			networker := networker.Mock{
				SetDNSErr: test.setDnsErr,
			}
			dnsGetter := mock.DNSGetter{}
			tplPublisher := &events.MockPublisherSubscriber[bool]{}
			publisher := events.SettingsEvents{ThreatProtectionLite: tplPublisher}

			rpc := RPC{
				cm:          configManager,
				netw:        &networker,
				nameservers: &dnsGetter,
				events:      &events.Events{Settings: &publisher},
			}

			resp, err := rpc.SetThreatProtectionLite(context.Background(),
				&pb.SetThreatProtectionLiteRequest{ThreatProtectionLite: test.desiredTpl})

			assert.Nil(t, err, "RPC ended with error.")
			assert.Equal(t, resp, test.expectedResponse, resp, "Invalid RPC response.")
		})
	}
}
