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
