package daemon

import (
	"log"

	"github.com/NordSecurity/nordvpn-linux/daemon/access"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
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

// DoDisconnect is the non-gRPC function for Disconect to be used directly.
func (r *RPC) DoDisconnect() (bool, error) {
	return access.Disconnect(access.DisconnectInput{
		Networker:     r.netw,
		ConfigManager: r.cm,
		Events:        r.events,
	})
}
