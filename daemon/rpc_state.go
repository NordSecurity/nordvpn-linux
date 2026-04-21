package daemon

import (
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/config/consent"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/log"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func configToProtobuf(cfg *config.Config, uid int64) *pb.Settings {
	ports := pb.Ports{}
	for port := range cfg.AutoConnectData.Allowlist.Ports.TCP {
		ports.Tcp = append(ports.Tcp, port)
	}
	for port := range cfg.AutoConnectData.Allowlist.Ports.UDP {
		ports.Udp = append(ports.Udp, port)
	}

	subnets := []string{}
	subnets = append(subnets, cfg.AutoConnectData.Allowlist.Subnets...)

	notifyOff := cfg.UsersData.NotifyOff[uid]
	trayOff := cfg.UsersData.TrayOff[uid]

	settings := pb.Settings{
		Technology:       cfg.Technology,
		Firewall:         cfg.Firewall,
		Fwmark:           cfg.FirewallMark,
		Routing:          cfg.Routing.Get(),
		AnalyticsConsent: consentToProtobuf(cfg.AnalyticsConsent),
		KillSwitch:       cfg.KillSwitch,
		AutoConnectData: &pb.AutoconnectData{
			Enabled:     cfg.AutoConnect,
			Country:     cfg.AutoConnectData.Country,
			City:        cfg.AutoConnectData.City,
			ServerGroup: cfg.AutoConnectData.Group,
		},
		Meshnet:              cfg.Mesh,
		Dns:                  cfg.AutoConnectData.DNS,
		ThreatProtectionLite: cfg.AutoConnectData.ThreatProtectionLite,
		Protocol:             cfg.AutoConnectData.Protocol,
		LanDiscovery:         cfg.LanDiscovery,
		Allowlist: &pb.Allowlist{
			Ports:   &ports,
			Subnets: subnets,
		},
		Obfuscate:       cfg.AutoConnectData.Obfuscate,
		VirtualLocation: cfg.VirtualLocation.Get(),
		UserSettings: &pb.UserSpecificSettings{
			Uid:    uid,
			Notify: !notifyOff,
			Tray:   !trayOff,
		},
		PostquantumVpn: cfg.AutoConnectData.PostquantumVpn,
		ArpIgnore:      cfg.ARPIgnore.Get(),
	}

	return &settings
}

func consentToProtobuf(analyticsConsent config.AnalyticsConsent) consent.ConsentMode {
	switch analyticsConsent {
	case config.ConsentDenied:
		return consent.ConsentMode_DENIED
	case config.ConsentGranted:
		return consent.ConsentMode_GRANTED
	case config.ConsentUndefined:
		return consent.ConsentMode_UNDEFINED
	}
	return consent.ConsentMode_UNDEFINED
}

// statusStream starts streaming status events received by stateChan to the subscriber. When the stream is stopped(i.e
// when subscribers stops listening), stopChan will be closed.
func statusStream(stateChan <-chan any,
	stopChan chan<- struct{},
	uid int64,
	srv pb.Daemon_SubscribeToStateChangesServer,
	requestedConnParamsStorage *RequestedConnParamsStorage,
) {
	for {
		select {
		case <-srv.Context().Done():
			log.Info("Subscription has been cancelled.")
			close(stopChan)
			return
		case ev := <-stateChan:
			switch e := ev.(type) {
			case events.DataConnectChangeNotif:
				status := pb.StatusResponse{
					State:                     e.Status.State,
					Ip:                        e.Status.IP.String(),
					Country:                   e.Status.Country,
					CountryCode:               e.Status.CountryCode,
					City:                      e.Status.City,
					Name:                      e.Status.Name,
					Hostname:                  e.Status.Hostname,
					IsMeshPeer:                e.Status.IsMeshnetPeer,
					ByUser:                    true,
					VirtualLocation:           e.Status.IsVirtualLocation,
					Technology:                e.Status.Technology,
					Protocol:                  e.Status.Protocol,
					Obfuscated:                e.Status.IsObfuscated,
					PostQuantum:               e.Status.IsPostQuantum,
					Upload:                    e.Status.Tx,
					Download:                  e.Status.Rx,
					PausedAt:                  timestamppb.New(e.Status.PausedAt),
					PauseRemainingDurationSec: e.Status.PauseRemainingTimeSec,
				}

				// for disconnected state connection parameters shall be left empty
				// otherwise e.g. GUI displays incorrect message when disconnected
				if status.State != pb.ConnectionState_DISCONNECTED {
					requestedConnParams := requestedConnParamsStorage.Get()
					status.Parameters = &pb.ConnectionParameters{
						ServerName:  requestedConnParams.ServerName,
						Source:      requestedConnParams.ConnectionSource,
						Country:     requestedConnParams.Country,
						City:        requestedConnParams.City,
						Group:       requestedConnParams.Group,
						CountryCode: requestedConnParams.CountryCode,
					}
				}

				if err := srv.Send(
					&pb.AppState{State: &pb.AppState_ConnectionStatus{ConnectionStatus: &status}}); err != nil {
					log.Error("vpn enabled failed to send state update:", err)
				}
			case pb.LoginEventType:
				if err := srv.Send(
					&pb.AppState{State: &pb.AppState_LoginEvent{
						LoginEvent: &pb.LoginEvent{Type: e},
					}}); err != nil {
					log.Error("login event failed to send state update:", err)
				}
			case *config.Config:
				config := configToProtobuf(e, uid)
				if err := srv.Send(
					&pb.AppState{State: &pb.AppState_SettingsChange{SettingsChange: config}}); err != nil {
					log.Error("config change failed to send state update:", err)
				}
			case pb.UpdateEvent:
				if err := srv.Send(
					&pb.AppState{State: &pb.AppState_UpdateEvent{UpdateEvent: e}}); err != nil {
					log.Error("update event failed to send state update:", err)
				}
			case *pb.AccountModification:
				if err := srv.Send(
					&pb.AppState{State: &pb.AppState_AccountModification{AccountModification: e}}); err != nil {
					log.Error("account updated failed to send state update:", err)
				}
			case *pb.VersionHealthStatus:
				if err := srv.Send(
					&pb.AppState{State: &pb.AppState_VersionHealth{VersionHealth: e}}); err != nil {
					log.Error("version health failed to send state update:", err)
				}
			default:
			}
		}
	}
}

func (r *RPC) SubscribeToStateChanges(_ *pb.Empty, srv pb.Daemon_SubscribeToStateChangesServer) error {
	log.Info("Received new subscription request")

	cred, err := getCallerCred(srv.Context())
	if err != nil {
		log.Error("SubscribeToStateChanges:", err)
		return srv.Send(&pb.AppState{
			State: &pb.AppState_Error{
				Error: pb.AppStateError_FAILED_TO_GET_UID,
			},
		})
	}
	uid := int64(cred.Uid)

	// Subscribe before fetching current state so no transition events are
	// missed between the snapshot and the start of the event loop.
	stateChan, stopChan := r.statePublisher.AddSubscriber()

	// Send the current state immediately so newly connected clients are
	// synchronised without waiting for the next state change event. This
	// is done synchronously on the same goroutine that will later call
	// srv.Send inside statusStream, so there is no concurrent-Send race.
	currentStatus, err := r.Status(srv.Context(), &pb.Empty{})
	if err == nil {
		if sendErr := srv.Send(&pb.AppState{
			State: &pb.AppState_ConnectionStatus{ConnectionStatus: currentStatus},
		}); sendErr != nil {
			log.Error("failed to send initial state to subscriber:", sendErr)
		}
	} else {
		log.Error("failed to fetch current status for new subscriber:", err)
	}

	statusStream(stateChan, stopChan, uid, srv, &r.RequestedConnParams)

	return nil
}
