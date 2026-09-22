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
	log.RPCSetDNS.Tracef("nameserverCount=%d", len(in.GetDns()))
	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.RPCSetDNS.Error("loading config:", err)
	}

	nameservers := in.GetDns()

	if len(nameservers) > 3 {
		log.RPCSetDNS.Tracef("too many nameservers provided: count=%d, max=3", len(nameservers))
		return &pb.SetDNSResponse{
			Response: &pb.SetDNSResponse_SetDnsStatus{SetDnsStatus: pb.SetDNSStatus_TOO_MANY_VALUES},
		}, nil
	}

	nameserverCheck := slices.Clone(nameservers)
	autoConnectDataCheck := slices.Clone(cfg.AutoConnectData.DNS)
	slices.Sort(nameserverCheck)
	slices.Sort(autoConnectDataCheck)
	if slices.Equal(nameserverCheck, autoConnectDataCheck) {
		log.RPCSetDNS.Trace("DNS already set to requested values, nothing to do")
		return &pb.SetDNSResponse{
			Response: &pb.SetDNSResponse_ErrorCode{ErrorCode: pb.SetErrorCode_ALREADY_SET},
		}, nil
	}

	for _, address := range nameservers {
		// Do not allow IPv6 servers
		if !internal.IsAddressValidAsDNSServer(address) {
			log.RPCSetDNS.Trace("invalid DNS server address, rejecting")
			return &pb.SetDNSResponse{
				Response: &pb.SetDNSResponse_SetDnsStatus{SetDnsStatus: pb.SetDNSStatus_INVALID_DNS_ADDRESS},
			}, nil
		}
	}

	newRealTimeProtectionStatus := cfg.AutoConnectData.RealTimeProtection

	if newRealTimeProtectionStatus && nameservers != nil {
		log.RPCSetDNS.Trace("custom DNS provided, disabling real-time protection")
		newRealTimeProtectionStatus = false
	}

	if nameservers == nil {
		log.RPCSetDNS.Tracef("no custom DNS provided, using default nameservers for rtp=%v", newRealTimeProtectionStatus)
		nameservers = r.nameservers.Get(newRealTimeProtectionStatus)
	}

	log.RPCSetDNS.Tracef("nameserverCount=%d rtpReset=%v",
		len(nameservers), newRealTimeProtectionStatus != cfg.AutoConnectData.RealTimeProtection)
	if err := r.netw.SetDNS(nameservers); err != nil {
		log.RPCSetDNS.Error("applying DNS to networker:", err)
		return &pb.SetDNSResponse{
			Response: &pb.SetDNSResponse_ErrorCode{ErrorCode: pb.SetErrorCode_FAILURE},
		}, nil
	}

	if err := r.cm.SaveWith(func(c config.Config) config.Config {
		c.AutoConnectData.RealTimeProtection = newRealTimeProtectionStatus
		c.AutoConnectData.DNS = in.GetDns()
		return c
	}); err != nil {
		log.RPCSetDNS.Error("saving DNS config:", err)
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
