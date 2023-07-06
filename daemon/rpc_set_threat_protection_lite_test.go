package daemon

import (
	"context"
	"fmt"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/stretchr/testify/assert"
)

func TestSetThreatProtectionLite_Success(t *testing.T) {
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
			expectedDNS:    tplNameserversV4,
			expectedStatus: pb.SetThreatProtectionLiteStatus_TPL_CONFIGURED,
		},
		{
			testName:       "set tpl ipv6",
			ipv6:           true,
			desiredTpl:     true,
			expectedDNS:    append(tplNameserversV4, tplNameserversV6...),
			expectedStatus: pb.SetThreatProtectionLiteStatus_TPL_CONFIGURED,
		},
		{
			testName:       "set tpl reset dns ipv4",
			desiredTpl:     true,
			currentDNS:     dns,
			expectedDNS:    tplNameserversV4,
			expectedStatus: pb.SetThreatProtectionLiteStatus_TPL_CONFIGURED_DNS_RESET,
		},
		{
			testName:       "set tpl reset dns ipv6",
			ipv6:           true,
			desiredTpl:     true,
			currentDNS:     dns,
			expectedDNS:    append(tplNameserversV4, tplNameserversV6...),
			expectedStatus: pb.SetThreatProtectionLiteStatus_TPL_CONFIGURED_DNS_RESET,
		},
		{
			testName:       "set tpl off ipv4",
			desiredTpl:     false,
			currentTpl:     true,
			currentDNS:     tplNameserversV4,
			expectedDNS:    defaultNameserversV4,
			expectedStatus: pb.SetThreatProtectionLiteStatus_TPL_CONFIGURED,
		},
		{
			testName:       "set tpl on ipv6",
			ipv6:           true,
			desiredTpl:     false,
			currentTpl:     true,
			currentDNS:     tplNameserversV4,
			expectedDNS:    append(defaultNameserversV4, defaultNameserversV6...),
			expectedStatus: pb.SetThreatProtectionLiteStatus_TPL_CONFIGURED,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			configManager := mockConfigManagerCommon{
				threatProtectionLite: test.currentTpl,
				dns:                  test.currentDNS,
				ipv6:                 test.ipv6,
			}
			networker := mockNetworker{}
			dnsGetter := mockDNSGetter{}
			tplPublisher := &mockPublisherSubcriber{}
			publisher := SettingsEvents{ThreatProtectionLite: tplPublisher}

			rpc := RPC{
				cm:          &configManager,
				netw:        &networker,
				nameservers: &dnsGetter,
				events:      &Events{Settings: &publisher},
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
			assert.Equal(t, test.expectedDNS, networker.dns, "Invalid nameservers were configured.")
			assert.Equal(t, test.desiredTpl, configManager.threatProtectionLite,
				"Threat protection lite was not saved in the config.")
			assert.Equal(t, true, tplPublisher.eventPublished, "TPL set event was not published.")
		})
	}
}

func TestSetThreatProtectionLite_Error(t *testing.T) {
	tests := []struct {
		testName         string
		desiredTpl       bool
		currentTpl       bool
		setDnsErr        error
		saveConfigErr    error
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
			testName:      "save config error",
			desiredTpl:    true,
			saveConfigErr: fmt.Errorf("Failed to save config"),
			expectedResponse: &pb.SetThreatProtectionLiteResponse{
				Response: &pb.SetThreatProtectionLiteResponse_ErrorCode{ErrorCode: pb.SetErrorCode_CONFIG_ERROR},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			configManager := mockConfigManagerCommon{
				threatProtectionLite: test.currentTpl,
				dns:                  defaultNameserversV4,
				saveConfigErr:        test.saveConfigErr,
			}
			networker := mockNetworker{
				setDNSErr: test.setDnsErr,
			}
			dnsGetter := mockDNSGetter{}
			tplPublisher := &mockPublisherSubcriber{}
			publisher := SettingsEvents{ThreatProtectionLite: tplPublisher}

			rpc := RPC{
				cm:          &configManager,
				netw:        &networker,
				nameservers: &dnsGetter,
				events:      &Events{Settings: &publisher},
			}

			resp, err := rpc.SetThreatProtectionLite(context.Background(),
				&pb.SetThreatProtectionLiteRequest{ThreatProtectionLite: test.desiredTpl})

			assert.Nil(t, err, "RPC ended with error.")
			assert.Equal(t, resp, test.expectedResponse, resp, "Invalid RPC response.")
		})
	}
}
