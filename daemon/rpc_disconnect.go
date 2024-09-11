package daemon

import (
	"log"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func (r *RPC) Disconnect(_ *pb.Empty, srv pb.Daemon_DisconnectServer) error {
	if !r.netw.IsVPNActive() {
		return srv.Send(&pb.Payload{
			Type: internal.CodeVPNNotRunning,
		})
	}

	// get user's current mfa status (here is vpn-connected and user is logged-in)
	r.checkMfaStatus()

	if err := r.netw.Stop(); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return internal.ErrUnhandled
	}

	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
	}

	r.events.Service.Disconnect.Publish(events.DataDisconnect{
		Protocol:             cfg.AutoConnectData.Protocol,
		EventStatus:          events.StatusSuccess,
		Technology:           cfg.Technology,
		ThreatProtectionLite: cfg.AutoConnectData.ThreatProtectionLite,
	})

	return srv.Send(&pb.Payload{
		Type: internal.CodeDisconnected,
	})
}
