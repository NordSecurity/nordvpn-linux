package daemon

import (
	"fmt"
	"log"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func (r *RPC) Disconnect(_ *pb.Empty, srv pb.Daemon_DisconnectServer) error {
	wasConnected, err := r.DoDisconnect()
	if err != nil {
		log.Println(internal.ErrorPrefix, err)
		return internal.ErrUnhandled
	}
	if !wasConnected {
		return srv.Send(&pb.Payload{
			Type: internal.CodeVPNNotRunning,
		})
	}
	return srv.Send(&pb.Payload{Type: internal.CodeDisconnected})
}

// DoDisconnect is the non-gRPC function for Disconect to be used dirrectly.
func (r *RPC) DoDisconnect() (bool, error) {
	startTime := time.Now()
	if !r.netw.IsVPNActive() {
		if err := r.netw.UnsetFirewall(); err != nil {
			log.Println(internal.WarningPrefix, "failed to force unset firewall on disconnect:", err)
		}
		return false, nil
	}

	var cfg config.Config
	var err error
	defer func() {
		status := events.StatusSuccess
		if err != nil {
			status = events.StatusFailure
		}
		r.events.Service.Disconnect.Publish(events.DataDisconnect{
			Protocol:             cfg.AutoConnectData.Protocol,
			EventStatus:          status,
			Technology:           cfg.Technology,
			ThreatProtectionLite: cfg.AutoConnectData.ThreatProtectionLite,
			Duration:             time.Since(startTime),
			Error:                err,
		})
	}()
	if err = r.netw.Stop(); err != nil {
		err = fmt.Errorf("stopping networker: %w", err)
		return true, err
	}
	if err = r.cm.Load(&cfg); err != nil {
		err = fmt.Errorf("loading config: %w", err)
		return true, err
	}

	return true, nil
}
