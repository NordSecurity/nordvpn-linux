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

	var responseType int64
	if versionData.newerVersionAvailable {
		responseType = internal.CodeOutdated
	} else {
		responseType = internal.CodeSuccess
	}

	return &pb.PingResponse{
		Type:     responseType,
		Major:    currentVersion.Major,
		Minor:    currentVersion.Minor,
		Patch:    currentVersion.Patch,
		Metadata: currentVersion.Metadata,
	}, nil
}
