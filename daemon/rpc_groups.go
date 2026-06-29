package daemon

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/config/remote"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
)

// Groups provides endpoint and autocompletion.
func (r *RPC) Groups(ctx context.Context, in *pb.Empty) (*pb.ServerGroupsList, error) {
	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Error(err)
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
		log.Error("failed to get group names", err)
		return &pb.ServerGroupsList{
			Type: internal.CodeEmptyPayloadError,
		}, nil
	}

	if r.remoteConfigGetter.IsFeatureEnabled(remote.FeatureDedicatedServer) {
		// Dedicated Server is to be always present in Tray
		groups = append(groups, &pb.ServerGroup{Name: internal.Title(dedicatedServersGroupTitle), VirtualLocation: false})
	}

	return &pb.ServerGroupsList{
		Type:    internal.CodeSuccess,
		Servers: groups,
	}, nil
}
