package daemon

import (
	"context"
	"log"
	"sort"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// Groups provides endpoint and autocompletion.
func (r *RPC) Groups(ctx context.Context, in *pb.GroupsRequest) (*pb.Payload, error) {
	var cfg config.Config
	if r.cm == nil || r.dm == nil {
		log.Println(internal.ErrorPrefix, "configuration Manager or DataManager is nil")
		return &pb.Payload{
			Type: internal.CodeEmptyPayloadError,
		}, nil
	}

	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.Payload{
			Type: internal.CodeConfigError,
		}, nil
	}

	groups := r.dm.GetAppData().GroupNames[in.GetObfuscate()][cfg.AutoConnectData.Protocol]
	if groups == nil {
		log.Println(internal.ErrorPrefix, "groups list are nil")
		return &pb.Payload{
			Type: internal.CodeEmptyPayloadError,
		}, nil
	}

	var groupNames []string
	for group := range groups.Iter() {
		groupNames = append(groupNames, group.(string))
	}

	sort.Strings(groupNames)
	return &pb.Payload{
		Type: internal.CodeSuccess,
		Data: groupNames,
	}, nil
}
