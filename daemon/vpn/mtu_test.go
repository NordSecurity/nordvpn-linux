package vpn

import (
	"net"
	"os/exec"
	"strconv"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/device"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"gotest.tools/v3/assert"
)

const testHeaderSize = 30
const testInterfaceName = "testif"

func TestSetMTU(t *testing.T) {
	category.Set(t, category.Link)

	tests := []struct {
		name        string
		currentMTU  int
		expectedMTU int
	}{
		// {
		// 	name:        "set MTU underlying MTU is 1500",
		// 	currentMTU:  1500,
		// 	expectedMTU: 1500 - testHeaderSize,
		// },
		{
			name:        "set MTU underlying MTU is 1000",
			currentMTU:  1450,
			expectedMTU: 1450 - testHeaderSize,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defaultGateway, _ := device.DefaultGateway()
			exec.Command("ip", "link", "set", "dev", defaultGateway.Name, "mtu", strconv.Itoa(test.currentMTU)).Run()
			defer exec.Command("ip",
				"link",
				"set",
				"dev",
				defaultGateway.Name,
				"mtu",
				strconv.Itoa(defaultGateway.MTU)).Run()

			exec.Command("ip", "link", "add", testInterfaceName, "type", "dummy").Run()
			exec.Command("sudo", "ip", "addr", "add", "192.168.100.1/24", "dev", "mydummy0").Run()
			exec.Command("sudo", "ip", "link", "set", "mydummy0", "up").Run()

			iface, _ := net.InterfaceByName(testInterfaceName)
			SetMTU(*iface, testHeaderSize)

			iface, _ = net.InterfaceByName(testInterfaceName)
			assert.Equal(t, test.expectedMTU, iface.MTU)
		})
	}
}
