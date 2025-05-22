package vpn

import (
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/tunnel"
)

type InternalVPNPublisher interface {
	ConnectionStatusNotifyConnect(e events.DataConnect) error
	ConnectionStatusNotifyDisconnect(_ events.DataDisconnect) error
}

type Events struct {
	Connected    events.PublishSubcriber[events.DataConnect]
	Disconnected events.PublishSubcriber[events.DataDisconnect]
}

func NewInternalVPNEvents() *Events {
	return &Events{
		Connected:    &subs.Subject[events.DataConnect]{},
		Disconnected: &subs.Subject[events.DataDisconnect]{},
	}
}

func (e *Events) Subscribe(to InternalVPNPublisher) {
	e.Connected.Subscribe(to.ConnectionStatusNotifyConnect)
	e.Disconnected.Subscribe(to.ConnectionStatusNotifyDisconnect)
}

func GetDataConnectEvent(technology config.Technology,
	protocol config.Protocol,
	connectType events.TypeEventStatus,
	server ServerData,
	tunnelStatistics tunnel.Statistics,
	isMeshPeer bool) events.DataConnect {
	return events.DataConnect{
		EventStatus:             connectType,
		TargetServerIP:          server.IP.String(),
		TargetServerCountry:     server.Country,
		TargetServerCountryCode: server.CountryCode,
		TargetServerCity:        server.City,
		TargetServerDomain:      server.Hostname,
		TargetServerName:        server.Name,
		IsMeshnetPeer:           isMeshPeer,
		IsVirtualLocation:       server.VirtualLocation,
		Technology:              technology,
		Protocol:                protocol,
		Upload:                  tunnelStatistics.Tx,
		Download:                tunnelStatistics.Rx,
		IsObfuscated:            server.Obfuscated,
		IsPostQuantum:           server.PostQuantum,
		IP:                      server.IP,
		Name:                    server.Name,
		Hostname:                server.Hostname,
	}
}
