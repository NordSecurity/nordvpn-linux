package daemon

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// Cities provides cities command and autocompletion.
func (r *RPC) Cities(ctx context.Context, in *pb.CitiesRequest) (*pb.ServerGroupsList, error) {
	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.ServerGroupsList{
			Type: internal.CodeConfigError,
		}, nil
	}

	cities, err := r.dm.Cities(
		in.GetCountry(),
		cfg.Technology,
		cfg.AutoConnectData.Protocol,
		cfg.AutoConnectData.Obfuscate,
		cfg.VirtualLocation.Get(),
	)
	if err != nil {
		log.Println(internal.ErrorPrefix, "failed to get cities for", in.GetCountry(), err)

		return &pb.ServerGroupsList{
			Type: internal.CodeEmptyPayloadError,
		}, nil
	}
	return &pb.ServerGroupsList{
		Type:    internal.CodeSuccess,
		Servers: cities,
	}, nil
}
