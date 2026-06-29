package daemon

import (
	"context"
	"errors"
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

type ensErrorInjector interface {
	InjectVPNConnectionError(code int32, publicKey string) error
}

// InjectVpnConnectionError injects a simulated ENS connection error for DEV.
func (r *RPC) InjectVpnConnectionError(
	ctx context.Context,
	in *pb.InjectVpnConnectionErrorRequest,
) (*pb.Payload, error) {
	if !internal.IsDevEnv(string(r.environment)) {
		return nil, errors.New("ENS injection is available in the dev environment only")
	}

	v, err := r.factory(config.Technology_NORDLYNX)
	if err != nil {
		return nil, fmt.Errorf("getting NordLynx VPN: %w", err)
	}

	injector, ok := v.(ensErrorInjector)
	if !ok {
		return nil, errors.New("active VPN backend does not support ENS injection")
	}

	if err := injector.InjectVPNConnectionError(in.GetTelioCode(), in.GetPubkey()); err != nil {
		return nil, fmt.Errorf("injecting ENS connection error: %w", err)
	}

	return &pb.Payload{Type: internal.CodeSuccess}, nil
}
