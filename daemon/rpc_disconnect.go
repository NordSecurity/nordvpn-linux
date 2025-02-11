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
		if err := r.netw.UnsetFirewall(); err != nil {
			log.Println(internal.WarningPrefix, "failed to force unset firewall on disconnect:", err)
		}
		return srv.Send(&pb.Payload{
			Type: internal.CodeVPNNotRunning,
		})
	}

	if err := r.netw.Stop(); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return internal.ErrUnhandled
	}

	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
	}

	if !cfg.Mesh && cfg.MeshPrivateKey != "" {
		err := r.cm.SaveWith(func(c config.Config) config.Config {
			c.MeshPrivateKey = ""
			return c
		})
		if err != nil {
			log.Println(internal.ErrorPrefix, "failed to clean up mesh private key:", err)
		}
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
