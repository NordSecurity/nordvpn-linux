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
		return &pb.Payload{Type: internal.CodeObfuscatedNeedsNordwhisper}, nil
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
