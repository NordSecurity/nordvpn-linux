package mock

import (
	"net"
	"net/netip"

	"github.com/NordSecurity/nordvpn-linux/tunnel"
)

// WorkingT stub of a github.com/NordSecurity/nordvpn-linux/tunnel.T interface.
type WorkingT struct{}

func (WorkingT) Interface() net.Interface { return En0Interface }
func (WorkingT) IPs() []netip.Addr {
	return []netip.Addr{netip.MustParseAddr("127.0.0.1")}
}

func (WorkingT) TransferRates() (tunnel.Statistics, error) {
	return tunnel.Statistics{Tx: 1337, Rx: 1337}, nil
}

// WorkingIPv6T stub of a github.com/NordSecurity/nordvpn-linux/tunnel.T interface.
type WorkingIPv6T struct{}

func (WorkingIPv6T) Interface() net.Interface { return En0Interface }
func (WorkingIPv6T) IPs() []netip.Addr {
	return []netip.Addr{netip.MustParseAddr("fde0:9c97:ec39:4691:6323:2d46:3321:9688")}
}

func (WorkingIPv6T) TransferRates() (tunnel.Statistics, error) {
	return tunnel.Statistics{Tx: 1337, Rx: 1337}, nil
}

type FailingTunnel struct{}

func (FailingTunnel) Interface() net.Interface { return net.Interface{} }
func (FailingTunnel) IPs() []netip.Addr        { return nil }
func (FailingTunnel) TransferRates() (tunnel.Statistics, error) {
	return tunnel.Statistics{}, ErrOnPurpose
}
