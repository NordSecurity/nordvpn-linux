package device

import (
	"net"
	"os"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"github.com/NordSecurity/nordvpn-linux/test/mock/fs"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"

	"github.com/stretchr/testify/assert"
)

func TestInterfaceNameFromIpRoute(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		input    string
		output   string
		hasError bool
	}{
		{
			input:    "172.17.0.0/16 dev docker0 proto kernel scope link src 172.17.0.1 linkdown",
			output:   "docker0",
			hasError: false,
		},
		{
			input:    "172.17.0.0/16 br0 proto kernel scope link src 172.17.0.1 linkdown",
			output:   "",
			hasError: true,
		},
		{
			input:    "",
			output:   "",
			hasError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			name, err := interfaceNameFromIPRoute(test.input)
			assert.Equal(t, test.hasError, err != nil)
			assert.Equal(t, name, test.output)
		})
	}
}

func TestInterfacesWithDefaultRoute(t *testing.T) {
	category.Set(t, category.Unit)

	setMockSysDeps(t)
	expectedNames := []string{"en2", "virtual10"}
	interfaces := InterfacesWithDefaultRoute(nil)
	var names []string
	for _, iface := range interfaces {
		names = append(names, iface.Name)
	}
	assert.ElementsMatch(t, expectedNames, names)
}

func TestListBlockedInterfaces(t *testing.T) {
	category.Set(t, category.Unit)

	setMockSysDeps(t)
	interfaces, err := OutsideCapableTrafficInterfaces()
	assert.NoError(t, err)

	expectedNames := []string{"en2", "en3", "en4", "virtual10", "virtual9"}
	var names []string
	for _, iface := range interfaces {
		names = append(names, iface.Name)
	}
	assert.ElementsMatch(t, expectedNames, names)

	// check that the interface names set is correct
	assert.Equal(t, mapset.NewSet(expectedNames...), OutsideCapableTrafficIfNames(nil))

	// when all interfaces are ignored returns empty list
	assert.True(t, OutsideCapableTrafficIfNames(mapset.NewSet(expectedNames...)).IsEmpty())
}

func mustCIDR(t *testing.T, s string) *net.IPNet {
	t.Helper()
	_, n, err := net.ParseCIDR(s)
	if err != nil {
		t.Fatalf("ParseCIDR(%q): %v", s, err)
	}
	return n
}

func setMockSysDeps(t *testing.T) {
	t.Helper()

	originalSysDeps := sysDepsImpl
	t.Cleanup(func() {
		sysDepsImpl = originalSysDeps
	})

	sysDepsImpl = &mock.MockSystemDeps{
		ExistingFiles: []string{
			"/sys/class/net/en2/device",
			"/sys/class/net/en3/device",
			"/sys/class/net/en4/device",
		},
		InterfacesList: []net.Interface{
			{Index: 2, Name: "en2"},
			{Index: 3, Name: "en3"},
			{Index: 4, Name: "en4"},
			{Index: 9, Name: "virtual9"},
			{Index: 10, Name: "virtual10"},
			{Index: 11, Name: "virtual11"},
		},
		RouteListRoutes: []netlink.Route{
			// default (Dst nil means 0/0)
			{Dst: nil, LinkIndex: 2, Table: unix.RT_TABLE_MAIN, Priority: 100},

			// non-default route should be ignored
			{Dst: mustCIDR(t, "10.0.0.0/8"), LinkIndex: 2},

			// duplicate default on same iface should be deduped
			{Dst: nil, LinkIndex: 2, Table: unix.RT_TABLE_MAIN, Priority: 50},

			// physical interface without default route
			{Dst: mustCIDR(t, "172.17.0.1/16"), LinkIndex: 3, Table: unix.RT_TABLE_MAIN, Priority: 200},

			// default on a virtual interface
			{Dst: mustCIDR(t, "0.0.0.0/0"), LinkIndex: 10},

			// docker0
			{Dst: mustCIDR(t, "172.17.0.1/16"), LinkIndex: 11},

			// 1.1.1.1 via 192.168.0.1 dev br0
			{
				Dst: &net.IPNet{
					IP:   net.IPv4(1, 1, 1, 1),
					Mask: net.CIDRMask(32, 32),
				},
				Gw:        net.IPv4(192, 168, 0, 1),
				LinkIndex: 9,
				Scope:     netlink.SCOPE_UNIVERSE,
			},
		},
		ReadDirEntries: []os.DirEntry{
			&fs.MockDirEntry{DirName: "virtual9"},
			&fs.MockDirEntry{DirName: "virtual10"},
			&fs.MockDirEntry{DirName: "virtual11"},
		},
	}
}
