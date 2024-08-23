package daemon

import (
	"log"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func configToProtobuf(cfg *config.Config) *pb.GlobalSettings {
	ports := pb.Ports{}
	for port := range cfg.AutoConnectData.Allowlist.Ports.TCP {
		ports.Tcp = append(ports.Tcp, port)
	}
	for port := range cfg.AutoConnectData.Allowlist.Ports.UDP {
		ports.Udp = append(ports.Udp, port)
	}

	subnets := []string{}
	for subnet := range cfg.AutoConnectData.Allowlist.Subnets {
		subnets = append(subnets, subnet)
	}

	userSet := make(map[int64]*pb.UserSpecificSettings)
	for uid, notifyOff := range cfg.UsersData.NotifyOff {
		userSet[uid] = &pb.UserSpecificSettings{
			Uid:    uid,
			Notify: !notifyOff,
		}
	}

	for uid, trayOff := range cfg.UsersData.NotifyOff {
		if userSettings, ok := userSet[uid]; ok {
			userSettings.Tray = !trayOff
			userSet[uid] = userSettings
		} else {
			userSet[uid] = &pb.UserSpecificSettings{
				Uid:  uid,
				Tray: !trayOff,
			}
		}
	}

	usersSettings := []*pb.UserSpecificSettings{}
	for _, userSettings := range userSet {
		usersSettings = append(usersSettings, userSettings)
	}

	settings := pb.GlobalSettings{
		Settings: &pb.Settings{
			Technology: cfg.Technology,
			Firewall:   cfg.Firewall,
			Fwmark:     cfg.FirewallMark,
			Routing:    cfg.Routing.Get(),
			Analytics:  cfg.Analytics.Get(),
			KillSwitch: cfg.KillSwitch,
			AutoConnectData: &pb.AutoconnectData{
				Enabled:       cfg.AutoConnect,
				ServerTag:     cfg.AutoConnectData.ServerTag,
				ServerTagType: cfg.AutoConnectData.ServerTagType,
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
		},
		UserSpecificSettings: usersSettings,
	}

	return &settings
}

// statusStream starts streaming status events received by stateChan to the subscriber. When the stream is stopped(i.e
// when subscribers stops listening), stopChan will be closed.
func statusStream(stateChan <-chan interface{}, stopChan chan<- struct{}, srv pb.Daemon_SubscribeToStateChangesServer) {
	for {
		select {
		case <-srv.Context().Done():
			close(stopChan)
			return
		case ev := <-stateChan:
			switch e := ev.(type) {
			case events.DataConnect:
				state := pb.ConnectionState_CONNECTING
				if e.EventStatus == events.StatusSuccess {
					state = pb.ConnectionState_CONNECTED
				}

				status := pb.ConnectionStatus{
					State:          state,
					ServerIp:       e.TargetServerIP,
					ServerCountry:  e.TargetServerCountry,
					ServerCity:     e.TargetServerCity,
					ServerName:     e.TargetServerName,
					ServerHostname: e.TargetServerDomain,
					IsMeshPeer:     e.IsMeshnetPeer,
				}
				if err := srv.Send(
					&pb.AppState{State: &pb.AppState_ConnectionStatus{ConnectionStatus: &status}}); err != nil {
					log.Println(internal.ErrorPrefix, "vpn enabled failed to send state update:", err)
				}
			case events.DataDisconnect:
				if err := srv.Send(
					&pb.AppState{State: &pb.AppState_ConnectionStatus{
						ConnectionStatus: &pb.ConnectionStatus{State: pb.ConnectionState_DISCONNECTED}}}); err != nil {
					log.Println(internal.ErrorPrefix, "vpn disabled failed to send state update:", err)
				}
			case pb.LoginEventType:
				if err := srv.Send(
					&pb.AppState{State: &pb.AppState_LoginEvent{
						LoginEvent: &pb.LoginEvent{Type: e}}}); err != nil {
					log.Println(internal.ErrorPrefix, "login event failed to send state update:", err)
				}
			case *config.Config:
				config := configToProtobuf(e)
				if err := srv.Send(
					&pb.AppState{State: &pb.AppState_SettingsChange{SettingsChange: config}}); err != nil {
					log.Println(internal.ErrorPrefix, "config change to send state update:", err)
				}
			default:
			}
		}
	}
}

func (r *RPC) SubscribeToStateChanges(_ *pb.Empty, srv pb.Daemon_SubscribeToStateChangesServer) error {
	log.Println(internal.InfoPrefix, "Received new subscription request")

	stateChan, stopChan := r.statePublisher.AddSubscriber()
	statusStream(stateChan, stopChan, srv)

	return nil
}
