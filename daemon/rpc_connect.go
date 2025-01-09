package daemon

import (
	"context"
	"errors"
	"log"
	"slices"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/features"
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
	var err error
	// TODO: Currently this only listens to a given context in `netw.Start()`, therefore gets
	// stopped on `ctx.Done()` only if it happens while `netw.Start()` is being executed.
	// Otherwise:
	//   * if context is done before `netw.Start()` is called, it will wait until
	//     `netw.Start()` is called and exit immediately with an error from the `ctx`.
	//   * if context is done after `netw.Start()` is done, it will ignore the event and resume
	//     whole `r.connect` until it exits.
	// In order to fix this, all of expensive operations should implement `ctx.Done()` handling
	// and have context bypassed to them.
	if !r.connectContext.TryExecuteWith(func(ctx context.Context) {
		err = r.connect(ctx, in, srv)
	}) {
		return srv.Send(&pb.Payload{Type: internal.CodeNothingToDo})
	}
	return err
}

func (r *RPC) connect(
	ctx context.Context,
	in *pb.ConnectRequest,
	srv pb.Daemon_ConnectServer,
) (retErr error) {
	if !r.ac.IsLoggedIn() {
		return internal.ErrNotLoggedIn
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

	if cfg.Technology == config.Technology_NORDWHISPER {
		if !features.NordWhisperEnabled {
			return srv.Send(&pb.Payload{Type: internal.CodeTechnologyDisabled})
		}

		nordWhisperEnabled, err := r.remoteConfigGetter.GetNordWhisperEnabled(r.version)
		if err != nil {
			log.Println(internal.ErrorPrefix, "failed to retrieve remote config for NordWhisper:", err)
			return srv.Send(&pb.Payload{Type: internal.CodeTechnologyDisabled})
		}

		if !nordWhisperEnabled {
			return srv.Send(&pb.Payload{Type: internal.CodeTechnologyDisabled})
		}
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
		PostQuantum:       cfg.AutoConnectData.PostquantumVpn,
		NordWhisperPort:   server.NordWhisperPort,
	}

	allowlist := cfg.AutoConnectData.Allowlist

	event.ServerFromAPI = remote
	event.TargetServerCity = country.City.Name
	event.TargetServerCountry = country.Name
	event.TargetServerDomain = server.Hostname
	event.TargetServerIP = subnet.Addr().String()
	event.DurationMs = max(int(time.Since(connectingStartTime).Milliseconds()), 1)

	parameters := GetServerParameters(in.GetServerTag(), in.GetServerGroup(), r.dm.GetCountryData().Countries)
	r.ConnectionParameters.SetConnectionParameters(pb.ConnectionSource_MANUAL, parameters)

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

	virtualServer := ""
	if server.IsVirtualLocation() {
		virtualServer = " - Virtual"
	}
	data := []string{r.lastServer.Name, r.lastServer.Hostname, virtualServer}

	if err := srv.Send(&pb.Payload{Type: internal.CodeConnecting, Data: data}); err != nil {
		log.Println(internal.ErrorPrefix, err)
	}

	err = r.netw.Start(
		ctx,
		creds,
		serverData,
		allowlist,
		cfg.AutoConnectData.DNS.Or(r.nameservers.Get(
			cfg.AutoConnectData.ThreatProtectionLite,
			server.SupportsIPv6(),
		)),
		true, // here vpn connect - enable routing to local LAN
	)
	if err != nil {
		event.DurationMs = max(int(time.Since(connectingStartTime).Milliseconds()), 1)
		event.Error = err
		event.EventStatus = events.StatusFailure
		t := internal.CodeFailure
		if errors.Is(err, context.Canceled) {
			t = internal.CodeDisconnected
			event.EventStatus = events.StatusCanceled
			event.Error = nil
		}
		r.events.Service.Connect.Publish(event)
		if err := srv.Send(&pb.Payload{
			Type: t,
			Data: data,
		}); err != nil {
			log.Println(internal.ErrorPrefix, err)
		}
		return nil
	}

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

	if err := srv.Send(&pb.Payload{Type: internal.CodeConnected, Data: data}); err != nil {
		log.Println(internal.ErrorPrefix, err)
	}

	return nil
}

type FactoryFunc func(config.Technology) (vpn.VPN, error)
