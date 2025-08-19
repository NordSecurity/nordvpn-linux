package vpn

import (
	"net"
	"os/exec"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

const testHeaderSize = 30

func TestCalculateMTU(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name          string
		ipRouteOutput string
		expectedMTU   int
	}{
		{
			name:        "no default route exist",
			expectedMTU: defaultMTU - testHeaderSize,
		},
		{
			name:          "incorrect ip route output",
			ipRouteOutput: "default via interface_name proto dhcp metric 600",
			expectedMTU:   defaultMTU - testHeaderSize,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mtu := calculateMTU(test.ipRouteOutput, testHeaderSize)
			assert.Equal(t, mtu, test.expectedMTU)
		})
	}
}

func TestRetrieveAndCalculateMTU(t *testing.T) {
	category.Set(t, category.Link)

	out, err := exec.Command("ip", "route", "show", "default").Output()
	assert.NoError(t, err)

	defaultGatewayName, err := getDefaultIpRouteInterface(string(out))
	assert.NoError(t, err)

	defaultGateway, err := net.InterfaceByName(defaultGatewayName)
	assert.NoError(t, err)

	mtu := retrieveAndCalculateMTU(testHeaderSize)
	assert.Equal(t, defaultGateway.MTU-testHeaderSize, mtu)
}

func TestGetDefaultIpRouteInterface(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		testName              string
		ipRoutes              string
		expectedInterfaceName string
		expectedError         error
	}{
		{
			testName:              "one route",
			ipRoutes:              "default via 192.168.0.4 dev wlp0s2f4 proto dhcp metric 600\n",
			expectedInterfaceName: "wlp0s2f4",
			expectedError:         nil,
		},
		{
			testName: "two routes",
			ipRoutes: `default via 192.168.0.4 dev wlp0s2f4 proto dhcp metric 600\n
					   default via 192.168.0.5 dev enp0s2f6 proto dhcp metric 600\n`,
			expectedInterfaceName: "wlp0s2f4",
			expectedError:         nil,
		},
		{
			testName:              "no routes",
			ipRoutes:              "",
			expectedInterfaceName: "",
			expectedError:         errNoDefaultIpRoute,
		},
		{
			testName:              "unrecognized output",
			ipRoutes:              "default via bad output",
			expectedInterfaceName: "",
			expectedError:         errUnrecognizedIpRouteOutput,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			ifName, err := getDefaultIpRouteInterface(test.ipRoutes)

			assert.Equal(t, ifName, test.expectedInterfaceName)
			assert.Equal(t, err, test.expectedError)
		})
	}
}

func TestGetDefaultIpRouteInterfaceFromCommandOutput(t *testing.T) {
	category.Set(t, category.Route)

	out, _ := exec.Command("ip", "route", "show", "default").Output()
	outString := string(out)

	// assume one default route
	expectedIfName := strings.Split(outString, " ")[4]

	ifName, _ := getDefaultIpRouteInterface(string(out))

	assert.Equal(t, expectedIfName, ifName)
}
