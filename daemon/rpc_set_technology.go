package daemon

import (
	"context"
	"log"
	"strconv"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/features"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func (r *RPC) SetTechnology(ctx context.Context, in *pb.SetTechnologyRequest) (*pb.Payload, error) {
	if in.Technology == config.Technology_NORDWHISPER {
		if !features.NordWhisperEnabled {
			log.Println(internal.DebugPrefix,
				"user requested a NordWhisper technology but the feature is hidden based on compile flag.")
			return &pb.Payload{
				Type: internal.CodeFeatureHidden,
			}, nil
		}
		nordWhisperEnabled, err := r.remoteConfigGetter.GetNordWhisperEnabled(r.version)
		if err != nil {
			log.Println(internal.ErrorPrefix, "failed to determine if NordWhisper is enabled by remote config:", err)
			return &pb.Payload{
				Type: internal.CodeFeatureHidden,
			}, nil
		}

		if !nordWhisperEnabled {
			log.Println(internal.ErrorPrefix,
				"user requested a NordWhisper technology but the feature is hidden based on remote config flag")
			return &pb.Payload{
				Type: internal.CodeFeatureHidden,
			}, nil
		}
	}

	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
	}

	if cfg.Technology == in.GetTechnology() {
		return &pb.Payload{
			Type: internal.CodeNothingToDo,
			Data: []string{config.TechNameToUpperCamelCase(in.GetTechnology())},
		}, nil
	}

	v, err := r.factory(in.GetTechnology())
	if err != nil {
		log.Println(internal.ErrorPrefix, err)
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
	if in.GetTechnology() == config.Technology_NORDLYNX {
		protocol = config.Protocol_UDP
		obfuscate = false
	} else if in.GetTechnology() == config.Technology_NORDWHISPER {
		protocol = config.Protocol_Webtunnel
		obfuscate = false
	}

	if in.GetTechnology() != config.Technology_NORDLYNX && cfg.AutoConnectData.PostquantumVpn {
		return &pb.Payload{
			Type: internal.CodePqWithoutNordlynx,
			Data: []string{config.TechNameToUpperCamelCase(cfg.Technology)},
		}, nil
	}

	if err := r.cm.SaveWith(func(c config.Config) config.Config {
		c.Technology = in.GetTechnology()
		c.AutoConnectData.Protocol = protocol
		c.AutoConnectData.Obfuscate = obfuscate
		return c
	}); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.Payload{
			Type: internal.CodeConfigError,
		}, nil
	}

	// change vpn only when all above checks succeed
	r.netw.SetVPN(v)

	r.events.Settings.Technology.Publish(in.GetTechnology())

	payload.Data = []string{strconv.FormatBool(r.netw.IsVPNActive()),
		config.TechNameToUpperCamelCase(in.GetTechnology())}
	return payload, nil
}
