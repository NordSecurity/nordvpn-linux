package daemon

import (
	"context"
	"errors"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/recents"
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
		r.events.Service.Disconnect.Publish(events.DataDisconnect{})
	}

	return err
}

// determineServerSelectionRule determines the server selection rule based on the provided
// parameters.
func determineServerSelectionRule(params ServerParameters) config.ServerSelectionRule {
	// defensive checks for all fields
	hasCountry := params.Country != ""
	hasCity := params.City != ""
	hasGroup := params.Group != config.ServerGroup_UNDEFINED
	hasServer := params.ServerName != ""

	switch {
	case params.Undefined():
		return config.ServerSelectionRule_RECOMMENDED

	case hasCountry && hasCity && !hasGroup && !hasServer:
		return config.ServerSelectionRule_CITY

	case hasCountry && !hasCity && !hasGroup && !hasServer:
		return config.ServerSelectionRule_COUNTRY

	case hasCountry && !hasCity && hasGroup && !hasServer:
		return config.ServerSelectionRule_COUNTRY_WITH_GROUP

	case !hasCountry && !hasCity && !hasGroup && hasServer:
		return config.ServerSelectionRule_SPECIFIC_SERVER

	case hasGroup && ((!hasServer && hasCity && hasCountry) || (hasServer && !hasCity && !hasCountry)):
		return config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP

	case !hasCountry && !hasCity && hasGroup && !hasServer:
		if _, ok := config.ServerGroup_name[int32(params.Group.Number())]; ok {
			return config.ServerSelectionRule_GROUP
		}
	}

	// Fallback for any unexpected combination
	log.Println(internal.WarningPrefix,
		"Failed to determine 'server-selection-rule':", params,
		". Defaulting to :", config.ServerSelectionRule_NONE)
	return config.ServerSelectionRule_NONE
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
	r.connectionInfo.SetInitialConnecting()

	// Set status to "Connecting" and send the connection attempt event without details
	// to inform clients about connection attempt as soon as possible so they can react.
	// The details will be filled and delivered to clients later.

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

	ip, err := server.IPv4()
	if err != nil {
		log.Println(internal.ErrorPrefix, err)
		return false, internal.ErrUnhandled
	}
	r.endpoint = network.NewIPv4Endpoint(ip)

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
	serverData := vpn.ServerData{
		IP:                subnet.Addr(),
		Hostname:          server.Hostname,
		Protocol:          cfg.AutoConnectData.Protocol,
		NordLynxPublicKey: server.NordLynxPublicKey,
		Obfuscated:        cfg.AutoConnectData.Obfuscate,
		PostQuantum:       cfg.AutoConnectData.PostquantumVpn,
		OpenVPNVersion:    server.Version(),
		NordWhisperPort:   server.NordWhisperPort,
	}

	allowlist := cfg.AutoConnectData.Allowlist

	parameters := GetServerParameters(in.GetServerTag(), in.GetServerGroup(), r.dm.GetCountryData().Countries)
	r.RequestedConnParams.Set(source, parameters)

	city := country.City.Name
	if len(server.Locations) > 0 {
		city = server.Locations[0].City.Name
	}

	event := events.DataConnect{
		Protocol:                cfg.AutoConnectData.Protocol,
		Technology:              cfg.Technology,
		ThreatProtectionLite:    cfg.AutoConnectData.ThreatProtectionLite,
		IsPostQuantum:           cfg.AutoConnectData.PostquantumVpn,
		DurationMs:              getElapsedTime(connectingStartTime),
		EventStatus:             events.StatusAttempt,
		TargetServerSelection:   determineServerSelectionRule(parameters),
		ServerFromAPI:           remote,
		IsVirtualLocation:       server.IsVirtualLocation(),
		TargetServerCity:        city,
		TargetServerCountry:     country.Name,
		TargetServerCountryCode: country.Code,
		TargetServerDomain:      server.Hostname,
		TargetServerGroup:       determineTargetServerGroup(server, parameters),
		TargetServerIP:          subnet.Addr(),
		TargetServerName:        server.Name,
	}

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
		if err := srv.Send(&pb.Payload{
			Type: t,
			Data: data,
		}); err != nil {
			log.Println(internal.ErrorPrefix, err)
		}
		return false, nil
	}

	event.EventStatus = events.StatusSuccess
	event.DurationMs = getElapsedTime(connectingStartTime)
	r.events.Service.Connect.Publish(event)

	if isRecentConnectionSupported(event.TargetServerSelection) {
		recentModel := recents.Model{
			CountryCode:        event.TargetServerCountryCode,
			Country:            event.TargetServerCountry,
			City:               event.TargetServerCity,
			SpecificServer:     strings.Split(event.TargetServerDomain, ".")[0],
			SpecificServerName: event.TargetServerName,
			Group:              parameters.Group,
			ConnectionType:     event.TargetServerSelection,
		}

		// do not add anything unrelated to connection type
		if recentModel.ConnectionType == config.ServerSelectionRule_COUNTRY ||
			recentModel.ConnectionType == config.ServerSelectionRule_COUNTRY_WITH_GROUP {
			recentModel.City = ""
		}

		r.recentVPNConnStore.Add(recentModel)
	}

	if err := srv.Send(&pb.Payload{Type: internal.CodeConnected, Data: data}); err != nil {
		log.Println(internal.ErrorPrefix, err)
	}

	return false, nil
}

// isRecentConnectionSupported returns true if server connection can be used for reconnection,
// otherwise returns false
func isRecentConnectionSupported(rule config.ServerSelectionRule) bool {
	return rule != config.ServerSelectionRule_RECOMMENDED && rule != config.ServerSelectionRule_NONE
}

// getElapsedTime calculates the time elapsed since the given start time in milliseconds.
// It ensures the returned value is at least 1 millisecond
func getElapsedTime(startTime time.Time) int {
	return max(int(time.Since(startTime).Milliseconds()), 1)
}

// determineTargetServerGroup returns the title of the server group based on the selected server and
// parameters. This function assumes parameters are already validated and contains a valid group ID.
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
