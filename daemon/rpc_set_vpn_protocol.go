package daemon

import (
	"context"
	"errors"
	"slices"
	"strconv"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
)

func (r *RPC) SetVPNProtocol(
	ctx context.Context,
	req *pb.SetVPNProtocolRequest,
) (*pb.Payload, error) {
	if !slices.Contains(config.AvailableVPNProtocols(), req.GetVpnProtocol()) {
		log.Debug("requested VPN protocol is not available in this build:", req.GetVpnProtocol())
		return &pb.Payload{Type: internal.CodeFeatureHidden}, nil
	}

	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Error("configuration loading at VPN protocol change", err)
		return &pb.Payload{Type: internal.CodeConfigError}, nil
	}

	if req.VpnProtocol == cfg.VPNProtocol {
		return &pb.Payload{
			Type: internal.CodeNothingToDo,
		}, nil
	}

	v, err := r.factory(req.GetVpnProtocol())
	if err != nil {
		log.Errorf("VPN factory fails for %s: %v", req.GetVpnProtocol().DisplayName(), err)
		return &pb.Payload{Type: internal.CodeFeatureHidden}, nil
	}

	pq := cfg.AutoConnectData.PostquantumVpn
	if pq && !req.GetVpnProtocol().IsNordLynx() {
		return &pb.Payload{
			Type: internal.CodePqWithoutNordlynx,
		}, nil
	}

	if cfg.AutoConnect {
		serverTag, serverGroup := generateTagFromAutoConnect(cfg)

		c := cfg
		c.VPNProtocol = req.GetVpnProtocol()

		_, err := r.validateServerInput(serverTag, serverGroup, c)
		if err != nil {
			log.Error("no server found for auto-connect data and new server protocol", req.GetVpnProtocol().DisplayName(), err)
			var errorCode *internal.ErrorWithCode
			if errors.As(err, &errorCode) {
				return &pb.Payload{
					Type: errorCode.Code,
				}, nil
			}

			return &pb.Payload{
				Type: internal.CodeProtocolIncompatibleWithAutoconnect,
			}, nil
		}
	}

	ech := cfg.AutoConnectData.ECH
	if ech.Get() && !req.GetVpnProtocol().IsNordWhisper() {
		ech = r.resetECHEnabledField()
	}

	if err := r.cm.SaveWith(func(c config.Config) config.Config {
		c.VPNProtocol = req.GetVpnProtocol()
		c.AutoConnectData.ECH = ech
		return c
	}); err != nil {
		log.Error("cannot save configuration at protocol change", err)
		return &pb.Payload{Type: internal.CodeConfigError}, nil
	}

	// change vpn only when all above checks succeed
	r.netw.SetVPN(v)

	r.events.Settings.VPNProtocol.Publish(req.GetVpnProtocol())

	return &pb.Payload{
		Type: internal.CodeSuccess,
		Data: []string{strconv.FormatBool(r.netw.IsVPNActive())},
	}, nil
}

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
