package daemon

import (
	"context"
	"sort"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// Groups provides endpoint and autocompletion.
func (r *RPC) Groups(ctx context.Context, in *pb.GroupsRequest) (*pb.Payload, error) {
	var groupNames []string
	for group := range r.dm.GetAppData().GroupNames[in.GetObfuscate()][in.GetProtocol()].Iter() {
		groupNames = append(groupNames, group.(string))
	}

	sort.Strings(groupNames)
	return &pb.Payload{
		Type: internal.CodeSuccess,
		Data: groupNames,
	}, nil
}
