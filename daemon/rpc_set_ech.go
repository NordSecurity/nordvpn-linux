package daemon

import (
	"context"
	"strconv"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/config/remote"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
)

// SetECH toggles the NordWhisper Encrypted Client Hello (ECH) user setting.
//
// Two guards reject the change instead of persisting: ECH only applies to the NordWhisper
// technology, and remote config can disable ECH globally. Both run before the no-op check so a
// blocked user always gets the explanatory message.
//
// On success this is a pure config write: Quench.Start() re-reads the ECH setting on every
// connect, so the new value is applied automatically on the next connection with no VPN object
// re-creation. The returned Data carries whether a VPN connection is currently active, so the
// CLI can print a "reconnect to apply" hint.
func (r *RPC) SetECH(ctx context.Context, in *pb.SetGenericRequest) (*pb.Payload, error) {
	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Error("failed to load config:", err)
		return &pb.Payload{
			Type: internal.CodeConfigError,
		}, nil
	}

	// Guard 1: ECH only applies to NordWhisper. Checked against the configured technology,
	// like post-quantum's CodePqWithoutNordlynx.
	if cfg.Technology != config.Technology_NORDWHISPER {
		return &pb.Payload{
			Type: internal.CodeECHTechUnsupported,
		}, nil
	}

	// Guard 2: remote config can disable ECH globally. Only an explicit false blocks the toggle;
	// missing/malformed/error defaults to enabled, matching the quench config getter.
	if !r.remoteECHEnabled() {
		return &pb.Payload{
			Type: internal.CodeECHGloballyDisabled,
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

// remoteECHEnabled reports whether the remote-config gate allows ECH. It defaults to enabled on
// any error, missing or malformed value, matching the default-on semantics of the quench config
// getter (daemon/vpn/quench/config.go).
func (r *RPC) remoteECHEnabled() bool {
	param, err := r.remoteConfigGetter.GetFeatureParam(remote.FeatureNordWhisper, "enable_ech")
	if err != nil {
		return true
	}
	enabled, err := strconv.ParseBool(param)
	if err != nil {
		return true
	}
	return enabled
}
