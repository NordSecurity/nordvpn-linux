package daemon

import (
	"context"
	"log"
	"sort"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// Cities provides cities command and autocompletion.
func (r *RPC) Cities(ctx context.Context, in *pb.CitiesRequest) (*pb.Payload, error) {
	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
	}

	// collect cities and sort them
	if value, ok := r.dm.GetAppData().CityNames[in.GetObfuscate()][cfg.AutoConnectData.Protocol][strings.ToLower(in.GetCountry())]; ok {
		var namesList []string
		for city := range value.Iter() {
			namesList = append(namesList, city.(string))
		}
		sort.Strings(namesList)
		return &pb.Payload{
			Type: internal.CodeSuccess,
			Data: namesList,
		}, nil
	}
	return &pb.Payload{
		Type: internal.CodeEmptyPayloadError,
	}, nil
}
