package daemon

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// Groups provides endpoint and autocompletion.
func (r *RPC) Groups(ctx context.Context, in *pb.Empty) (*pb.ServerGroupsList, error) {
	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.ServerGroupsList{
			Type: internal.CodeConfigError,
		}, nil
	}

	groups, err := r.dm.Groups(
		cfg.Technology,
		cfg.AutoConnectData.Protocol,
		cfg.AutoConnectData.Obfuscate,
		cfg.VirtualLocation.Get(),
	)
	if err != nil {
		log.Println(internal.ErrorPrefix, "failed to get group names", err)
		return &pb.ServerGroupsList{
			Type: internal.CodeEmptyPayloadError,
		}, nil
	}

	return &pb.ServerGroupsList{
		Type:    internal.CodeSuccess,
		Servers: groups,
	}, nil
}
