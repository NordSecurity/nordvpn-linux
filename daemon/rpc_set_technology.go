package daemon

import (
	"context"
	"strconv"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/features"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
)

func (r *RPC) SetTechnology(ctx context.Context, in *pb.SetTechnologyRequest) (*pb.Payload, error) {
	log.RPCSetTechnology.Tracef("requested technology=%v vpnActive=%v", in.GetTechnology(), r.netw.IsVPNActive())
	if in.Technology == config.Technology_NORDWHISPER {
		if !features.NordWhisperEnabled {
			log.RPCSetTechnology.Debug("user requested a NordWhisper technology but the feature is hidden based on compile flag.")
			return &pb.Payload{
				Type: internal.CodeFeatureHidden,
			}, nil
		}
	}

	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.RPCSetTechnology.Error("loading config:", err)
	}

	if cfg.Technology == in.GetTechnology() {
		log.RPCSetTechnology.Tracef("technology already set to %v, nothing to do", in.GetTechnology())
		return &pb.Payload{
			Type: internal.CodeNothingToDo,
			Data: []string{config.TechNameToUpperCamelCase(in.GetTechnology())},
		}, nil
	}

	if cfg.AutoConnect &&
		cfg.AutoConnectData.Group == config.ServerGroup_DEDICATED_SERVER &&
		in.GetTechnology() != config.Technology_NORDLYNX {
		log.RPCSetTechnology.Tracef("auto-connect is set to dedicated server group, technology must be NordLynx, requested=%v", in.GetTechnology())
		return &pb.Payload{
			Type: internal.CodeDedicatedServersNoNordlynx,
		}, nil
	}

	log.RPCSetTechnology.Tracef("creating VPN factory: technology=%v currentTechnology=%v",
		in.GetTechnology(), cfg.Technology)
	v, err := r.factory(in.GetTechnology())
	if err != nil {
		log.RPCSetTechnology.Error("creating VPN factory:", err)
		return &pb.Payload{
			Type: internal.CodeConfigError,
		}, nil
	}

	payload := &pb.Payload{}
	// payload.Type gets overridden in case of failure
	//
	// Previously it was at the end of the function, which overrode any failures
	// and user was not given any error messages because of it. Most notably,
	// internal.CodeSuccessWithoutAC was overridden with generic internal.CodeSuccess
	payload.Type = internal.CodeSuccess

	protocol := cfg.AutoConnectData.Protocol
	obfuscate := cfg.AutoConnectData.Obfuscate
	if in.GetTechnology() != config.Technology_OPENVPN {
		obfuscate = false
	}

	ech := cfg.AutoConnectData.ECH
	if in.GetTechnology() != config.Technology_NORDWHISPER {
		ech = r.resetECHEnabledField()
	}

	if in.GetTechnology() == config.Technology_NORDWHISPER {
		protocol = config.Protocol_Webtunnel
	} else {
		protocol = config.Protocol_UDP
	}

	log.RPCSetTechnology.Tracef("derived settings: protocol=%v obfuscate=%v ech=%v", protocol, obfuscate, ech)

	if in.GetTechnology() != config.Technology_NORDLYNX && cfg.AutoConnectData.PostquantumVpn {
		log.RPCSetTechnology.Tracef("post-quantum requires NordLynx, rejecting technology change to %v", in.GetTechnology())
		return &pb.Payload{
			Type: internal.CodePqWithoutNordlynx,
			Data: []string{config.TechNameToUpperCamelCase(in.GetTechnology())},
		}, nil
	}

	if err := r.cm.SaveWith(func(c config.Config) config.Config {
		c.Technology = in.GetTechnology()
		c.AutoConnectData.Protocol = protocol
		c.AutoConnectData.Obfuscate = obfuscate
		c.AutoConnectData.ECH = ech
		return c
	}); err != nil {
		log.RPCSetTechnology.Error("saving technology config:", err)
		return &pb.Payload{
			Type: internal.CodeConfigError,
		}, nil
	}

	// change vpn only when all above checks succeed
	log.RPCSetTechnology.Tracef("applying technology change to %v", in.GetTechnology())
	r.netw.SetVPN(v)

	r.events.Settings.Technology.Publish(in.GetTechnology())

	payload.Data = []string{strconv.FormatBool(r.netw.IsVPNActive()),
		config.TechNameToUpperCamelCase(in.GetTechnology())}
	return payload, nil
}
