package nordlynx

import (
	"net"
	"os/exec"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

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

func TestUpWGInterface(t *testing.T) {
	category.Set(t, category.Link)

	iName := "winterface"
	_, err := net.InterfaceByName(iName)
	assert.Error(t, err)

	err = upWGInterface(iName)
	assert.NoError(t, err)

	iface, err := net.InterfaceByName(iName)
	assert.NoError(t, err)
	defer deleteInterface(*iface)
}

func TestAddDevice(t *testing.T) {
	category.Set(t, category.Link)

	t.Run("successful adding", func(t *testing.T) {
		device := "testdev"
		defer removeDevice(device)

		err := addDevice(device)
		assert.NoError(t, err)
	})

	t.Run("duplicate adding", func(t *testing.T) {
		device := "faildev"
		defer removeDevice(device)

		err := addDevice(device)
		assert.NoError(t, err)

		err = addDevice(device)
		assert.Error(t, err)
	})
}

func TestRemoveDevice(t *testing.T) {
	category.Set(t, category.Link)

	t.Run("non existing device", func(t *testing.T) {
		device := "nodev"

		_, err := removeDevice(device)
		assert.Error(t, err)
	})
}

func TestCalculateMTU(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name          string
		ipRouteOutput string
		expectedMTU   int
	}{
		{
			name:        "no default route exist",
			expectedMTU: defaultMTU - wireguardHeaderSize,
		},
		{
			name:          "incorrect ip route output",
			ipRouteOutput: "default via interface_name proto dhcp metric 600",
			expectedMTU:   defaultMTU - wireguardHeaderSize,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mtu := calculateMTU(test.ipRouteOutput)
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

	mtu := retrieveAndCalculateMTU()
	assert.Equal(t, defaultGateway.MTU-wireguardHeaderSize, mtu)
}
