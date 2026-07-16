package daemon

import (
	"context"
	"strconv"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
)

// SetECH toggles the NordWhisper Encrypted Client Hello (ECH) user setting.
//
// This is a pure config write: Quench.Start() re-reads the ECH setting on every connect, so the
// new value is applied automatically on the next connection with no VPN object re-creation. The
// returned Data carries whether a VPN connection is currently active, so the CLI can print a
// "reconnect to apply" hint.
func (r *RPC) SetECH(ctx context.Context, in *pb.SetGenericRequest) (*pb.Payload, error) {
	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Error("failed to load config:", err)
		return &pb.Payload{
			Type: internal.CodeConfigError,
		}, nil
	}

	if cfg.ECH.Get() == in.GetEnabled() {
		return &pb.Payload{
			Type: internal.CodeNothingToDo,
		}, nil
	}

	if err := r.cm.SaveWith(func(c config.Config) config.Config {
		c.ECH.Set(in.GetEnabled())
		return c
	}); err != nil {
		log.Error("failed to save config:", err)
		return &pb.Payload{
			Type: internal.CodeConfigError,
		}, nil
	}

	return &pb.Payload{
		Type: internal.CodeSuccess,
		Data: []string{strconv.FormatBool(r.netw.IsVPNActive())},
	}, nil
}
