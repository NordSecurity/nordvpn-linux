package daemon

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
)

// Settings returns system daemon settings
func (r *RPC) Settings(ctx context.Context, in *pb.Empty) (*pb.SettingsResponse, error) {
	cred, err := internal.UcredFromContext(ctx)
	if err != nil {
		log.Error("Settings:", err)
		return &pb.SettingsResponse{Type: internal.CodeFailure}, nil
	}
	uid := int64(cred.Uid)

	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Error(err)
		return &pb.SettingsResponse{
			Type: internal.CodeConfigError,
		}, nil
	}

	// ECH value is stored in config but controlled by remote config as well
	cfg.AutoConnectData.ECH = r.getECHEnabledField(cfg)

	settings := configToProtobuf(&cfg, uid)

	return &pb.SettingsResponse{
		Type: internal.CodeSuccess,
		Data: settings,
	}, nil
}

func (r *RPC) SettingsProtocols(ctx context.Context, _ *pb.Empty) (*pb.Payload, error) {
	return &pb.Payload{
		Type: internal.CodeSuccess,
		Data: []string{config.Protocol_UDP.String(), config.Protocol_TCP.String()},
	}, nil
}

func (r *RPC) SettingsTechnologies(ctx context.Context, _ *pb.Empty) (*pb.Payload, error) {
	return &pb.Payload{
		Type: internal.CodeSuccess,
		Data: []string{
			config.Technology_OPENVPN.String(),
			config.Technology_NORDLYNX.String(),
			config.Technology_NORDWHISPER.String(),
		},
	}, nil
}
