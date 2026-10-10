package daemon

import (
	"context"
	"slices"

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
		log.Error(err)
		// preventing following checks/actions on an empty config
		return &pb.Payload{Type: internal.CodeConfigError}, nil
	}

	if req.VpnProtocol == cfg.VPNProtocol {
		return &pb.Payload{
			Type: internal.CodeNothingToDo,
			Data: []string{req.GetVpnProtocol().DisplayName()},
		}, nil
	}

	if cfg.AutoConnect &&
		cfg.AutoConnectData.Group == config.ServerGroup_DEDICATED_SERVER &&
		!req.GetVpnProtocol().IsNordLynx() {
		return &pb.Payload{Type: internal.CodeDedicatedServersNoNordlynx}, nil
	}

	v, err := r.factory(req.GetVpnProtocol().Technology())
	if err != nil {
		log.Error(err)
		return &pb.Payload{Type: internal.CodeConfigError}, nil
	}

	if cfg.AutoConnect && cfg.AutoConnectData.Group == config.ServerGroup_NW_OBFUSCATED &&
		!req.GetVpnProtocol().IsNordWhisper() {
		return &pb.Payload{
			Type: internal.CodeObfuscatedNeedsNordwhisper,
			Data: []string{req.GetVpnProtocol().DisplayName()},
		}, nil
	}

	ech := cfg.AutoConnectData.ECH
	if ech.Get() && !req.GetVpnProtocol().IsNordWhisper() {
		ech = r.resetECHEnabledField()
	}

	pq := cfg.AutoConnectData.PostquantumVpn
	if pq && !req.GetVpnProtocol().IsNordLynx() {
		return &pb.Payload{
			Type: internal.CodePqWithoutNordlynx,
			Data: []string{req.GetVpnProtocol().DisplayName()},
		}, nil
	}

	if cfg.AutoConnect {
		serverTag, serverGroup := generateTagFromAutoConnect(cfg)
		c := cfg
		c.VPNProtocol = req.GetVpnProtocol()
		c.AutoConnectData.ECH = ech
		insights := r.dm.GetInsightsData().Insights
		if _, err := selectServer(r, &insights, c, serverTag, serverGroup, ""); err != nil {
			log.Error("no server found for auto-connect data and new server technology: ", cfg.AutoConnectData, req.GetVpnProtocol().DisplayName(), err)
			return &pb.Payload{
				Type: internal.CodeProtocolIncompatibleWithAutoconnect,
			}, nil
		}
	}

	updateConfigFn := func(c config.Config) config.Config {
		c.VPNProtocol = req.GetVpnProtocol()
		c.AutoConnectData.ECH = ech
		return c
	}

	if err := r.cm.SaveWith(updateConfigFn); err != nil {
		log.Error(err)
		return &pb.Payload{Type: internal.CodeConfigError}, nil
	}

	// change vpn only when all above checks succeed
	r.netw.SetVPN(v)

	r.events.Settings.VPNProtocol.Publish(req.GetVpnProtocol())

	payload := &pb.Payload{
		Type: internal.CodeSuccess,
		Data: []string{req.GetVpnProtocol().DisplayName()},
	}
	if r.netw.IsVPNActive() {
		payload.Type = internal.CodeSuccessReconnectRequired
	}

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
