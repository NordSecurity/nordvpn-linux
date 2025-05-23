package daemon

import (
	"log"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/tunnel"
	"google.golang.org/grpc/peer"
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
		Technology: cfg.Technology,
		Firewall:   cfg.Firewall,
		Fwmark:     cfg.FirewallMark,
		Routing:    cfg.Routing.Get(),
		Analytics:  cfg.Analytics.Get(),
		KillSwitch: cfg.KillSwitch,
		AutoConnectData: &pb.AutoconnectData{
			Enabled:     cfg.AutoConnect,
			Country:     cfg.AutoConnectData.Country,
			City:        cfg.AutoConnectData.City,
			ServerGroup: cfg.AutoConnectData.Group,
		},
		Ipv6:                 cfg.IPv6,
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
	}

	return &settings
}

// statusStream starts streaming status events received by stateChan to the subscriber. When the stream is stopped(i.e
// when subscribers stops listening), stopChan will be closed.
func statusStream(stateChan <-chan any,
	stopChan chan<- struct{},
	uid int64,
	srv pb.Daemon_SubscribeToStateChangesServer,
	requestedConnParamsStorage *RequestedConnParamsStorage,
	r *RPC,
) {
	for {
		select {
		case <-srv.Context().Done():
			close(stopChan)
			return
		case ev := <-stateChan:
			switch e := ev.(type) {
			case events.DataConnectChangeNotif:
				currentStatus := r.connectionInfo.Status()
				requestedConnParams := requestedConnParamsStorage.Get()
				status := pb.StatusResponse{
					State:           currentStatus.State,
					Ip:              currentStatus.IP.String(),
					Country:         currentStatus.Country,
					CountryCode:     currentStatus.CountryCode,
					City:            currentStatus.City,
					Name:            currentStatus.Name,
					Hostname:        currentStatus.Hostname,
					IsMeshPeer:      currentStatus.MeshnetPeer,
					ByUser:          true,
					VirtualLocation: currentStatus.VirtualLocation,
					Technology:      currentStatus.Technology,
					Protocol:        currentStatus.Protocol,
					Obfuscated:      currentStatus.Obfuscated,
					PostQuantum:     currentStatus.PostQuantum,
					Parameters: &pb.ConnectionParameters{
						ServerName:  requestedConnParams.ServerName,
						Source:      requestedConnParams.ConnectionSource,
						Country:     requestedConnParams.Country,
						City:        requestedConnParams.City,
						Group:       requestedConnParams.Group,
						CountryCode: requestedConnParams.CountryCode,
					},
				}

				if currentStatus.State == pb.ConnectionState_CONNECTED {
					transferStats, err := tunnel.GetTransferRates(currentStatus.TunnelName)
					if err != nil {
						log.Println(internal.ErrorPrefix, "failed to get transfer rates for tunnel:", err)
					}
					status.Upload = transferStats.Tx
					status.Download = transferStats.Rx
				}
				if err := srv.Send(
					&pb.AppState{State: &pb.AppState_ConnectionStatus{ConnectionStatus: &status}}); err != nil {
					log.Println(internal.ErrorPrefix, "vpn enabled failed to send state update:", err)
				}
			case events.DataDisconnect:
				if err := srv.Send(
					&pb.AppState{State: &pb.AppState_ConnectionStatus{
						ConnectionStatus: &pb.StatusResponse{
							State:      pb.ConnectionState_DISCONNECTED,
							ByUser:     e.ByUser,
							Technology: e.Technology,
							Protocol:   e.Protocol,
						},
					}}); err != nil {
					log.Println(internal.ErrorPrefix, "vpn disabled failed to send state update:", err)
				}
			case pb.LoginEventType:
				if err := srv.Send(
					&pb.AppState{State: &pb.AppState_LoginEvent{
						LoginEvent: &pb.LoginEvent{Type: e},
					}}); err != nil {
					log.Println(internal.ErrorPrefix, "login event failed to send state update:", err)
				}
			case *config.Config:
				config := configToProtobuf(e, uid)
				if err := srv.Send(
					&pb.AppState{State: &pb.AppState_SettingsChange{SettingsChange: config}}); err != nil {
					log.Println(internal.ErrorPrefix, "config change failed to send state update:", err)
				}
			case pb.UpdateEvent:
				if err := srv.Send(
					&pb.AppState{State: &pb.AppState_UpdateEvent{UpdateEvent: e}}); err != nil {
					log.Println(internal.ErrorPrefix, "update event failed to send state update:", err)
				}
			case *pb.AccountModification:
				if err := srv.Send(
					&pb.AppState{State: &pb.AppState_AccountModification{AccountModification: e}}); err != nil {
					log.Println(internal.ErrorPrefix, "account updated failed to send state update:", err)
				}
			default:
			}
		}
	}
}

func (r *RPC) SubscribeToStateChanges(_ *pb.Empty, srv pb.Daemon_SubscribeToStateChangesServer) error {
	log.Println(internal.InfoPrefix, "Received new subscription request")

	peer, ok := peer.FromContext(srv.Context())
	var uid int64
	if ok {
		cred, ok := peer.AuthInfo.(internal.UcredAuth)
		if !ok {
			return srv.Send(&pb.AppState{
				State: &pb.AppState_Error{
					Error: pb.AppStateError_FAILED_TO_GET_UID,
				},
			})
		}
		uid = int64(cred.Uid)
	}

	stateChan, stopChan := r.statePublisher.AddSubscriber()
	statusStream(stateChan, stopChan, uid, srv, &r.RequestedConnParams, r)

	return nil
}
