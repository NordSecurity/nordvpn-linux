package daemon

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// Countries provides country command and country autocompletion.
func (r *RPC) Countries(ctx context.Context, in *pb.Empty) (*pb.ServerGroupsList, error) {
	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.ServerGroupsList{
			Type: internal.CodeConfigError,
		}, nil
	}

	countries, err := r.dm.Countries(
		cfg.Technology,
		cfg.AutoConnectData.Protocol,
		cfg.AutoConnectData.Obfuscate,
		cfg.VirtualLocation.Get(),
	)
	if err != nil {
		log.Println(internal.ErrorPrefix, "failed to get countries list", err)
		return &pb.ServerGroupsList{
			Type: internal.CodeEmptyPayloadError,
		}, nil
	}
	return &pb.ServerGroupsList{
		Type:    internal.CodeSuccess,
		Servers: countries,
	}, nil
}
