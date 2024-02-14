package daemon

import (
	"errors"
	"fmt"
	"log"

	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/network"
)

// Connect initiates and handles the VPN connection process
func (r *RPC) Connect(in *pb.ConnectRequest, srv pb.Daemon_ConnectServer) (retErr error) {
	if !r.ac.IsLoggedIn() {
		return internal.ErrNotLoggedIn
	}

	if r.systemInfoFunc != nil && r.networkInfoFunc != nil {
		log.Printf("PRE_CONNECT system info:\n%s\n%s\n", r.systemInfoFunc(r.version), r.networkInfoFunc())
	}

	vpnExpired, err := r.ac.IsVPNExpired()
	if err != nil {
		log.Println(internal.ErrorPrefix, "checking VPN expiration: ", err)
		return srv.Send(&pb.Payload{Type: internal.CodeTokenRenewError})
	} else if vpnExpired {
		return srv.Send(&pb.Payload{Type: internal.CodeAccountExpired})
	}

	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
	}

	insights := r.dm.GetInsightsData().Insights

	event := events.DataConnect{
		APIHostname:                r.api.Base(),
		Auto:                       false,
		Protocol:                   cfg.AutoConnectData.Protocol,
		Technology:                 cfg.Technology,
		ThreatProtectionLite:       cfg.AutoConnectData.ThreatProtectionLite,
		ResponseServersCount:       1,
		ResponseTime:               0,
		Type:                       events.ConnectAttempt,
		ServerFromAPI:              true,
		TargetServerCity:           "",
		TargetServerCountry:        "",
		TargetServerDomain:         "",
		TargetServerGroup:          "",
		TargetServerIP:             "",
		TargetServerPick:           "",
		TargetServerPickerResponse: "",
	}
	r.events.Service.Connect.Publish(event)

	defer func() {
		// send failure event if the function returns an error
		// and no connected or failure event was sent
		if retErr != nil && event.Type == events.ConnectAttempt {
			event.Type = events.ConnectFailure
			r.events.Service.Connect.Publish(event)
		}
	}()

	log.Println(internal.DebugPrefix, "picking servers for", cfg.Technology, "technology")
	server, remote, err := PickServer(
		r.serversAPI,
		r.dm.GetCountryData().Countries,
		r.dm.GetServersData().Servers,
		insights.Longitude,
		insights.Latitude,
		cfg.Technology,
		cfg.AutoConnectData.Protocol,
		cfg.AutoConnectData.Obfuscate,
		in.GetServerTag(),
		in.GetServerGroup(),
	)

	if err != nil {
		log.Println(internal.ErrorPrefix, "picking servers:", err)
		switch {
		case errors.Is(err, core.ErrUnauthorized):
			if err := r.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID)); err != nil {
				return err
			}
			return internal.ErrNotLoggedIn
		case errors.Is(err, internal.ErrTagDoesNotExist),
			errors.Is(err, internal.ErrGroupDoesNotExist),
			errors.Is(err, internal.ErrServerIsUnavailable),
			errors.Is(err, internal.ErrDoubleGroup):
			return err
		default:
			return internal.ErrUnhandled
		}
	}

	country, err := server.Locations.Country()
	if err != nil {
		log.Println(internal.ErrorPrefix, err)
	}

	if cfg.IPv6 {
		if err := r.netw.PermitIPv6(); err != nil {
			log.Println(internal.ErrorPrefix, "failed to re-enable ipv6:", err)
		}
		r.endpoint = network.DefaultEndpoint(r.endpointResolver, server.IPs())
	} else {
		ip, err := server.IPv4()
		if err != nil {
			log.Println(internal.ErrorPrefix, err)
			return internal.ErrUnhandled
		}
		r.endpoint = network.NewIPv4Endpoint(ip)
	}

	subnet, err := r.endpoint.Network()
	if err != nil {
		log.Println(internal.ErrorPrefix, err)
		return internal.ErrUnhandled
	}
	r.lastServer = server

	eventCh := make(chan ConnectEvent)

	tokenData := cfg.TokensData[cfg.AutoConnectData.ID]
	creds := vpn.Credentials{
		OpenVPNUsername:    tokenData.OpenVPNUsername,
		OpenVPNPassword:    tokenData.OpenVPNPassword,
		NordLynxPrivateKey: tokenData.NordLynxPrivateKey,
	}
	var city string
	if len(server.Locations) > 0 {
		city = server.Locations[0].City.Name
	}
	serverData := vpn.ServerData{
		IP:                subnet.Addr(),
		Hostname:          server.Hostname,
		Country:           country.Name,
		City:              city,
		Protocol:          cfg.AutoConnectData.Protocol,
		NordLynxPublicKey: server.NordLynxPublicKey,
		Obfuscated:        cfg.AutoConnectData.Obfuscate,
		OpenVPNVersion:    server.Version(),
	}

	allowlist := cfg.AutoConnectData.Allowlist
	if cfg.LanDiscovery {
		allowlist = addLANPermissions(allowlist)
	}

	event.ServerFromAPI = remote
	event.TargetServerCity = country.City.Name
	event.TargetServerCountry = country.Name
	event.TargetServerDomain = server.Hostname
	event.TargetServerIP = subnet.Addr().String()

	go Connect(
		eventCh,
		creds,
		serverData,
		allowlist,
		cfg.AutoConnectData.DNS.Or(
			r.nameservers.Get(cfg.AutoConnectData.ThreatProtectionLite, server.SupportsIPv6()),
		),
		r.netw,
	)

	data := []string{r.lastServer.Name, r.lastServer.Hostname}
	for ev := range eventCh {
		switch ev.Code {
		case internal.CodeConnected:
			// If server has at least one IPv6 address
			// regardless if IPv4 or IPv6 is used to connect
			// to the server - DO NOT DISABLE IPv6.
			if !server.SupportsIPv6() {
				if err := r.netw.DenyIPv6(); err != nil {
					log.Println(internal.ErrorPrefix, "failed to disable ipv6:", err)
				}
			}
			event.Type = events.ConnectSuccess
			r.events.Service.Connect.Publish(event)

			if err := srv.Send(&pb.Payload{Type: ev.Code, Data: data}); err != nil {
				log.Println(internal.ErrorPrefix, err)
				return internal.ErrUnhandled
			}
			r.publisher.Publish("connected to vpn")
			if r.systemInfoFunc != nil && r.networkInfoFunc != nil {
				defer func() {
					log.Printf("POST_CONNECT system info:\n%s\n", r.networkInfoFunc())
				}()
			}
			return Notify(r.cm, internal.NotificationConnected, data)
		case internal.CodeFailure:
			log.Println(internal.ErrorPrefix, ev.Message)
			r.publisher.Publish(fmt.Sprintf("failed to connect to %s", server.Hostname))
			r.publisher.Publish(ev.Message)
			event.Type = events.ConnectFailure
			r.events.Service.Connect.Publish(event)
		case internal.CodeDisconnected:
		case internal.CodeVPNNotRunning:
			// nothing to do here, because already connected to VPN
			continue
		default:
		}
		if err := srv.Send(&pb.Payload{Type: ev.Code, Data: data}); err != nil {
			log.Println(internal.ErrorPrefix, err)
			return internal.ErrUnhandled
		}
	}
	return nil
}

type FactoryFunc func(config.Technology) (vpn.VPN, error)
