package networker

import (
	"context"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	firewallmock "github.com/NordSecurity/nordvpn-linux/test/mock/firewall"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// callOrder records the sequence in which the firewall and the VPN are driven.
type callOrder struct {
	events []string
}

type recordingFirewall struct {
	firewall.Service
	order *callOrder
}

func (f *recordingFirewall) Configure(cfg firewall.Config) error {
	f.order.events = append(f.order.events, "firewall")
	return f.Service.Configure(cfg)
}

type recordingVPN struct {
	vpn.VPN
	order *callOrder
}

func (v *recordingVPN) Start(ctx context.Context, creds vpn.Credentials, server vpn.ServerData) error {
	v.order.events = append(v.order.events, "vpn")
	return v.VPN.Start(ctx, creds, server)
}

// The firewall's output chain marks the VPN transport connection with
// "meta mark -> ct mark set", a rule that only matches packets carrying
// skb->mark, i.e. packets sent by the VPN process itself. With OpenVPN DCO the
// kernel module owns the socket once the handshake completes and never sets
// skb->mark again, so the handshake is the only chance to mark the connection.
// Configuring the firewall after the handshake leaves the conntrack entry
// unmarked and the post-connect drop policy then blackholes the data channel
// for the rest of the session.
func TestCombined_Start_ConfiguresFirewallBeforeVPNStarts(t *testing.T) {
	category.Set(t, category.Unit)

	order := &callOrder{}

	netw := NewCombined(
		&recordingVPN{VPN: &mock.WorkingVPN{}, order: order},
		nil,
		workingGateway{},
		&subs.Subject[string]{},
		workingRouter{},
		&workingDNS{},
		&recordingFirewall{Service: firewallmock.NewFirewall(), order: order},
		nil,
		&workingRoutingSetup{},
		nil,
		workingRouter{},
		nil,
		0,
		false,
		&workingIpv6{},
		false,
		&mock.SysctlSetterMock{},
		config.Allowlist{},
		&mock.SysctlSetterMock{},
	)

	err := netw.Start(
		context.Background(),
		vpn.Credentials{},
		vpn.ServerData{},
		config.NewAllowlist(nil, nil, nil),
		[]string{"1.1.1.1"},
		true,
		func(time.Time, error) {},
	)
	require.NoError(t, err)

	require.NotEmpty(t, order.events, "neither the firewall nor the VPN was driven")
	assert.Equal(t, "firewall", order.events[0],
		"the firewall must be configured before the VPN handshake, otherwise the "+
			"transport connection (OVPN) is never marked and DCO traffic is dropped")
	assert.Contains(t, order.events, "vpn", "the VPN was never started")
}
