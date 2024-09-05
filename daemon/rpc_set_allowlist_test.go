package daemon

import (
	"context"
	"fmt"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/events"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"github.com/NordSecurity/nordvpn-linux/test/mock/networker"
	"github.com/stretchr/testify/assert"
)

func newMockedRPC(allowlist config.Allowlist, configLoadErr error) (*mock.ConfigManager, *RPC) {
	config := config.Config{
		AutoConnectData: config.AutoConnectData{
			Allowlist: allowlist,
		},
	}
	configManager := mock.NewMockConfigManager()
	configManager.Cfg = &config
	configManager.LoadErr = configLoadErr

	networker := networker.Mock{}
	ev := events.NewEventsEmpty()

	r := RPC{
		cm:     configManager,
		netw:   &networker,
		events: ev,
	}

	return configManager, &r
}

func TestSetAllowlist_Subnet(t *testing.T) {
	category.Set(t, category.Unit)

	subnet1 := "1.1.1.1/24"
	subnet2 := "156.37.220.88/22"
	subnet3 := "116.83.95.53/8"
	invalidSubnet := "354.333.95/7"

	tests := []struct {
		name               string
		subnet             string
		currentAllowlist   config.Allowlist
		configLoadErr      error
		expectedAllowlist  config.Allowlist
		expectedReturnCode int64
	}{
		{
			name:               "add subnet to empty allowlist success",
			subnet:             subnet1,
			currentAllowlist:   config.NewAllowlist([]int64{}, []int64{}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{}, []int64{}, []string{subnet1}),
			expectedReturnCode: internal.CodeSuccess,
		},
		{
			name:               "add subnet to non-empty allowlist success",
			subnet:             subnet1,
			currentAllowlist:   config.NewAllowlist([]int64{5001}, []int64{6001}, []string{subnet2, subnet3}),
			expectedAllowlist:  config.NewAllowlist([]int64{5001}, []int64{6001}, []string{subnet1, subnet2, subnet3}),
			expectedReturnCode: internal.CodeSuccess,
		},
		{
			name:               "add subnet allowlist failure invalid subnet",
			subnet:             invalidSubnet,
			currentAllowlist:   config.NewAllowlist([]int64{5001}, []int64{6001}, []string{subnet2, subnet3}),
			expectedAllowlist:  config.NewAllowlist([]int64{5001}, []int64{6001}, []string{subnet2, subnet3}),
			expectedReturnCode: internal.CodeAllowlistInvalidSubnet,
		},
		{
			name:               "add subnet allowlist failure subnet already added",
			subnet:             subnet2,
			currentAllowlist:   config.NewAllowlist([]int64{5001}, []int64{6001}, []string{subnet2, subnet3}),
			expectedAllowlist:  config.NewAllowlist([]int64{5001}, []int64{6001}, []string{subnet2, subnet3}),
			expectedReturnCode: internal.CodeAllowlistSubnetNoop,
		},
		{
			name:               "add subnet allowlist failure config load error",
			subnet:             subnet1,
			currentAllowlist:   config.NewAllowlist([]int64{5001}, []int64{6001}, []string{subnet2, subnet3}),
			configLoadErr:      fmt.Errorf("config load failure"),
			expectedAllowlist:  config.NewAllowlist([]int64{5001}, []int64{6001}, []string{subnet2, subnet3}),
			expectedReturnCode: internal.CodeConfigError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			configManager, r := newMockedRPC(test.currentAllowlist, test.configLoadErr)

			req := pb.SetAllowlistRequest{
				Request: &pb.SetAllowlistRequest_SetAllowlistSubnetRequest{
					SetAllowlistSubnetRequest: &pb.SetAllowlistSubnetRequest{
						Subnet: test.subnet,
					},
				},
			}

			response, _ := r.SetAllowlist(context.Background(), &req)
			assert.Equal(t,
				test.expectedAllowlist,
				configManager.Cfg.AutoConnectData.Allowlist,
				"Allowlist configured incorrectly.")
			assert.Equal(t,
				test.expectedReturnCode,
				response.Type,
				"Invalid return code after setting allowlist.")
		})
	}
}

func TestUnsetAllowlist_Subnet(t *testing.T) {
	category.Set(t, category.Unit)

	subnet1 := "1.1.1.1/24"
	subnet2 := "156.37.220.88/22"
	subnet3 := "116.83.95.53/8"
	invalidSubnet := "354.333.95/7"

	tests := []struct {
		name               string
		subnet             string
		currentAllowlist   config.Allowlist
		configLoadErr      error
		expectedAllowlist  config.Allowlist
		expectedReturnCode int64
	}{
		{
			name:               "remove subnet success",
			subnet:             subnet1,
			currentAllowlist:   config.NewAllowlist([]int64{}, []int64{}, []string{subnet1, subnet2, subnet3}),
			expectedAllowlist:  config.NewAllowlist([]int64{}, []int64{}, []string{subnet2, subnet3}),
			expectedReturnCode: internal.CodeSuccess,
		},
		{
			name:               "remove subnet failure invalid subnet",
			subnet:             invalidSubnet,
			currentAllowlist:   config.NewAllowlist([]int64{}, []int64{}, []string{subnet1, subnet2, subnet3}),
			expectedAllowlist:  config.NewAllowlist([]int64{}, []int64{}, []string{subnet1, subnet2, subnet3}),
			expectedReturnCode: internal.CodeAllowlistInvalidSubnet,
		},
		{
			name:               "remove subnet failure subnet not added",
			subnet:             subnet1,
			currentAllowlist:   config.NewAllowlist([]int64{}, []int64{}, []string{subnet2, subnet3}),
			expectedAllowlist:  config.NewAllowlist([]int64{}, []int64{}, []string{subnet2, subnet3}),
			expectedReturnCode: internal.CodeAllowlistSubnetNoop,
		},
		{
			name:               "remove subnet failure config load failure",
			subnet:             subnet1,
			currentAllowlist:   config.NewAllowlist([]int64{}, []int64{}, []string{subnet1, subnet2, subnet3}),
			configLoadErr:      fmt.Errorf("failed to load config"),
			expectedAllowlist:  config.NewAllowlist([]int64{}, []int64{}, []string{subnet1, subnet2, subnet3}),
			expectedReturnCode: internal.CodeConfigError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			configManager, r := newMockedRPC(test.currentAllowlist, test.configLoadErr)

			req := pb.SetAllowlistRequest{
				Request: &pb.SetAllowlistRequest_SetAllowlistSubnetRequest{
					SetAllowlistSubnetRequest: &pb.SetAllowlistSubnetRequest{
						Subnet: test.subnet,
					},
				},
			}

			response, _ := r.UnsetAllowlist(context.Background(), &req)
			assert.Equal(t,
				test.expectedAllowlist,
				configManager.Cfg.AutoConnectData.Allowlist,
				"Allowlist configured incorrectly.")
			assert.Equal(t,
				test.expectedReturnCode,
				response.Type,
				"Invalid return code after setting allowlist.")
		})
	}
}

func TestSetAllowlist_Ports(t *testing.T) {
	category.Set(t, category.Unit)

	port1 := int64(50)
	port2 := int64(70)
	invalidPort1 := int64(70000)
	invalidPort2 := int64(0)

	portRange := []int64{port1}
	for p := port1 + 1; p <= port2; p++ {
		portRange = append(portRange, p)
	}

	expectedPortRangeUDP := []int64{44, 2000}
	expectedPortRangeUDP = append(expectedPortRangeUDP, portRange...)

	expectedPortRangeTCP := []int64{22, 2500}
	expectedPortRangeTCP = append(expectedPortRangeTCP, portRange...)

	tests := []struct {
		name               string
		portStart          int64
		portStop           int64
		isUDP              bool
		isTCP              bool
		currentAllowlist   config.Allowlist
		configLoadErr      error
		expectedAllowlist  config.Allowlist
		expectedReturnCode int64
	}{
		{
			name:               "add UDP port success",
			portStart:          port1,
			portStop:           0,
			isUDP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, port1, 2000}, []int64{22, 2500}, []string{}),
			expectedReturnCode: internal.CodeSuccess,
		},
		{
			name:               "add TCP port success",
			portStart:          port1,
			portStop:           0,
			isTCP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, 2000}, []int64{22, port1, 2500}, []string{}),
			expectedReturnCode: internal.CodeSuccess,
		},
		{
			name:               "add TCP/UDP port success",
			portStart:          port1,
			portStop:           0,
			isUDP:              true,
			isTCP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, port1, 2000}, []int64{22, port1, 2500}, []string{}),
			expectedReturnCode: internal.CodeSuccess,
		},
		{
			name:               "add UDP port range success",
			portStart:          port1,
			portStop:           port2,
			isUDP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist(expectedPortRangeUDP, []int64{22, 2500}, []string{}),
			expectedReturnCode: internal.CodeSuccess,
		},
		{
			name:               "add TCP port range success",
			portStart:          port1,
			portStop:           port2,
			isTCP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, 2000}, expectedPortRangeTCP, []string{}),
			expectedReturnCode: internal.CodeSuccess,
		},
		{
			name:               "add TCP/UDP port range success",
			portStart:          port1,
			portStop:           port2,
			isUDP:              true,
			isTCP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist(expectedPortRangeUDP, expectedPortRangeTCP, []string{}),
			expectedReturnCode: internal.CodeSuccess,
		},
		// invalid UDP port
		{
			name:               "add UDP port failure invalid port(over range)",
			portStart:          invalidPort1,
			isUDP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedReturnCode: internal.CodeAllowlistPortOutOfRange,
		},
		{
			name:               "add UDP port failure invalid port(under range)",
			portStart:          invalidPort2,
			isUDP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedReturnCode: internal.CodeAllowlistPortOutOfRange,
		},
		// invalid TCP port
		{
			name:               "add TCP port failure invalid port(under range)",
			portStart:          invalidPort1,
			isTCP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedReturnCode: internal.CodeAllowlistPortOutOfRange,
		},
		{
			name:               "add TCP port failure invalid port(under range)",
			portStart:          invalidPort2,
			isTCP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedReturnCode: internal.CodeAllowlistPortOutOfRange,
		},
		// invalid TCP/UDP port
		{
			name:               "add TCP/UDP port failure invalid port(under range)",
			portStart:          invalidPort1,
			isTCP:              true,
			isUDP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedReturnCode: internal.CodeAllowlistPortOutOfRange,
		},
		{
			name:               "add TCP/UDP port failure invalid port(under range)",
			portStart:          invalidPort2,
			isTCP:              true,
			isUDP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedReturnCode: internal.CodeAllowlistPortOutOfRange,
		},
		// invalid UDP port range
		{
			name:               "add UDP port range failure invalid port(over range)",
			portStart:          invalidPort1,
			portStop:           port1,
			isUDP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedReturnCode: internal.CodeAllowlistPortOutOfRange,
		},
		{
			name:               "add UDP port range failure invalid port(under range)",
			portStart:          invalidPort2,
			portStop:           port1,
			isUDP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedReturnCode: internal.CodeAllowlistPortOutOfRange,
		},
		{
			name:               "add UDP port range failure invalid port(start port greater than stop port)",
			portStart:          port2,
			portStop:           port1,
			isUDP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedReturnCode: internal.CodeAllowlistPortOutOfRange,
		},
		// invalid TCP port range
		{
			name:               "add TCP port range failure invalid port(over range)",
			portStart:          invalidPort1,
			portStop:           port1,
			isTCP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedReturnCode: internal.CodeAllowlistPortOutOfRange,
		},
		{
			name:               "add TCP port range failure invalid port(under range)",
			portStart:          invalidPort2,
			portStop:           port1,
			isTCP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedReturnCode: internal.CodeAllowlistPortOutOfRange,
		},
		{
			name:               "add TCP port range failure invalid port(start port greater than stop port)",
			portStart:          port2,
			portStop:           port1,
			isTCP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedReturnCode: internal.CodeAllowlistPortOutOfRange,
		},
		// invalid TCP/UDP port range
		{
			name:               "add TCP/UDP port range failure invalid port(over range)",
			portStart:          invalidPort1,
			portStop:           port1,
			isTCP:              true,
			isUDP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedReturnCode: internal.CodeAllowlistPortOutOfRange,
		},
		{
			name:               "add TCP/UDP port range failure invalid port(under range)",
			portStart:          invalidPort2,
			portStop:           port1,
			isTCP:              true,
			isUDP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedReturnCode: internal.CodeAllowlistPortOutOfRange,
		},
		{
			name:               "add TCP/UDP port range failure invalid port(start port greater than stop port)",
			portStart:          port2,
			portStop:           port1,
			isTCP:              true,
			isUDP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedReturnCode: internal.CodeAllowlistPortOutOfRange,
		},
		// port already added
		{
			name:               "add UDP port failure port already added",
			portStart:          port1,
			isUDP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{port1}, []int64{}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{port1}, []int64{}, []string{}),
			expectedReturnCode: internal.CodeAllowlistPortNoop,
		},
		{
			name:               "add TCP port failure port already added",
			portStart:          port1,
			isTCP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{}, []int64{port1}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{}, []int64{port1}, []string{}),
			expectedReturnCode: internal.CodeAllowlistPortNoop,
		},
		{
			name:               "add TCP/UDP port failure port already added",
			portStart:          port1,
			isTCP:              true,
			isUDP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{port1}, []int64{port1}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{port1}, []int64{port1}, []string{}),
			expectedReturnCode: internal.CodeAllowlistPortNoop,
		},
		// config failure
		{
			name:               "add TCP/UDP port failure config failure",
			portStart:          port1,
			isTCP:              true,
			isUDP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{}, []int64{}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{}, []int64{}, []string{}),
			configLoadErr:      fmt.Errorf("config load error"),
			expectedReturnCode: internal.CodeConfigError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			configManager, r := newMockedRPC(test.currentAllowlist, test.configLoadErr)

			req := pb.SetAllowlistRequest{
				Request: &pb.SetAllowlistRequest_SetAllowlistPortsRequest{
					SetAllowlistPortsRequest: &pb.SetAllowlistPortsRequest{
						IsUdp: test.isUDP,
						IsTcp: test.isTCP,
						PortRange: &pb.PortRange{
							StartPort: test.portStart,
							EndPort:   test.portStop,
						},
					},
				},
			}

			response, _ := r.SetAllowlist(context.Background(), &req)
			assert.Equal(t,
				test.expectedAllowlist,
				configManager.Cfg.AutoConnectData.Allowlist,
				"Allowlist configured incorrectly.")
			assert.Equal(t,
				test.expectedReturnCode,
				response.Type,
				"Invalid return code after setting allowlist.")
		})
	}
}

func TestUnsetAllowlist_Ports(t *testing.T) {
	category.Set(t, category.Unit)

	port1 := int64(50)
	port2 := int64(70)
	invalidPort1 := int64(70000)
	invalidPort2 := int64(0)

	portRange := []int64{port1}
	for p := port1 + 1; p <= port2; p++ {
		portRange = append(portRange, p)
	}

	currentPortRangeUDP := []int64{44, 2000}
	currentPortRangeUDP = append(currentPortRangeUDP, portRange...)

	currentPortRangeTCP := []int64{22, 2500}
	currentPortRangeTCP = append(currentPortRangeTCP, portRange...)

	tests := []struct {
		name               string
		portStart          int64
		portStop           int64
		isUDP              bool
		isTCP              bool
		currentAllowlist   config.Allowlist
		configLoadErr      error
		expectedAllowlist  config.Allowlist
		expectedReturnCode int64
	}{
		{
			name:               "remove UDP port success",
			portStart:          port1,
			portStop:           0,
			isUDP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, port1, 2000}, []int64{22, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedReturnCode: internal.CodeSuccess,
		},
		{
			name:               "remove TCP port success",
			portStart:          port1,
			portStop:           0,
			isTCP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, 2000}, []int64{22, port1, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedReturnCode: internal.CodeSuccess,
		},
		{
			name:               "remove TCP/UDP port success",
			portStart:          port1,
			portStop:           0,
			isUDP:              true,
			isTCP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, port1, 2000}, []int64{22, port1, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedReturnCode: internal.CodeSuccess,
		},
		{
			name:               "remove TCP port range success",
			portStart:          port1,
			portStop:           port2,
			isTCP:              true,
			currentAllowlist:   config.NewAllowlist(currentPortRangeUDP, currentPortRangeTCP, []string{}),
			expectedAllowlist:  config.NewAllowlist(currentPortRangeUDP, []int64{22, 2500}, []string{}),
			expectedReturnCode: internal.CodeSuccess,
		},
		{
			name:               "remove UDP port range success",
			portStart:          port1,
			portStop:           port2,
			isUDP:              true,
			currentAllowlist:   config.NewAllowlist(currentPortRangeUDP, currentPortRangeTCP, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, 2000}, currentPortRangeTCP, []string{}),
			expectedReturnCode: internal.CodeSuccess,
		},
		{
			name:               "add TCP/UDP port range success",
			portStart:          port1,
			portStop:           port2,
			isUDP:              true,
			isTCP:              true,
			currentAllowlist:   config.NewAllowlist(currentPortRangeUDP, currentPortRangeTCP, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedReturnCode: internal.CodeSuccess,
		},
		// invalid UDP port
		{
			name:               "add UDP port failure invalid port(over range)",
			portStart:          invalidPort1,
			isUDP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedReturnCode: internal.CodeAllowlistPortOutOfRange,
		},
		{
			name:               "add UDP port failure invalid port(under range)",
			portStart:          invalidPort2,
			isUDP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedReturnCode: internal.CodeAllowlistPortOutOfRange,
		},
		// invalid TCP port
		{
			name:               "add TCP port failure invalid port(under range)",
			portStart:          invalidPort1,
			isTCP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedReturnCode: internal.CodeAllowlistPortOutOfRange,
		},
		{
			name:               "add TCP port failure invalid port(under range)",
			portStart:          invalidPort2,
			isTCP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedReturnCode: internal.CodeAllowlistPortOutOfRange,
		},
		// invalid TCP/UDP port
		{
			name:               "add TCP/UDP port failure invalid port(under range)",
			portStart:          invalidPort1,
			isTCP:              true,
			isUDP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedReturnCode: internal.CodeAllowlistPortOutOfRange,
		},
		{
			name:               "add TCP/UDP port failure invalid port(under range)",
			portStart:          invalidPort2,
			isTCP:              true,
			isUDP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedReturnCode: internal.CodeAllowlistPortOutOfRange,
		},
		// invalid UDP port range
		{
			name:               "add UDP port range failure invalid port(over range)",
			portStart:          invalidPort1,
			portStop:           port1,
			isUDP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedReturnCode: internal.CodeAllowlistPortOutOfRange,
		},
		{
			name:               "add UDP port range failure invalid port(under range)",
			portStart:          invalidPort2,
			portStop:           port1,
			isUDP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedReturnCode: internal.CodeAllowlistPortOutOfRange,
		},
		{
			name:               "add UDP port range failure invalid port(start port greater than stop port)",
			portStart:          port2,
			portStop:           port1,
			isUDP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedReturnCode: internal.CodeAllowlistPortOutOfRange,
		},
		// invalid TCP port range
		{
			name:               "add TCP port range failure invalid port(over range)",
			portStart:          invalidPort1,
			portStop:           port1,
			isTCP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedReturnCode: internal.CodeAllowlistPortOutOfRange,
		},
		{
			name:               "add TCP port range failure invalid port(under range)",
			portStart:          invalidPort2,
			portStop:           port1,
			isTCP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedReturnCode: internal.CodeAllowlistPortOutOfRange,
		},
		{
			name:               "add TCP port range failure invalid port(start port greater than stop port)",
			portStart:          port2,
			portStop:           port1,
			isTCP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedReturnCode: internal.CodeAllowlistPortOutOfRange,
		},
		// invalid TCP/UDP port range
		{
			name:               "add TCP/UDP port range failure invalid port(over range)",
			portStart:          invalidPort1,
			portStop:           port1,
			isTCP:              true,
			isUDP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedReturnCode: internal.CodeAllowlistPortOutOfRange,
		},
		{
			name:               "add TCP/UDP port range failure invalid port(under range)",
			portStart:          invalidPort2,
			portStop:           port1,
			isTCP:              true,
			isUDP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedReturnCode: internal.CodeAllowlistPortOutOfRange,
		},
		{
			name:               "add TCP/UDP port range failure invalid port(start port greater than stop port)",
			portStart:          port2,
			portStop:           port1,
			isTCP:              true,
			isUDP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{}),
			expectedReturnCode: internal.CodeAllowlistPortOutOfRange,
		},
		// config failure
		{
			name:               "add TCP/UDP port failure config failure",
			portStart:          port1,
			isTCP:              true,
			isUDP:              true,
			currentAllowlist:   config.NewAllowlist([]int64{}, []int64{}, []string{}),
			expectedAllowlist:  config.NewAllowlist([]int64{}, []int64{}, []string{}),
			configLoadErr:      fmt.Errorf("config load error"),
			expectedReturnCode: internal.CodeConfigError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			configManager, r := newMockedRPC(test.currentAllowlist, test.configLoadErr)

			req := pb.SetAllowlistRequest{
				Request: &pb.SetAllowlistRequest_SetAllowlistPortsRequest{
					SetAllowlistPortsRequest: &pb.SetAllowlistPortsRequest{
						IsUdp: test.isUDP,
						IsTcp: test.isTCP,
						PortRange: &pb.PortRange{
							StartPort: test.portStart,
							EndPort:   test.portStop,
						},
					},
				},
			}

			response, _ := r.UnsetAllowlist(context.Background(), &req)
			assert.Equal(t,
				test.expectedAllowlist,
				configManager.Cfg.AutoConnectData.Allowlist,
				"Allowlist configured incorrectly.")
			assert.Equal(t,
				test.expectedReturnCode,
				response.Type,
				"Invalid return code after setting allowlist.")
		})
	}
}

func TestUnsetAllAllowlist_Ports(t *testing.T) {
	category.Set(t, category.Unit)

	allowlist := config.NewAllowlist([]int64{44, 2000}, []int64{22, 2500}, []string{})
	configManager, r := newMockedRPC(allowlist, nil)

	response, _ := r.UnsetAllAllowlist(context.Background(), &pb.Empty{})

	assert.Equal(t,
		config.NewAllowlist([]int64{}, []int64{}, []string{}),
		configManager.Cfg.AutoConnectData.Allowlist,
		"Allowlist configured incorrectly.")
	assert.Equal(t, response.Type, internal.CodeSuccess,
		"Invalid return code after setting allowlist.")
}
