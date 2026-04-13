package daemon

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/log"
)

// RecommendedServer fetches the recommended server location
func (r *RPC) RecommendedServer(ctx context.Context, in *pb.Empty) (*pb.RecommendedServerLocation, error) {
	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Error(err)
		return &pb.RecommendedServerLocation{}, nil
	}
	insights := r.dm.GetInsightsData().Insights

	serverSelection, err := selectServer(r, &insights, cfg, "", "")
	if err == nil {
		country := serverSelection.server.Country()
		return &pb.RecommendedServerLocation{
			CityName:    country.City.Name,
			CountryCode: country.Code,
			CountryName: country.Name,
		}, nil
	} else {
		log.Error("Failed to fetch the recommended server", err)
	}

	return &pb.RecommendedServerLocation{}, nil
}
