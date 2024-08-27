package daemon

import (
	"errors"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/network"
)

func isDedicatedIP(server core.Server) bool {
	index := slices.IndexFunc(server.Groups, func(group core.Group) bool {
		return group.ID == config.ServerGroup_DEDICATED_IP
	})

	return index != -1
}

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

	// Measure the time it takes to obtain recommended servers list as the connection attempt event duration
	connectingStartTime := time.Now()

	event := events.DataConnect{
		APIHostname:                r.api.Base(),
		Auto:                       false,
		Protocol:                   cfg.AutoConnectData.Protocol,
		Technology:                 cfg.Technology,
		ThreatProtectionLite:       cfg.AutoConnectData.ThreatProtectionLite,
		ResponseServersCount:       1,
		ResponseTime:               0,
		DurationMs:                 -1,
		EventStatus:                events.StatusAttempt,
		ServerFromAPI:              true,
		TargetServerCity:           "",
		TargetServerCountry:        "",
		TargetServerDomain:         "",
		TargetServerGroup:          "",
		TargetServerIP:             "",
		TargetServerPick:           "",
		TargetServerPickerResponse: "",
	}

	inputServerTag := internal.RemoveNonAlphanumeric(in.GetServerTag())

	log.Println(internal.DebugPrefix, "picking servers for", cfg.Technology, "technology", "input",
		in.GetServerTag(), in.GetServerGroup())

	server, remote, err := selectServer(r, &insights, cfg, inputServerTag, in.GetServerGroup())
	if err != nil {
		var errorCode *internal.ErrorWithCode
		if errors.As(err, &errorCode) {
			return srv.Send(&pb.Payload{Type: errorCode.Code})
		}

		return err
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
	r.lastServer = *server

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
		Name:              server.Name,
		Country:           country.Name,
		City:              city,
		Protocol:          cfg.AutoConnectData.Protocol,
		NordLynxPublicKey: server.NordLynxPublicKey,
		Obfuscated:        cfg.AutoConnectData.Obfuscate,
		OpenVPNVersion:    server.Version(),
		VirtualLocation:   server.IsVirtualLocation(),
	}

	allowlist := cfg.AutoConnectData.Allowlist

	event.ServerFromAPI = remote
	event.TargetServerCity = country.City.Name
	event.TargetServerCountry = country.Name
	event.TargetServerDomain = server.Hostname
	event.TargetServerIP = subnet.Addr().String()
	event.DurationMs = max(int(time.Since(connectingStartTime).Milliseconds()), 1)

	// Send the connection attempt event
	r.events.Service.Connect.Publish(event)

	// Reset the connecting start timer, as the connect success and failure events should not include time taken
	// for getting the recommended servers, which was already reported as the attempt event duration.
	connectingStartTime = time.Now()

	defer func() {
		// Send connect failure event if this function will return an error
		// and no connect success or connect failure event was sent.
		if retErr != nil && event.EventStatus == events.StatusAttempt {
			event.EventStatus = events.StatusFailure
			event.DurationMs = max(int(time.Since(connectingStartTime).Milliseconds()), 1)
			r.events.Service.Connect.Publish(event)
		}
	}()

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

	virtualServer := ""
	if server.IsVirtualLocation() {
		virtualServer = " - Virtual"
	}
	data := []string{r.lastServer.Name, r.lastServer.Hostname, virtualServer}

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
			event.EventStatus = events.StatusSuccess
			event.DurationMs = max(int(time.Since(connectingStartTime).Milliseconds()), 1)
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
			return nil
		case internal.CodeFailure:
			log.Println(internal.ErrorPrefix, ev.Message)
			r.publisher.Publish(fmt.Sprintf("failed to connect to %s", server.Hostname))
			r.publisher.Publish(ev.Message)
			event.EventStatus = events.StatusFailure
			event.DurationMs = max(int(time.Since(connectingStartTime).Milliseconds()), 1)
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
