package events

import (
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
)

type Publisher interface {
	SettingsPublisher
	ServicePublisher
}

func NewEventsEmpty() *Events {
	return NewEvents(
		&subs.Subject[bool]{},
		&subs.Subject[bool]{},
		&subs.Subject[events.DataDNS]{},
		&subs.Subject[bool]{},
		&subs.Subject[config.Protocol]{},
		&subs.Subject[events.DataAllowlist]{},
		&subs.Subject[config.Technology]{},
		&subs.Subject[bool]{},
		&subs.Subject[bool]{},
		&subs.Subject[bool]{},
		&subs.Subject[bool]{},
		&subs.Subject[bool]{},
		&subs.Subject[bool]{},
		&subs.Subject[bool]{},
		&subs.Subject[any]{},
		&subs.Subject[events.DataConnect]{},
		&subs.Subject[events.DataDisconnect]{},
		&subs.Subject[events.DataAuthorization]{},
		&subs.Subject[events.DataAuthorization]{},
		&subs.Subject[core.ServicesResponse]{},
		&subs.Subject[events.UiItemsAction]{},
		&subs.Subject[int]{},
		&subs.Subject[core.Insights]{},
		&subs.Subject[bool]{},
		&subs.Subject[bool]{},
	)
}

func NewEvents(
	killswitch events.PublishSubcriber[bool],
	autoconnect events.PublishSubcriber[bool],
	dns events.PublishSubcriber[events.DataDNS],
	tplite events.PublishSubcriber[bool],
	protocol events.PublishSubcriber[config.Protocol],
	allowlist events.PublishSubcriber[events.DataAllowlist],
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
	login events.PublishSubcriber[events.DataAuthorization],
	logout events.PublishSubcriber[events.DataAuthorization],
	accountCheck events.PublishSubcriber[core.ServicesResponse],
	uiItemsClick events.PublishSubcriber[events.UiItemsAction],
	heartBeat events.PublishSubcriber[int],
	deviceLocation events.PublishSubcriber[core.Insights],
	lanDiscovery events.PublishSubcriber[bool],
	virtualLocation events.PublishSubcriber[bool],
) *Events {
	return &Events{
		Settings: &SettingsEvents{
			Killswitch:           killswitch,
			Autoconnect:          autoconnect,
			DNS:                  dns,
			ThreatProtectionLite: tplite,
			Protocol:             protocol,
			Allowlist:            allowlist,
			Technology:           technology,
			Obfuscate:            obfuscate,
			Firewall:             firewall,
			Routing:              routing,
			Notify:               notify,
			Meshnet:              meshnet,
			Ipv6:                 ipv6,
			Defaults:             defaults,
			LANDiscovery:         lanDiscovery,
			VirtualLocation:      virtualLocation,
		},
		Service: &ServiceEvents{
			Connect:        connect,
			Disconnect:     disconnect,
			Login:          login,
			Logout:         logout,
			AccountCheck:   accountCheck,
			UiItemsClick:   uiItemsClick,
			HeartBeat:      heartBeat,
			DeviceLocation: deviceLocation,
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
	NotifyAllowlist(events.DataAllowlist) error
	NotifyTechnology(config.Technology) error
	NotifyObfuscate(bool) error
	NotifyFirewall(bool) error
	NotifyRouting(bool) error
	NotifyNotify(bool) error
	NotifyMeshnet(bool) error
	NotifyIpv6(bool) error
	NotifyDefaults(any) error
	NotifyLANDiscovery(bool) error
	NotifyVirtualLocation(bool) error
}

type SettingsEvents struct {
	Killswitch           events.PublishSubcriber[bool]
	Autoconnect          events.PublishSubcriber[bool]
	DNS                  events.PublishSubcriber[events.DataDNS]
	ThreatProtectionLite events.PublishSubcriber[bool]
	Protocol             events.PublishSubcriber[config.Protocol]
	Allowlist            events.PublishSubcriber[events.DataAllowlist]
	Technology           events.PublishSubcriber[config.Technology]
	Obfuscate            events.PublishSubcriber[bool]
	Firewall             events.PublishSubcriber[bool]
	Routing              events.PublishSubcriber[bool]
	Notify               events.PublishSubcriber[bool]
	Meshnet              events.PublishSubcriber[bool]
	Ipv6                 events.PublishSubcriber[bool]
	Defaults             events.PublishSubcriber[any]
	LANDiscovery         events.PublishSubcriber[bool]
	VirtualLocation      events.PublishSubcriber[bool]
}

func (s *SettingsEvents) Subscribe(to SettingsPublisher) {
	s.Killswitch.Subscribe(to.NotifyKillswitch)
	s.Autoconnect.Subscribe(to.NotifyAutoconnect)
	s.DNS.Subscribe(to.NotifyDNS)
	s.ThreatProtectionLite.Subscribe(to.NotifyThreatProtectionLite)
	s.Protocol.Subscribe(to.NotifyProtocol)
	s.Allowlist.Subscribe(to.NotifyAllowlist)
	s.Technology.Subscribe(to.NotifyTechnology)
	s.Obfuscate.Subscribe(to.NotifyObfuscate)
	s.Firewall.Subscribe(to.NotifyFirewall)
	s.Routing.Subscribe(to.NotifyRouting)
	s.Notify.Subscribe(to.NotifyNotify)
	s.Meshnet.Subscribe(to.NotifyMeshnet)
	s.Ipv6.Subscribe(to.NotifyIpv6)
	s.Defaults.Subscribe(to.NotifyDefaults)
	s.LANDiscovery.Subscribe(to.NotifyLANDiscovery)
	s.VirtualLocation.Subscribe(to.NotifyVirtualLocation)
}

type ServicePublisher interface {
	NotifyConnect(events.DataConnect) error
	NotifyDisconnect(events.DataDisconnect) error
	NotifyLogin(events.DataAuthorization) error
	NotifyLogout(events.DataAuthorization) error
	NotifyAccountCheck(core.ServicesResponse) error
	NotifyUiItemsClick(events.UiItemsAction) error
	NotifyHeartBeat(int) error
	NotifyDeviceLocation(core.Insights) error
}

type ServiceEvents struct {
	Connect        events.PublishSubcriber[events.DataConnect]
	Disconnect     events.PublishSubcriber[events.DataDisconnect]
	Login          events.PublishSubcriber[events.DataAuthorization]
	Logout         events.PublishSubcriber[events.DataAuthorization]
	AccountCheck   events.PublishSubcriber[core.ServicesResponse]
	UiItemsClick   events.PublishSubcriber[events.UiItemsAction]
	HeartBeat      events.PublishSubcriber[int]
	DeviceLocation events.PublishSubcriber[core.Insights]
}

func (s *ServiceEvents) Subscribe(to ServicePublisher) {
	s.Connect.Subscribe(to.NotifyConnect)
	s.Disconnect.Subscribe(to.NotifyDisconnect)
	s.Login.Subscribe(to.NotifyLogin)
	s.Logout.Subscribe(to.NotifyLogout)
	s.AccountCheck.Subscribe(to.NotifyAccountCheck)
	s.UiItemsClick.Subscribe(to.NotifyUiItemsClick)
	s.HeartBeat.Subscribe(to.NotifyHeartBeat)
	s.DeviceLocation.Subscribe(to.NotifyDeviceLocation)
}

func (s *SettingsEvents) Publish(cfg config.Config) {
	s.Killswitch.Publish(cfg.KillSwitch)
	s.Firewall.Publish(cfg.Firewall)
	s.Routing.Publish(cfg.Routing.Get())
	s.Autoconnect.Publish(cfg.AutoConnect)
	s.DNS.Publish(events.DataDNS{Ips: cfg.AutoConnectData.DNS})
	s.ThreatProtectionLite.Publish(cfg.AutoConnectData.ThreatProtectionLite)
	s.Protocol.Publish(cfg.AutoConnectData.Protocol)
	s.Allowlist.Publish(events.DataAllowlist{
		TCPPorts: cfg.AutoConnectData.Allowlist.Ports.TCP.ToSlice(),
		UDPPorts: cfg.AutoConnectData.Allowlist.Ports.UDP.ToSlice(),
		Subnets:  cfg.AutoConnectData.Allowlist.Subnets.ToSlice(),
	})
	s.Meshnet.Publish(cfg.Mesh)
	s.Ipv6.Publish(cfg.IPv6)
	s.Technology.Publish(cfg.Technology)
	s.Obfuscate.Publish(cfg.AutoConnectData.Obfuscate)
	s.Notify.Publish(!(cfg.UsersData.NotifyOff != nil && len(cfg.UsersData.NotifyOff) > 0))
	s.LANDiscovery.Publish(cfg.LanDiscovery)
	s.VirtualLocation.Publish(cfg.VirtualLocation.Get())
}

type MockPublisherSubscriber[T any] struct {
	EventPublished bool
}

func (mp *MockPublisherSubscriber[T]) Publish(message T) {
	mp.EventPublished = true
}
func (*MockPublisherSubscriber[T]) Subscribe(handler events.Handler[T]) {}
