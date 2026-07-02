package daemon

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/config/remote"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	devicekey "github.com/NordSecurity/nordvpn-linux/device_key"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/features"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
	"github.com/NordSecurity/nordvpn-linux/network"
)

func isDedicatedIP(server core.Server) bool {
	index := slices.IndexFunc(server.Groups, func(group core.Group) bool {
		return group.ID == config.ServerGroup_DEDICATED_IP
	})

	return index != -1
}

// determineServerGroupIDs returns the IDs of every group the server belongs to.
func determineServerGroupIDs(server *core.Server) []config.ServerGroup {
	ids := make([]config.ServerGroup, 0, len(server.Groups))
	for _, g := range server.Groups {
		ids = append(ids, g.ID)
	}
	return ids
}

// isDedicatedServer returns true if either serverTag or serverGroup represents the dedicated server group
func isDedicatedServer(serverTag string, serverGroup string) bool {
	return groupConvert(serverTag) == config.ServerGroup_DEDICATED_SERVER ||
		groupConvert(serverGroup) == config.ServerGroup_DEDICATED_SERVER
}

// Connect initiates and handles the VPN connection process
func (r *RPC) Connect(in *pb.ConnectRequest, srv pb.Daemon_ConnectServer) (retErr error) {
	return r.connectFromRequest(in, srv, pb.ConnectionSource_MANUAL)
}

func (r *RPC) connectFromRequest(in *pb.ConnectRequest, srv pb.Daemon_ConnectServer, source pb.ConnectionSource) error {
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
		didFail, err = r.connectWithParameters(ctx, in, srv, source)
	}) {
		return srv.Send(&pb.Payload{Type: internal.CodeNothingToDo})
	}

	// reconcile connection status after a failed attempt
	if didFail || err != nil {
		r.reconcileStatusAfterFailedConnect()
	}

	return err
}

// reconcileStatusAfterFailedConnect fixes up the reported connection status after a connection
// attempt failed. If the attempt aborted before the existing tunnel was touched, the VPN is still
// up (netw.IsVPNActive()==true) and we restore the previous status instead of falsely reporting a
// disconnect. Only when there is genuinely no active tunnel do we publish a Disconnect.
func (r *RPC) reconcileStatusAfterFailedConnect() {
	if r.netw.IsVPNActive() {
		r.connectionInfo.RestorePreviousStatus()
		return
	}
	r.events.Service.Disconnect.Publish(events.DataDisconnect{})
}

func (r *RPC) connectFromLastSelection(srv pb.Daemon_ConnectServer,
	source pb.ConnectionSource,
	pauseDuration time.Duration) error {
	var err error
	var didFail bool
	if !r.connectContext.TryExecuteWith(func(ctx context.Context) {
		didFail, err = r.connectWithStoredServerSelection(ctx, srv, pauseDuration)
	}) {
		return srv.Send(&pb.Payload{Type: internal.CodeNothingToDo})
	}

	// reconcile connection status after a failed attempt
	if didFail || err != nil {
		r.reconcileStatusAfterFailedConnect()
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
	log.Warn("Failed to determine 'server-selection-rule':", params,
		". Defaulting to :", config.ServerSelectionRule_NONE)
	return config.ServerSelectionRule_NONE
}

func (r *RPC) isVPNExpired() int64 {
	vpnExpired, err := r.ac.IsVPNExpired()
	if err != nil {
		log.Error("checking VPN expiration: ", err)
		return internal.CodeTokenRenewError
	} else if vpnExpired {
		return internal.CodeAccountExpired
	}
	return internal.CodeSuccess
}

func (r *RPC) connectWithStoredServerSelection(ctx context.Context,
	srv pb.Daemon_ConnectServer,
	pauseDuration time.Duration) (bool, error) {
	if ok, err := r.ac.IsLoggedIn(); !ok {
		if errors.Is(err, core.ErrUnauthorized) {
			_ = srv.Send(&pb.Payload{Type: internal.CodeRevokedAccessToken})
		}
		return false, internal.ErrNotLoggedIn
	}

	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Error(err)
		return false, fmt.Errorf("reading config: %w", err)
	}
	r.connectionInfo.SetInitialConnecting()

	if IsServerDedicated(*r.lastServerSelection.server) {
		// first, check if feature is enabled at all
		if !r.remoteConfigGetter.IsFeatureEnabled(remote.FeatureDedicatedServer) {
			// if user is trying to connect here while this feature is disabled,
			// show general error because anyways he should not get here
			return true, srv.Send(&pb.Payload{Type: internal.CodeFailure})
		}
		// second, if feature is enabled, check if technology is correct
		if cfg.Technology != config.Technology_NORDLYNX {
			return true, srv.Send(&pb.Payload{Type: internal.CodeDedicatedServersNoNordlynx})
		}
		// third, if technology is correct, check if post quantum is enabled(pq is not supported for dedicated servers)
		if cfg.AutoConnectData.PostquantumVpn {
			return true, srv.Send(&pb.Payload{Type: internal.CodeDedicatedServersPq})
		}
	}

	expirationCheckResult := r.isVPNExpired()
	if expirationCheckResult != internal.CodeSuccess {
		return true, srv.Send(&pb.Payload{Type: expirationCheckResult})
	}

	return r.connect(ctx,
		srv,
		cfg,
		r.lastServerSelection,
		r.RequestedConnParams.Get().ServerParameters,
		time.Now(),
		false,
		pauseDuration)
}

func (r *RPC) connectWithParameters(ctx context.Context,
	in *pb.ConnectRequest,
	srv pb.Daemon_ConnectServer,
	source pb.ConnectionSource,
) (didFail bool, retErr error) {
	pauseDuration := r.pauseManager.CancelReconnection()
	if ok, err := r.ac.IsLoggedIn(); !ok {
		if errors.Is(err, core.ErrUnauthorized) {
			_ = srv.Send(&pb.Payload{Type: internal.CodeRevokedAccessToken})
		}
		return false, internal.ErrNotLoggedIn
	}

	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Error(err)
	}
	prelimGroup := groupConvert(in.GetServerTag())
	if prelimGroup == config.ServerGroup_UNDEFINED {
		prelimGroup = groupConvert(in.GetServerGroup())
	}
	r.RequestedConnParams.Set(source, ServerParameters{Group: prelimGroup})
	r.connectionInfo.SetInitialConnecting()

	if isDedicatedServer(in.ServerTag, in.ServerGroup) {
		// first, check if feature is enabled at all
		if !r.remoteConfigGetter.IsFeatureEnabled(remote.FeatureDedicatedServer) {
			// if user is trying to connect here while this feature is disabled,
			// show general error because anyways he should not get here
			return true, srv.Send(&pb.Payload{Type: internal.CodeGroupNonexisting})
		}
		// second, if feature is enabled, check if technology is correct
		if cfg.Technology != config.Technology_NORDLYNX {
			return true, srv.Send(&pb.Payload{Type: internal.CodeDedicatedServersNoNordlynx})
		}
		// third, if technology is correct, check if post quantum is enabled(pq is not supported for dedicated servers)
		if cfg.AutoConnectData.PostquantumVpn {
			return true, srv.Send(&pb.Payload{Type: internal.CodeDedicatedServersPq})
		}
	}

	// Set status to "Connecting" and send the connection attempt event without details
	// to inform clients about connection attempt as soon as possible so they can react.
	// The details will be filled and delivered to clients later.

	expirationCheckResult := r.isVPNExpired()
	if expirationCheckResult != internal.CodeSuccess {
		return true, srv.Send(&pb.Payload{Type: expirationCheckResult})
	}

	if cfg.Technology == config.Technology_NORDWHISPER && !features.NordWhisperEnabled {
		return true, srv.Send(&pb.Payload{Type: internal.CodeTechnologyDisabled})
	}

	insights := r.dm.GetInsightsData().Insights

	// Measure the time it takes to obtain recommended servers list as the connection attempt event duration
	connectingStartTime := time.Now()

	inputServerTag := internal.RemoveNonAlphanumeric(in.GetServerTag())

	log.Debug("picking servers for", cfg.Technology, "technology", "input",
		in.GetServerTag(), in.GetServerGroup())

	serverSelection, err := selectServer(r, &insights, cfg, inputServerTag, in.GetServerGroup())

	if err != nil {
		var errorCode *internal.ErrorWithCode
		if errors.As(err, &errorCode) {
			if errorCode.Code == internal.CodeDedicatedServersNotReady {
				r.publishDedicatedServerStatus(serverSelection.dedicatedServerStatus)
			}
			return true, srv.Send(&pb.Payload{Type: errorCode.Code})
		}

		if errors.Is(err, core.ErrUnauthorized) {
			return true, srv.Send(&pb.Payload{Type: internal.CodeRevokedAccessToken})
		}

		if errors.Is(err, internal.ErrServerIsUnavailable) {
			return true, srv.Send(&pb.Payload{Type: internal.CodeServerUnavailable})
		}

		if errors.Is(err, internal.ErrVirtualServerSelected) {
			return true, srv.Send(&pb.Payload{Type: internal.CodeVirtualLocationDisabled})
		}

		return false, err
	}
	r.lastServerSelection = serverSelection

	parameters := GetServerParameters(in.GetServerTag(), in.GetServerGroup(), r.dm.GetCountryData().Countries)
	r.RequestedConnParams.Set(source, parameters)

	return r.connect(ctx, srv, cfg, serverSelection, parameters, connectingStartTime, true, pauseDuration)
}

func (r *RPC) connect(
	ctx context.Context,
	srv pb.Daemon_ConnectServer,
	cfg config.Config,
	serverSelection serverSelection,
	parameters ServerParameters,
	connectingStartTime time.Time,
	pauseInterrupted bool,
	pauseDuration time.Duration,
) (didFail bool, retErr error) {
	country, err := serverSelection.server.Locations.Country()
	if err != nil {
		log.Error(err)
	}

	tokenData := cfg.TokensData[cfg.AutoConnectData.ID]
	creds := vpn.Credentials{
		OpenVPNUsername:    tokenData.OpenVPNUsername,
		OpenVPNPassword:    tokenData.OpenVPNPassword,
		NordLynxPrivateKey: tokenData.NordLynxPrivateKey,
	}

	isServerDedicated := IsServerDedicated(*serverSelection.server)
	// if server is a dedicated server, we need to use the device key instead of NordLynx private key
	if isServerDedicated {
		dedicatedServersDeviceData := r.dedicatedServerKeyManager.CheckAndRegisterDedicatedServers()
		if dedicatedServersDeviceData == nil {
			log.Error("failed to fetch the device key for dedicated server connection")
			return false, internal.ErrUnhandled
		}

		dedicatedServerConnectionData, err := getDedicatedServerConnectionData(
			r.dedicatedServersAPI,
			serverSelection.server.DedicatedServerUUID,
			*dedicatedServersDeviceData)
		if errors.Is(err, core.ErrDedicatedServersDeviceNotFound) {
			dedicatedServersDeviceData = r.dedicatedServerKeyManager.ForceRegisterDedicatedServers()
			if dedicatedServersDeviceData == nil {
				log.Error("failed to force dedicated server device registration")
				return true, srv.Send(&pb.Payload{Type: internal.CodeDedicatedServersCanNotConnect})
			}
			dedicatedServerConnectionData, err = getDedicatedServerConnectionData(
				r.dedicatedServersAPI,
				serverSelection.server.DedicatedServerUUID,
				*dedicatedServersDeviceData)
		}

		if err != nil {
			log.Error("fetching dedicated server connection data:", err)
			switch {
			case errors.Is(err, core.ErrDedicatedServersSessionMaxLimitReached):
				return true, srv.Send(&pb.Payload{Type: internal.CodeDedicatedServersSessionMaxLimitReached})
			case errors.Is(err, core.ErrDedicatedServersDeviceNotFound),
				errors.Is(err, core.ErrDedicatedServersDeviceNotRegistered),
				errors.Is(err, core.ErrDedicatedServersPublicKeyMismatch),
				errors.Is(err, core.ErrDedicatedServersServerOffline),
				errors.Is(err, core.ErrDedicatedServersServerNotFound),
				errors.Is(err, core.ErrDedicatedServersInvalidFormData):
				return true, srv.Send(&pb.Payload{Type: internal.CodeDedicatedServersCanNotConnect})
			}
			return false, internal.ErrUnhandled
		}

		creds.NordLynxPrivateKey = dedicatedServersDeviceData.DevicePrivateKey
		serverSelection.server.Station = dedicatedServerConnectionData.ip
		serverSelection.server.DedicatedServersPort = dedicatedServerConnectionData.port
		serverSelection.server.NordLynxPublicKey = dedicatedServerConnectionData.publicKey

		r.publishDedicatedServerStatus(serverSelection.dedicatedServerStatus)
	}

	ip, err := serverSelection.server.IPv4()
	if err != nil {
		log.Error(err)
		return false, internal.ErrUnhandled
	}
	r.endpoint = network.NewIPv4Endpoint(ip)

	subnet, err := r.endpoint.Network()
	if err != nil {
		log.Error(err)
		return false, internal.ErrUnhandled
	}

	serverData := vpn.ServerData{
		IP:                  subnet.Addr(),
		Hostname:            serverSelection.server.Hostname,
		Protocol:            cfg.AutoConnectData.Protocol,
		NordLynxPublicKey:   serverSelection.server.NordLynxPublicKey,
		Obfuscated:          cfg.AutoConnectData.Obfuscate,
		PostQuantum:         cfg.AutoConnectData.PostquantumVpn,
		OpenVPNVersion:      serverSelection.server.Version(),
		NordWhisperPort:     serverSelection.server.NordWhisperPort,
		DedicatedServerPort: serverSelection.server.DedicatedServersPort,
	}

	allowlist := cfg.AutoConnectData.Allowlist

	city := country.City.Name
	if len(serverSelection.server.Locations) > 0 {
		city = serverSelection.server.Locations[0].City.Name
	}

	serverSelectionRule := determineServerSelectionRule(parameters)
	r.connectionInfo.SetServerSelectionData(serverSelectionRule, serverSelection.remote)

	event := events.DataConnect{
		Protocol:                cfg.AutoConnectData.Protocol,
		Technology:              cfg.Technology,
		ThreatProtectionLite:    cfg.AutoConnectData.ThreatProtectionLite,
		IsObfuscated:            cfg.AutoConnectData.Obfuscate,
		IsPostQuantum:           cfg.AutoConnectData.PostquantumVpn,
		DurationMs:              getElapsedTime(connectingStartTime),
		EventStatus:             events.StatusAttempt,
		TargetServerSelection:   determineServerSelectionRule(parameters),
		ServerFromAPI:           serverSelection.remote,
		IsVirtualLocation:       serverSelection.server.IsVirtualLocation(),
		TargetServerCity:        city,
		TargetServerCountry:     country.Name,
		TargetServerCountryCode: country.Code,
		TargetServerDomain:      serverSelection.server.Hostname,
		TargetServerGroup:       determineTargetServerGroup(serverSelection.server, parameters),
		ServerGroups:            determineServerGroupIDs(serverSelection.server),
		TargetServerIP:          subnet.Addr(),
		TargetServerName:        serverSelection.server.Name,
		RecommendationUUID:      string(serverSelection.recommendationUUID),
		PauseInterval:           pauseDuration,
		UnpausedByUser:          pauseInterrupted,
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
	if serverSelection.server.IsVirtualLocation() {
		virtualServer = " - Virtual"
	}
	lastServer := r.lastServerSelection.server

	data := []string{lastServer.Name, lastServer.Hostname, virtualServer}
	// In case of dedicated servers we only return server name, as hostname is not available.
	if isServerDedicated {
		data = []string{lastServer.Name}
	}

	if err := srv.Send(&pb.Payload{Type: internal.CodeConnecting, Data: data}); err != nil {
		log.Error(err)
	}

	disconnectSender := events.NewDisconnectSender(events.DataDisconnect{
		Protocol:             cfg.AutoConnectData.Protocol,
		Technology:           cfg.Technology,
		ThreatProtectionLite: cfg.AutoConnectData.ThreatProtectionLite,
		RecommendationUUID:   string(serverSelection.recommendationUUID),
	}, r.events.Service.Disconnect.Publish)

	err = r.netw.Start(
		ctx,
		creds,
		serverData,
		allowlist,
		cfg.AutoConnectData.DNS.Or(r.nameservers.Get(
			cfg.AutoConnectData.ThreatProtectionLite,
		)),
		true, // here vpn connect - enable routing to local LAN
		disconnectSender.PublishDisconnect,
	)

	defer func() {
		storePendingRecentConnection(r.recentVPNConnStore)
		connectionEstablished := event.EventStatus == events.StatusSuccess
		if connectionEstablished && isRecentConnectionSupported(event.TargetServerSelection) {
			recentModel, err := buildRecentConnectionModel(event, parameters, serverSelection.server, r.dm, cfg)
			if err != nil {
				log.Warn("Failed to build recent VPN connection model:", err)
				return
			}
			r.recentVPNConnStore.AddPending(recentModel)
		}
	}()

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
			log.Error(err)
		}
		return false, nil
	}

	event.EventStatus = events.StatusSuccess
	event.DurationMs = getElapsedTime(connectingStartTime)

	r.events.Service.Connect.Publish(event)
	// The "first time used" event correlates to the device location, which should already be known once a VPN connection is established.
	// Hence, this event is emitted here.
	r.events.Service.FirstTimeOpened.Publish(struct{}{})

	if err := srv.Send(&pb.Payload{Type: internal.CodeConnected, Data: data}); err != nil {
		log.Error(err)
	}

	return false, nil
}

func (r *RPC) publishDedicatedServerStatus(status core.DedicatedServerStatus) {
	if status != "" {
		r.events.Service.DedicatedServerStatus.Publish(
			events.DataDedicatedServerStatus{Status: string(status)},
		)
	}
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

type dedicatedServerConnectionData struct {
	ip        string
	port      int64
	publicKey string
}

func getDedicatedServerConnectionData(api core.DedicatedServersAPI,
	serverUUID string,
	deviceConnectionData devicekey.DedicatedServersConnectionData) (dedicatedServerConnectionData, error) {
	connectResponse, err := api.DedicatedServerConnectCheck(serverUUID, core.DedicatedServerConnectRequest{
		DeviceUUID:      deviceConnectionData.DeviceUUID.String(),
		DevicePublicKey: deviceConnectionData.DevicePublicKey,
	})
	if err != nil {
		return dedicatedServerConnectionData{}, fmt.Errorf("getting dedicated server connection data: %w", err)
	}

	addrPort := strings.Split(connectResponse.ServerEndpoint, ":")

	ip := addrPort[0]

	var dedicatedServerPort int64
	if len(addrPort) > 1 {
		port, err := strconv.Atoi(addrPort[1])
		if err != nil {
			log.Error("parsing dedicated server port:", err)
			return dedicatedServerConnectionData{}, fmt.Errorf("parsing dedicated server port: %w", err)
		}
		dedicatedServerPort = int64(port)
	}

	return dedicatedServerConnectionData{
		ip:        ip,
		port:      dedicatedServerPort,
		publicKey: connectResponse.ServerPublicKey,
	}, nil
}

type FactoryFunc func(config.Technology) (vpn.VPN, error)
