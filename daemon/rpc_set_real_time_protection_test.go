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
	"github.com/NordSecurity/nordvpn-linux/test/mock/fs"
	"github.com/NordSecurity/nordvpn-linux/test/mock/networker"
)

func TestSetRealTimeProtection_Success(t *testing.T) {
	category.Set(t, category.Unit)

	dns := []string{"0.0.0.0", "8.8.8.8", "1.1.1.1"}

	tests := []struct {
		testName       string
		desiredRTP     bool
		currentRTP     bool
		currentDNS     []string
		expectedDNS    []string
		expectedStatus pb.SetRealTimeProtectionStatus
	}{
		{
			testName:       "set rtp ipv4",
			desiredRTP:     true,
			expectedDNS:    mock.RealTimeProtectionNameserversV4,
			expectedStatus: pb.SetRealTimeProtectionStatus_RTP_CONFIGURED,
		},
		{
			testName:       "set rtp reset dns ipv4",
			desiredRTP:     true,
			currentDNS:     dns,
			expectedDNS:    mock.RealTimeProtectionNameserversV4,
			expectedStatus: pb.SetRealTimeProtectionStatus_RTP_CONFIGURED_DNS_RESET,
		},
		{
			testName:       "set rtp off ipv4",
			desiredRTP:     false,
			currentRTP:     true,
			currentDNS:     mock.RealTimeProtectionNameserversV4,
			expectedDNS:    mock.DefaultNameserversV4,
			expectedStatus: pb.SetRealTimeProtectionStatus_RTP_CONFIGURED,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			uuid, _ := uuid.NewUUID()
			filesystem := fs.NewSystemFileHandleMock(t)
			configManager := config.NewFilesystemConfigManager(
				"/location", "/vault", "",
				&machineIDGetterMock{machineID: uuid},
				&filesystem,
				nil)

			configManager.SaveWith(func(c config.Config) config.Config {
				c.AutoConnectData = config.AutoConnectData{
					RealTimeProtection: test.currentRTP,
					DNS:                test.currentDNS,
				}

				return c
			})

			networker := networker.Mock{}
			dnsGetter := mock.DNSGetter{}
			protectionPublisher := &events.MockPublisherSubscriber[bool]{}
			publisher := events.SettingsEvents{RealTimeProtection: protectionPublisher}

			rpc := RPC{
				cm:          configManager,
				netw:        &networker,
				nameservers: &dnsGetter,
				events:      &events.Events{Settings: &publisher},
			}

			resp, err := rpc.SetRealTimeProtection(context.Background(),
				&pb.SetRealTimeProtectionRequest{RealTimeProtection: test.desiredRTP})

			assert.Nil(t, err, "RPC ended with error.")
			assert.IsType(t,
				resp.Response,
				&pb.SetRealTimeProtectionResponse_SetRealTimeProtectionStatus{},
				"RPC response is of invalid type.")
			assert.Equal(t,
				resp.GetSetRealTimeProtectionStatus(),
				test.expectedStatus,
				"Invalid response from RPC.")
			assert.Equal(t, test.expectedDNS, networker.Dns, "Invalid nameservers were configured.")

			var config config.Config
			configManager.Load(&config)

			assert.Equal(t, test.desiredRTP, config.AutoConnectData.RealTimeProtection,
				"Real time protection was not saved in the config.")
			assert.Equal(t, true, protectionPublisher.EventPublished, "protection set event was not published.")
		})
	}
}

func TestSetRealTimeProtection_Error(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		testName         string
		desiredRTP       bool
		currentRTP       bool
		setDnsErr        error
		writeConfigErr   error
		expectedResponse *pb.SetRealTimeProtectionResponse
	}{
		{
			testName:   "already set on",
			desiredRTP: true,
			currentRTP: true,
			expectedResponse: &pb.SetRealTimeProtectionResponse{
				Response: &pb.SetRealTimeProtectionResponse_ErrorCode{ErrorCode: pb.SetErrorCode_ALREADY_SET},
			},
		},
		{
			testName:   "already set off",
			desiredRTP: false,
			currentRTP: false,
			expectedResponse: &pb.SetRealTimeProtectionResponse{
				Response: &pb.SetRealTimeProtectionResponse_ErrorCode{ErrorCode: pb.SetErrorCode_ALREADY_SET},
			},
		},
		{
			testName:   "set dns error",
			desiredRTP: true,
			setDnsErr:  fmt.Errorf("Failed to set dns."),
			expectedResponse: &pb.SetRealTimeProtectionResponse{
				Response: &pb.SetRealTimeProtectionResponse_ErrorCode{ErrorCode: pb.SetErrorCode_CONFIG_ERROR},
			},
		},
		{
			testName:       "save config error",
			desiredRTP:     true,
			writeConfigErr: fmt.Errorf("Failed to save config"),
			expectedResponse: &pb.SetRealTimeProtectionResponse{
				Response: &pb.SetRealTimeProtectionResponse_ErrorCode{ErrorCode: pb.SetErrorCode_CONFIG_ERROR},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			uuid, _ := uuid.NewUUID()
			filesystem := fs.NewSystemFileHandleMock(t)
			filesystem.WriteErr = test.writeConfigErr
			configManager := config.NewFilesystemConfigManager(
				"/location", "/vault", "",
				&machineIDGetterMock{machineID: uuid},
				&filesystem,
				nil)

			configManager.SaveWith(func(c config.Config) config.Config {
				c.AutoConnectData = config.AutoConnectData{
					RealTimeProtection: test.currentRTP,
					DNS:                mock.DefaultNameserversV4,
				}

				return c
			})

			networker := networker.Mock{
				SetDNSErr: test.setDnsErr,
			}
			dnsGetter := mock.DNSGetter{}
			rtpPublisher := &events.MockPublisherSubscriber[bool]{}
			publisher := events.SettingsEvents{RealTimeProtection: rtpPublisher}

			rpc := RPC{
				cm:          configManager,
				netw:        &networker,
				nameservers: &dnsGetter,
				events:      &events.Events{Settings: &publisher},
			}

			resp, err := rpc.SetRealTimeProtection(context.Background(),
				&pb.SetRealTimeProtectionRequest{RealTimeProtection: test.desiredRTP})

			assert.Nil(t, err, "RPC ended with error.")
			assert.Equal(t, resp, test.expectedResponse, resp, "Invalid RPC response.")
		})
	}
}
