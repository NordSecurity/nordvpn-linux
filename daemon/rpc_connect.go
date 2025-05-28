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
	return r.connectWithContext(in, srv, pb.ConnectionSource_MANUAL)
}

func (r *RPC) connectWithContext(in *pb.ConnectRequest, srv pb.Daemon_ConnectServer, source pb.ConnectionSource) error {
	var err error
	var didFail bool
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
		didFail, err = r.connect(ctx, in, srv, source)
	}) {
		return srv.Send(&pb.Payload{Type: internal.CodeNothingToDo})
	}

	// set connection status to "Disconnected"
	if didFail || err != nil {
		r.vpnEvents.Disconnected.Publish(events.DataDisconnect{})
	}

	return err
}

func determineServerSelectionRule(params ServerParameters) string {
	// defensive checks for all fields
	hasCountry := params.Country != ""
	hasCity := params.City != ""
	hasGroup := params.Group != config.ServerGroup_UNDEFINED
	hasServer := params.ServerName != ""
	hasCountryCode := params.CountryCode != ""

	switch {
	case params.Undefined():
		return config.ServerSelectionRule_RECOMMENDED.String()

	case hasCountry && hasCity && !hasGroup && !hasServer && hasCountryCode:
		return config.ServerSelectionRule_CITY.String()

	case hasCountry && !hasCity && !hasGroup && !hasServer && hasCountryCode:
		return config.ServerSelectionRule_COUNTRY.String()

	case hasCountry && !hasCity && hasGroup && !hasServer && hasCountryCode:
		return config.ServerSelectionRule_COUNTRY_WITH_GROUP.String()

	case !hasCountry && !hasCity && !hasGroup && hasServer && !hasCountryCode:
		return config.ServerSelectionRule_SPECIFIC_SERVER.String()

	case !hasCountry && !hasCity && hasGroup && hasServer && !hasCountryCode:
		return config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP.String()

	case !hasCountry && !hasCity && hasGroup && !hasServer && !hasCountryCode:
		if _, ok := config.ServerGroup_name[int32(params.Group.Number())]; ok {
			return config.ServerSelectionRule_GROUP.String()
		}
	}

	// Fallback for any unexpected combination
	log.Println(internal.WarningPrefix, "Failed to determine 'ServerSelectionRule':", params)
	return ""
}

func (r *RPC) connect(
	ctx context.Context,
	in *pb.ConnectRequest,
	srv pb.Daemon_ConnectServer,
	source pb.ConnectionSource,
) (didFail bool, retErr error) {
	if !r.ac.IsLoggedIn() {
		return false, internal.ErrNotLoggedIn
	}

	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
	}

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
		TargetServerSelection:      "",
		ServerFromAPI:              true,
		TargetServerCity:           "",
		TargetServerCountry:        "",
		TargetServerCountryCode:    "",
		TargetServerDomain:         "",
		TargetServerGroup:          "",
		TargetServerIP:             "",
		TargetServerPick:           "",
		TargetServerPickerResponse: "",
	}

	// Set status to "Connecting" and send the connection attempt event without details
	// to inform clients about connection attempt as soon as possible so they can react.
	// The details will be filled and delivered to clients later.
	r.vpnEvents.Connected.Publish(event)

	vpnExpired, err := r.ac.IsVPNExpired()
	if err != nil {
		log.Println(internal.ErrorPrefix, "checking VPN expiration: ", err)
		return true, srv.Send(&pb.Payload{Type: internal.CodeTokenRenewError})
	} else if vpnExpired {
		return true, srv.Send(&pb.Payload{Type: internal.CodeAccountExpired})
	}

	if cfg.Technology == config.Technology_NORDWHISPER && !features.NordWhisperEnabled {
		return true, srv.Send(&pb.Payload{Type: internal.CodeTechnologyDisabled})
	}

	insights := r.dm.GetInsightsData().Insights

	// Measure the time it takes to obtain recommended servers list as the connection attempt event duration
	connectingStartTime := time.Now()

	inputServerTag := internal.RemoveNonAlphanumeric(in.GetServerTag())

	log.Println(internal.DebugPrefix, "picking servers for", cfg.Technology, "technology", "input",
		in.GetServerTag(), in.GetServerGroup())

	server, remote, err := selectServer(r, &insights, cfg, inputServerTag, in.GetServerGroup())
	if err != nil {
		var errorCode *internal.ErrorWithCode
		if errors.As(err, &errorCode) {
			return true, srv.Send(&pb.Payload{Type: errorCode.Code})
		}

		return false, err
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
			return false, internal.ErrUnhandled
		}
		r.endpoint = network.NewIPv4Endpoint(ip)
	}

	subnet, err := r.endpoint.Network()
	if err != nil {
		log.Println(internal.ErrorPrefix, err)
		return false, internal.ErrUnhandled
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
		CountryCode:       country.Code,
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
	event.DurationMs = getElapsedTime(connectingStartTime)

	parameters := GetServerParameters(in.GetServerTag(), in.GetServerGroup(), r.dm.GetCountryData().Countries)
	r.RequestedConnParams.Set(source, parameters)

	event.ServerFromAPI = remote
	event.TargetServerSelection = determineServerSelectionRule(parameters)
	event.TargetServerCity = country.City.Name
	event.TargetServerCountry = country.Name
	event.TargetServerCountryCode = country.Code
	event.TargetServerDomain = server.Hostname
	event.TargetServerGroup = determineTargetServerGroup(server, parameters)
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
			event.DurationMs = getElapsedTime(connectingStartTime)
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
		event.DurationMs = getElapsedTime(connectingStartTime)
		event.Error = err
		event.EventStatus = events.StatusFailure
		t := internal.CodeFailure
		if errors.Is(err, context.Canceled) {
			t = internal.CodeDisconnected
			event.EventStatus = events.StatusCanceled
			event.Error = nil
		}
		r.events.Service.Connect.Publish(event)
		r.vpnEvents.Disconnected.Publish(events.DataDisconnect{})
		if err := srv.Send(&pb.Payload{
			Type: t,
			Data: data,
		}); err != nil {
			log.Println(internal.ErrorPrefix, err)
		}
		return false, nil
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
	event.DurationMs = getElapsedTime(connectingStartTime)
	r.events.Service.Connect.Publish(event)

	if err := srv.Send(&pb.Payload{Type: internal.CodeConnected, Data: data}); err != nil {
		log.Println(internal.ErrorPrefix, err)
	}

	return false, nil
}

// getElapsedTime calculates the time elapsed since the given start time in milliseconds.
// It ensures the returned value is at least 1 millisecond
func getElapsedTime(startTime time.Time) int {
	return max(int(time.Since(startTime).Milliseconds()), 1)
}

func determineTargetServerGroup(server *core.Server, parameters ServerParameters) string {
	findServerGroupTitle := func(gid config.ServerGroup) (string, bool) {
		index := slices.IndexFunc(server.Groups, func(g core.Group) bool { return g.ID == gid })
		if index != -1 {
			return server.Groups[index].Title, true
		}
		return "", false
	}

	if parameters.Group != config.ServerGroup_UNDEFINED {
		if title, ok := findServerGroupTitle(parameters.Group); ok {
			return title
		}
	}

	if title, ok := findServerGroupTitle(config.ServerGroup_OBFUSCATED); ok {
		return title
	}

	if title, ok := findServerGroupTitle(config.ServerGroup_STANDARD_VPN_SERVERS); ok {
		return title
	}

	return ""
}

type FactoryFunc func(config.Technology) (vpn.VPN, error)
