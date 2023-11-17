package daemon

import (
	"context"
	"log"
	"sort"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// Countries provides country command and country autocompletion.
func (r *RPC) Countries(ctx context.Context, in *pb.Empty) (*pb.Payload, error) {
	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.Payload{
			Type: internal.CodeConfigError,
		}, nil
	}

	countries, ok := r.dm.GetAppData().CountryNames[cfg.AutoConnectData.Obfuscate][cfg.AutoConnectData.Protocol]
	if !ok {
		return &pb.Payload{
			Type: internal.CodeEmptyPayloadError,
		}, nil
	}
	var countryNames []string
	for country := range countries.Iter() {
		countryNames = append(countryNames, country.(string))
	}
	sort.Strings(countryNames)
	return &pb.Payload{
		Type: internal.CodeSuccess,
		Data: countryNames,
	}, nil
}
