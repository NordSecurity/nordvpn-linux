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
	return r.setTechnologyImpl(in, features.NordWhisperEnabled)
}

func (r *RPC) setTechnologyImpl(in *pb.SetTechnologyRequest, isNordWhisperEnabled bool) (*pb.Payload, error) {
	if in.GetTechnology() == config.Technology_NORDWHISPER {
		if !isNordWhisperEnabled {
			log.Debug("user requested a NordWhisper technology but the feature is hidden based on compile flag.")
			return &pb.Payload{
				Type: internal.CodeFeatureHidden,
			}, nil
		}
	}

	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Error(err)
		return &pb.Payload{Type: internal.CodeConfigError}, nil
	}

	if cfg.Technology == in.GetTechnology() {
		return &pb.Payload{
			Type: internal.CodeNothingToDo,
			Data: []string{config.TechNameToUpperCamelCase(in.GetTechnology())},
		}, nil
	}

	if cfg.AutoConnect &&
		cfg.AutoConnectData.Group == config.ServerGroup_DEDICATED_SERVER &&
		in.GetTechnology() != config.Technology_NORDLYNX {
		return &pb.Payload{
			Type: internal.CodeDedicatedServersNoNordlynx,
		}, nil
	}

	v, err := r.factory(in.GetTechnology())
	if err != nil {
		log.Error(err)
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

	obfuscate := cfg.AutoConnectData.Obfuscate
	if in.GetTechnology() != config.Technology_OPENVPN {
		obfuscate = false
	}

	ech := cfg.AutoConnectData.ECH
	if in.GetTechnology() != config.Technology_NORDWHISPER {
		ech = r.resetECHEnabledField()
	}

	var protocol config.Protocol
	if in.GetTechnology() == config.Technology_NORDWHISPER {
		protocol = config.Protocol_Webtunnel
	} else {
		protocol = config.Protocol_UDP
	}

	if in.GetTechnology() != config.Technology_NORDLYNX && cfg.AutoConnectData.PostquantumVpn {
		return &pb.Payload{
			Type: internal.CodePqWithoutNordlynx,
			Data: []string{config.TechNameToUpperCamelCase(in.GetTechnology())},
		}, nil
	}

	if cfg.AutoConnect {
		serverTag, serverGroup := generateTagFromAutoConnect(cfg)
		// TODO: when switching to single protocol value,
		// 		this is not needed because it can be passed the protocol
		c := cfg
		c.Technology = in.GetTechnology()
		c.AutoConnectData.Protocol = protocol
		c.AutoConnectData.Obfuscate = obfuscate
		c.AutoConnectData.ECH = ech
		insights := r.dm.GetInsightsData().Insights
		if _, err := selectServer(r, &insights, c, serverTag, serverGroup, ""); err != nil {
			log.Error("no server found for auto-connect data and new server technology: ", cfg.AutoConnectData, in.GetTechnology().String(), err)
			return &pb.Payload{
				Type: internal.CodeTechnologyIncompatibleWithAutoconnect,
			}, nil
		}
	}

	if err := r.cm.SaveWith(func(c config.Config) config.Config {
		c.Technology = in.GetTechnology()
		c.AutoConnectData.Protocol = protocol
		c.AutoConnectData.Obfuscate = obfuscate
		c.AutoConnectData.ECH = ech
		return c
	}); err != nil {
		log.Error(err)
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

// TODO: check if this is still needed after LVPN-9355
func generateTagFromAutoConnect(cfg config.Config) (string, string) {
	var tag, group string
	if cfg.AutoConnectData.Group != config.ServerGroup_UNDEFINED {
		group = cfg.AutoConnectData.Group.String()
	}

	if cfg.AutoConnectData.Country != "" {
		tag = cfg.AutoConnectData.Country
		if cfg.AutoConnectData.City != "" {
			tag += " " + cfg.AutoConnectData.City
		}
	}

	return tag, group
}
