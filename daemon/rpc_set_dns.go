package daemon

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func (r *RPC) SetDNS(ctx context.Context, in *pb.SetDNSRequest) (*pb.Payload, error) {
	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
	}

	var nameservers []string
	if in.GetDns() != nil {
		nameservers = in.GetDns()
	} else {
		subnet, _ := r.endpoint.Network() // safe to ignore the error
		nameservers = r.nameservers.Get(in.GetThreatProtectionLite(), subnet.Addr().Is6())
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
	enabled := len(in.GetDns()) > 0
	r.events.Settings.DNS.Publish(events.DataDNS{Enabled: enabled, Ips: in.GetDns()})

	return &pb.Payload{
		Type: internal.CodeSuccess,
	}, nil
}
