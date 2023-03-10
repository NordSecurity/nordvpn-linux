package daemon

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func (r *RPC) SetThreatProtectionLite(
	ctx context.Context,
	in *pb.SetThreatProtectionLiteRequest,
) (*pb.Payload, error) {
	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
	}

	var nameservers []string
	if in.GetDns() != nil {
		nameservers = in.GetDns()
	} else {
		nameservers = r.nameservers.Get(in.GetThreatProtectionLite(), cfg.IPv6)
	}

	if err := r.netw.SetDNS(nameservers); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.Payload{
			Type: internal.CodeFailure,
		}, nil
	}

	if err := r.cm.SaveWith(func(c config.Config) config.Config {
		c.AutoConnectData.ThreatProtectionLite = in.GetThreatProtectionLite()
		c.AutoConnectData.DNS = in.GetDns()
		return c
	}); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.Payload{
			Type: internal.CodeConfigError,
		}, nil
	}
	r.events.Settings.ThreatProtectionLite.Publish(in.GetThreatProtectionLite())

	return &pb.Payload{
		Type: internal.CodeSuccess,
	}, nil
}
