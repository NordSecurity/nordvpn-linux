package daemon

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
	"golang.org/x/exp/slices"
)

func (r *RPC) SetDNS(ctx context.Context, in *pb.SetDNSRequest) (*pb.SetDNSResponse, error) {
	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Error(err)
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
		// Do not allow IPv6 servers
		if !internal.IsAddressValidAsDNSServer(address) {
			return &pb.SetDNSResponse{
				Response: &pb.SetDNSResponse_SetDnsStatus{SetDnsStatus: pb.SetDNSStatus_INVALID_DNS_ADDRESS},
			}, nil
		}
	}

	newRealTimeProtectionStatus := cfg.AutoConnectData.RealTimeProtection

	if newRealTimeProtectionStatus && nameservers != nil {
		newRealTimeProtectionStatus = false
	}

	if nameservers == nil {
		nameservers = r.nameservers.Get(newRealTimeProtectionStatus)
	}

	if err := r.netw.SetDNS(nameservers); err != nil {
		log.Error(err)
		return &pb.SetDNSResponse{
			Response: &pb.SetDNSResponse_ErrorCode{ErrorCode: pb.SetErrorCode_FAILURE},
		}, nil
	}

	if err := r.cm.SaveWith(func(c config.Config) config.Config {
		c.AutoConnectData.RealTimeProtection = newRealTimeProtectionStatus
		c.AutoConnectData.DNS = in.GetDns()
		return c
	}); err != nil {
		log.Error(err)
		return &pb.SetDNSResponse{
			Response: &pb.SetDNSResponse_ErrorCode{ErrorCode: pb.SetErrorCode_CONFIG_ERROR},
		}, nil
	}
	r.events.Settings.DNS.Publish(events.DataDNS{Ips: in.GetDns()})

	if newRealTimeProtectionStatus != cfg.AutoConnectData.RealTimeProtection {
		return &pb.SetDNSResponse{
			Response: &pb.SetDNSResponse_SetDnsStatus{
				SetDnsStatus: pb.SetDNSStatus_DNS_CONFIGURED_RTP_RESET}}, nil
	}

	return &pb.SetDNSResponse{
		Response: &pb.SetDNSResponse_SetDnsStatus{SetDnsStatus: pb.SetDNSStatus_DNS_CONFIGURED}}, nil
}
