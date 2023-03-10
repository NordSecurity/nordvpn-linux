package tunnel

import (
	"net"
	"net/netip"

	"github.com/NordSecurity/nordvpn-linux/test/device"
	"github.com/NordSecurity/nordvpn-linux/tunnel"
)

// Working stub of a github.com/NordSecurity/nordvpn-linux/tunnel.T interface.
type Working struct{}

func (Working) Interface() net.Interface { return device.En0Interface }
func (Working) IPs() []netip.Addr {
	return []netip.Addr{netip.MustParseAddr("127.0.0.1")}
}

func (Working) TransferRates() (tunnel.Statistics, error) {
	return tunnel.Statistics{Tx: 1337, Rx: 1337}, nil
}

// WorkingIPv6 stub of a github.com/NordSecurity/nordvpn-linux/tunnel.T interface.
type WorkingIPv6 struct{}

func (WorkingIPv6) Interface() net.Interface { return device.En0Interface }
func (WorkingIPv6) IPs() []netip.Addr {
	return []netip.Addr{netip.MustParseAddr("fde0:9c97:ec39:4691:6323:2d46:3321:9688")}
}

func (WorkingIPv6) TransferRates() (tunnel.Statistics, error) {
	return tunnel.Statistics{Tx: 1337, Rx: 1337}, nil
}
