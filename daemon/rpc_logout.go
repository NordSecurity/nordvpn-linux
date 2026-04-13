package daemon

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/daemon/access"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/log"
)

func (r *RPC) Logout(ctx context.Context, in *pb.LogoutRequest) (*pb.Payload, error) {
	if r.connectionInfo.IsPaused() {
		r.CancelPause()
	}
	result := access.Logout(access.LogoutInput{
		AuthChecker:                  r.ac,
		CredentialsAPI:               r.credentialsAPI,
		Netw:                         r.netw,
		NcClient:                     r.ncClient,
		ConfigManager:                r.cm,
		UserLogoutEventPublisherFunc: r.events.User.Logout.Publish,
		DebugPublisherFunc:           r.publisher.Publish,
		PersistToken:                 in.GetPersistToken(),
		DisconnectFunc:               r.DoDisconnect,
	})

	if err := r.recentVPNConnStore.Clean(); err != nil {
		log.Warnf("[rpc] failed to clean recent connections on logout: %v\n", err)
	}

	if result.Status == 0 {
		return nil, result.Err
	}

	return &pb.Payload{Type: result.Status}, result.Err
}
