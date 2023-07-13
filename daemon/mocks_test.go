package daemon

import (
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/networker"
)

type mockConfigManagerCommon struct {
	dns                  []string
	threatProtectionLite bool
	ipv6                 bool
	protocol             config.Protocol
	technology           config.Technology
	saveConfigErr        error
}

func (m *mockConfigManagerCommon) SaveWith(f config.SaveFunc) error {
	if m.saveConfigErr != nil {
		return m.saveConfigErr
	}

	var conf config.Config
	conf = f(conf)
	m.dns = conf.AutoConnectData.DNS
	m.threatProtectionLite = conf.AutoConnectData.ThreatProtectionLite
	m.protocol = conf.AutoConnectData.Protocol
	m.technology = conf.Technology
	return nil
}

func (m *mockConfigManagerCommon) Load(c *config.Config) error {
	c.AutoConnectData.DNS = m.dns
	c.AutoConnectData.ThreatProtectionLite = m.threatProtectionLite
	c.IPv6 = m.ipv6
	c.AutoConnectData.Protocol = m.protocol
	c.Technology = m.technology
	return nil
}

func (*mockConfigManagerCommon) Reset() error {
	return nil
}

type mockNetworker struct {
	dns       []string
	setDNSErr error
	vpnActive bool
}

func (mockNetworker) Start(
	vpn.Credentials,
	vpn.ServerData,
	config.Whitelist,
	config.DNS,
) error {
	return nil
}
func (*mockNetworker) Stop() error      { return nil }
func (*mockNetworker) UnSetMesh() error { return nil }

func (mn *mockNetworker) SetDNS(nameservers []string) error {
	mn.dns = nameservers
	return mn.setDNSErr
}

func (*mockNetworker) UnsetDNS() error { return nil }

func (mn *mockNetworker) IsVPNActive() bool {
	return mn.vpnActive
}

func (*mockNetworker) ConnectionStatus() (networker.ConnectionStatus, error) {
	return networker.ConnectionStatus{}, nil
}
func (*mockNetworker) EnableFirewall() error                { return nil }
func (*mockNetworker) DisableFirewall() error               { return nil }
func (*mockNetworker) EnableRouting()                       {}
func (*mockNetworker) DisableRouting()                      {}
func (*mockNetworker) SetWhitelist(config.Whitelist) error  { return nil }
func (*mockNetworker) UnsetWhitelist() error                { return nil }
func (*mockNetworker) IsNetworkSet() bool                   { return false }
func (*mockNetworker) SetKillSwitch(config.Whitelist) error { return nil }
func (*mockNetworker) UnsetKillSwitch() error               { return nil }
func (*mockNetworker) PermitIPv6() error                    { return nil }
func (*mockNetworker) DenyIPv6() error                      { return nil }
func (*mockNetworker) SetVPN(vpn.VPN)                       {}
func (*mockNetworker) LastServerName() string               { return "" }

var tplNameserversV4 []string = []string{
	"103.86.96.96",
	"103.86.99.99",
}

var tplNameserversV6 []string = []string{
	"2400:bb40:4444::103",
	"2400:bb40:8888::103",
}

var defaultNameserversV4 []string = []string{
	"103.86.96.100",
	"103.86.99.100",
}

var defaultNameserversV6 []string = []string{
	"2400:bb40:4444::100",
	"2400:bb40:8888::100",
}

type mockDNSGetter struct {
}

func (md *mockDNSGetter) Get(isThreatProtectionLite bool, isIPv6 bool) []string {
	if isThreatProtectionLite {
		nameservers := tplNameserversV4
		if isIPv6 {
			nameservers = append(nameservers, tplNameserversV6...)
		}
		return nameservers
	}

	nameservers := defaultNameserversV4
	if isIPv6 {
		nameservers = append(nameservers, defaultNameserversV6...)
	}
	return nameservers
}

type mockPublisherSubcriber struct {
	eventPublished bool
}

func (mp *mockPublisherSubcriber) Publish(message bool) {
	mp.eventPublished = true
}
func (*mockPublisherSubcriber) Subscribe(handler events.Handler[bool]) {}
