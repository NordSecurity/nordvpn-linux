package daemon

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/daemon/access"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
)

func (r *RPC) Logout(ctx context.Context, in *pb.LogoutRequest) (*pb.Payload, error) {
	result := access.Logout(access.LogoutInput{
		AuthChecker:    r.ac,
		CredentialsAPI: r.credentialsAPI,
		Netw:           r.netw,
		NcClient:       r.ncClient,
		ConfigManager:  r.cm,
		Events:         r.events,
		Publisher:      r.publisher,
		PersistToken:   in.GetPersistToken(),
		DisconnectAll:  r.DoDisconnect,
	})

	if result.Status == 0 {
		return nil, result.Err
	}

	return &pb.Payload{Type: result.Status}, result.Err
}
