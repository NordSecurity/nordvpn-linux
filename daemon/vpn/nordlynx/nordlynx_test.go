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
		devType := "wireguard"
		defer removeDevice(device)

		err := addDevice(device, devType)
		assert.NoError(t, err)
	})

	t.Run("duplicate adding", func(t *testing.T) {
		device := "faildev"
		devType := "wireguard"
		defer removeDevice(device)

		err := addDevice(device, devType)
		assert.NoError(t, err)

		err = addDevice(device, devType)
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
