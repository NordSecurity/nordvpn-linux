package daemon

import (
	"context"
	"log"
	"net"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"golang.org/x/exp/slices"
)

func (r *RPC) SetDNS(ctx context.Context, in *pb.SetDNSRequest) (*pb.SetDNSResponse, error) {
	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
	}

	nameservers := in.GetDns()

	if len(nameservers) > 3 {
		return &pb.SetDNSResponse{
			Response: &pb.SetDNSResponse_SetDnsStatus{SetDnsStatus: pb.SetDNSStatus_TOO_MANY_VALUES},
		}, nil
	}

	nameserverCheck := slices.Clone(nameservers)
	autoConnectDataCheck := slices.Clone(cfg.AutoConnectData.DNS)
	slices.Sort(nameserverCheck)
	slices.Sort(autoConnectDataCheck)
	if slices.Equal(nameserverCheck, autoConnectDataCheck) {
		return &pb.SetDNSResponse{
			Response: &pb.SetDNSResponse_ErrorCode{ErrorCode: pb.SetErrorCode_ALREADY_SET},
		}, nil
	}

	for _, address := range nameservers {
		if parsedAddress := net.ParseIP(address); parsedAddress == nil {
			return &pb.SetDNSResponse{
				Response: &pb.SetDNSResponse_SetDnsStatus{SetDnsStatus: pb.SetDNSStatus_INVALID_DNS_ADDRESS},
			}, nil
		}
	}

	newThreatProtectionLiteStatus := cfg.AutoConnectData.ThreatProtectionLite

	if newThreatProtectionLiteStatus && nameservers != nil {
		newThreatProtectionLiteStatus = false
	}

	if nameservers == nil {
		subnet, _ := r.endpoint.Network() // safe to ignore the error
		nameservers = r.nameservers.Get(newThreatProtectionLiteStatus, subnet.Addr().Is6())
	}

	if err := r.netw.SetDNS(nameservers); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.SetDNSResponse{
			Response: &pb.SetDNSResponse_ErrorCode{ErrorCode: pb.SetErrorCode_FAILURE},
		}, nil
	}

	if err := r.cm.SaveWith(func(c config.Config) config.Config {
		c.AutoConnectData.ThreatProtectionLite = newThreatProtectionLiteStatus
		c.AutoConnectData.DNS = in.GetDns()
		return c
	}); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.SetDNSResponse{
			Response: &pb.SetDNSResponse_ErrorCode{ErrorCode: pb.SetErrorCode_CONFIG_ERROR},
		}, nil
	}
	r.events.Settings.DNS.Publish(events.DataDNS{Ips: in.GetDns()})

	if newThreatProtectionLiteStatus != cfg.AutoConnectData.ThreatProtectionLite {
		return &pb.SetDNSResponse{
			Response: &pb.SetDNSResponse_SetDnsStatus{
				SetDnsStatus: pb.SetDNSStatus_DNS_CONFIGURED_TPL_RESET}}, nil
	}

	return &pb.SetDNSResponse{
		Response: &pb.SetDNSResponse_SetDnsStatus{SetDnsStatus: pb.SetDNSStatus_DNS_CONFIGURED}}, nil
}
