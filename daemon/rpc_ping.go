package daemon

import (
	"context"

	"github.com/coreos/go-semver/semver"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func (r *RPC) Ping(ctx context.Context, in *pb.Empty) (*pb.PingResponse, error) {
	versionData := r.dm.GetVersionData()
	currentVersion := semver.New(r.version)
	if versionData.newerVersionAvailable {
		return &pb.PingResponse{
			Type:     internal.CodeOutdated,
			Major:    currentVersion.Major,
			Minor:    currentVersion.Minor,
			Patch:    currentVersion.Patch,
			Metadata: currentVersion.Metadata,
		}, nil
	}

	return &pb.PingResponse{
		Type:     internal.CodeSuccess,
		Major:    currentVersion.Major,
		Minor:    currentVersion.Minor,
		Patch:    currentVersion.Patch,
		Metadata: currentVersion.Metadata,
	}, nil
}
