package tunnel

import (
	"net"
	"net/netip"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/milosgajdos/tenus"
	"github.com/stretchr/testify/assert"
)

func TestFind(t *testing.T) {
	category.Set(t, category.Link)

	tunnelName := "nordtuna"
	ip := net.IPv4(192, 254, 0, 123)
	ipnet := net.IPNet{
		IP:   ip,
		Mask: net.IPv4Mask(255, 255, 255, 255),
	}

	defer tenus.DeleteLink(tunnelName)
	iface, err := tenus.NewLink(tunnelName)
	assert.NoError(t, err)
	assert.NotNil(t, iface)
	err = iface.SetLinkIp(ip, &ipnet)
	assert.NoError(t, err)

	address, ok := netip.AddrFromSlice(ip)
	assert.True(t, ok)
	tun, err := Find(address)
	assert.NoError(t, err)
	assert.Equal(t, tunnelName, tun.iface.Name)
	address, ok = tun.IP()
	assert.Equal(t, ip, address)
}

func TestTunnel_TransferRates(t *testing.T) {
	category.Set(t, category.Integration)

	_, err := Tunnel{}.TransferRates()
	assert.Error(t, err)

	iface, err := net.InterfaceByName("lo")
	assert.NoError(t, err)

	_, err = Tunnel{iface: *iface}.TransferRates()
	assert.NoError(t, err)
}

func TestFromDummy(t *testing.T) {
	category.Set(t, category.Link)

	ifaceName := "nordtest0"
	ipAddr := netip.MustParseAddr("192.254.0.123")
	err := exec.Command("ip",
		[]string{
			"link", "add", ifaceName, "type", "dummy",
		}...,
	).Run()
	assert.NoError(t, err)
	defer exec.Command("ip", "link", "del", ifaceName).Run()

	err = exec.Command("ip",
		[]string{
			"addr", "add", ipAddr.String(), "dev", ifaceName,
		}...,
	).Run()
	assert.NoError(t, err)

	got, err := Find(ipAddr)
	assert.NoError(t, err)
	assert.Equal(t, ifaceName, got.iface.Name)
	assert.Equal(t, ipAddr, got.prefix.Addr())
}

func TestTunnelTransferRatesWithSys(t *testing.T) {
	category.Set(t, category.Integration)

	nonExistent := Tunnel{
		iface: net.Interface{Name: "iface0321"},
	}
	_, err := nonExistent.TransferRates()
	assert.Error(t, err)

	paths, err := filepath.Glob("/sys/class/net/*")
	assert.NoError(t, err)

	for _, path := range paths {
		pathParts := strings.Split(path, "/")
		path = pathParts[len(pathParts)-1]
		tun := Tunnel{
			iface: net.Interface{Name: path},
		}
		_, err := tun.TransferRates()
		assert.NoError(t, err)
	}
}

func TestTunnel_AddAddrs(t *testing.T) {
	category.Set(t, category.Link)

	iface, err := net.InterfaceByName("lo")
	assert.NoError(t, err)

	prefix := netip.MustParsePrefix("10.121.0.2/32")
	assert.NotContains(t, getIPs(t, iface), prefix.Addr())
	tunnel := &Tunnel{
		iface:  *iface,
		prefix: prefix,
	}

	err = tunnel.AddAddrs()
	assert.NoError(t, err)

	iface, err = net.InterfaceByName("lo")
	assert.NoError(t, err)
	assert.Contains(t, getIPs(t, iface), prefix)
}

func TestTunnel_Up(t *testing.T) {
	category.Set(t, category.Link)

	out, err := exec.Command("ip", "link", "set", "lo", "down").CombinedOutput()
	assert.NoError(t, err, string(out))

	iface, err := net.InterfaceByName("lo")
	assert.NoError(t, err)

	tunnel := &Tunnel{
		iface: *iface,
	}

	err = tunnel.Up()
	assert.NoError(t, err)

	iface, err = net.InterfaceByName("lo")
	assert.NoError(t, err)
	assert.Equal(t, net.FlagUp, iface.Flags)
}

func getIPs(t *testing.T, iface *net.Interface) []net.IP {
	t.Helper()

	var ips []net.IP
	addrs, err := iface.Addrs()
	assert.NoError(t, err)

	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		ips = append(ips, ip)
	}
	return ips
}
