package daemon

import (
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/events"
)

type Publisher interface {
	SettingsPublisher
	ServicePublisher
}

func NewEvents(
	killswitch events.PublishSubcriber[bool],
	autoconnect events.PublishSubcriber[bool],
	dns events.PublishSubcriber[events.DataDNS],
	tplite events.PublishSubcriber[bool],
	protocol events.PublishSubcriber[config.Protocol],
	whitelist events.PublishSubcriber[events.DataWhitelist],
	technology events.PublishSubcriber[config.Technology],
	obfuscate events.PublishSubcriber[bool],
	firewall events.PublishSubcriber[bool],
	routing events.PublishSubcriber[bool],
	analytics events.PublishSubcriber[bool],
	notify events.PublishSubcriber[bool],
	meshnet events.PublishSubcriber[bool],
	ipv6 events.PublishSubcriber[bool],
	defaults events.PublishSubcriber[any],
	connect events.PublishSubcriber[events.DataConnect],
	disconnect events.PublishSubcriber[events.DataDisconnect],
	login events.PublishSubcriber[any],
	accountCheck events.PublishSubcriber[core.ServicesResponse],
	rate events.PublishSubcriber[events.ServerRating],
	heartBeat events.PublishSubcriber[int],
) *Events {
	return &Events{
		Settings: &SettingsEvents{
			Killswitch:           killswitch,
			Autoconnect:          autoconnect,
			DNS:                  dns,
			ThreatProtectionLite: tplite,
			Protocol:             protocol,
			Whitelist:            whitelist,
			Technology:           technology,
			Obfuscate:            obfuscate,
			Firewall:             firewall,
			Routing:              routing,
			Notify:               notify,
			Meshnet:              meshnet,
			Ipv6:                 ipv6,
			Defaults:             defaults,
		},
		Service: &ServiceEvents{
			Connect:      connect,
			Disconnect:   disconnect,
			Login:        login,
			AccountCheck: accountCheck,
			Rate:         rate,
			HeartBeat:    heartBeat,
		},
	}
}

type Events struct {
	Settings *SettingsEvents
	Service  *ServiceEvents
}

func (e *Events) Subscribe(to Publisher) {
	e.Settings.Subscribe(to)
	e.Service.Subscribe(to)
}

type SettingsPublisher interface {
	NotifyKillswitch(bool) error
	NotifyAutoconnect(bool) error
	NotifyDNS(events.DataDNS) error
	NotifyThreatProtectionLite(bool) error
	NotifyProtocol(config.Protocol) error
	NotifyWhitelist(events.DataWhitelist) error
	NotifyTechnology(config.Technology) error
	NotifyObfuscate(bool) error
	NotifyFirewall(bool) error
	NotifyRouting(bool) error
	NotifyNotify(bool) error
	NotifyMeshnet(bool) error
	NotifyIpv6(bool) error
	NotifyDefaults(any) error
}

type SettingsEvents struct {
	Killswitch           events.PublishSubcriber[bool]
	Autoconnect          events.PublishSubcriber[bool]
	DNS                  events.PublishSubcriber[events.DataDNS]
	ThreatProtectionLite events.PublishSubcriber[bool]
	Protocol             events.PublishSubcriber[config.Protocol]
	Whitelist            events.PublishSubcriber[events.DataWhitelist]
	Technology           events.PublishSubcriber[config.Technology]
	Obfuscate            events.PublishSubcriber[bool]
	Firewall             events.PublishSubcriber[bool]
	Routing              events.PublishSubcriber[bool]
	Notify               events.PublishSubcriber[bool]
	Meshnet              events.PublishSubcriber[bool]
	Ipv6                 events.PublishSubcriber[bool]
	Defaults             events.PublishSubcriber[any]
}

func (s *SettingsEvents) Subscribe(to SettingsPublisher) {
	s.Killswitch.Subscribe(to.NotifyKillswitch)
	s.Autoconnect.Subscribe(to.NotifyAutoconnect)
	s.DNS.Subscribe(to.NotifyDNS)
	s.ThreatProtectionLite.Subscribe(to.NotifyThreatProtectionLite)
	s.Protocol.Subscribe(to.NotifyProtocol)
	s.Whitelist.Subscribe(to.NotifyWhitelist)
	s.Technology.Subscribe(to.NotifyTechnology)
	s.Obfuscate.Subscribe(to.NotifyObfuscate)
	s.Firewall.Subscribe(to.NotifyFirewall)
	s.Routing.Subscribe(to.NotifyRouting)
	s.Notify.Subscribe(to.NotifyNotify)
	s.Meshnet.Subscribe(to.NotifyMeshnet)
	s.Ipv6.Subscribe(to.NotifyIpv6)
	s.Defaults.Subscribe(to.NotifyDefaults)
}

type ServicePublisher interface {
	NotifyConnect(events.DataConnect) error
	NotifyDisconnect(events.DataDisconnect) error
	NotifyLogin(any) error
	NotifyAccountCheck(core.ServicesResponse) error
	NotifyRate(events.ServerRating) error
	NotifyHeartBeat(int) error
}

type ServiceEvents struct {
	Connect      events.PublishSubcriber[events.DataConnect]
	Disconnect   events.PublishSubcriber[events.DataDisconnect]
	Login        events.PublishSubcriber[any]
	AccountCheck events.PublishSubcriber[core.ServicesResponse]
	Rate         events.PublishSubcriber[events.ServerRating]
	HeartBeat    events.PublishSubcriber[int]
}

func (s *ServiceEvents) Subscribe(to ServicePublisher) {
	s.Connect.Subscribe(to.NotifyConnect)
	s.Disconnect.Subscribe(to.NotifyDisconnect)
	s.Login.Subscribe(to.NotifyLogin)
	s.AccountCheck.Subscribe(to.NotifyAccountCheck)
	s.Rate.Subscribe(to.NotifyRate)
	s.HeartBeat.Subscribe(to.NotifyHeartBeat)
}

func (s *SettingsEvents) Publish(cfg config.Config) {
	s.Killswitch.Publish(cfg.KillSwitch)
	s.Firewall.Publish(cfg.Firewall)
	s.Routing.Publish(cfg.Routing.Get())
	s.Autoconnect.Publish(cfg.AutoConnect)
	s.DNS.Publish(events.DataDNS{Enabled: len(cfg.AutoConnectData.DNS) != 0, Ips: cfg.AutoConnectData.DNS})
	s.ThreatProtectionLite.Publish(cfg.AutoConnectData.ThreatProtectionLite)
	s.Protocol.Publish(cfg.AutoConnectData.Protocol)
	s.Whitelist.Publish(events.DataWhitelist{
		TCPPorts: len(cfg.AutoConnectData.Whitelist.Ports.TCP),
		UDPPorts: len(cfg.AutoConnectData.Whitelist.Ports.UDP),
		Subnets:  len(cfg.AutoConnectData.Whitelist.Subnets),
	})
	s.Meshnet.Publish(cfg.Mesh)
	s.Ipv6.Publish(cfg.IPv6)
	s.Technology.Publish(cfg.Technology)
}
