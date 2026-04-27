package daemon

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/daemon/access"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
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
		DeviceKeyInvalidator:         r.dedicatedServersKeyManager,
	})

	if err := r.recentVPNConnStore.Clean(); err != nil {
		log.Printf("%s [rpc] failed to clean recent connections on logout: %v\n", internal.WarningPrefix, err)
	}

	if result.Status == 0 {
		return nil, result.Err
	}

	return &pb.Payload{Type: result.Status}, result.Err
}
