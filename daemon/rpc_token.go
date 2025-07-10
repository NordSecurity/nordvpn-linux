package daemon

import (
	"context"
	"errors"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// TokenInfo returns token information.
func (r *RPC) TokenInfo(ctx context.Context, _ *pb.Empty) (*pb.TokenInfoResponse, error) {
	if ok, err := r.ac.IsLoggedIn(); !ok {
		if errors.Is(err, core.ErrLoginTokenRevoked) {
			return &pb.TokenInfoResponse{Type: internal.CodeRevokedAccessToken}, nil
		}
		return nil, internal.ErrNotLoggedIn
	}

	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		return &pb.TokenInfoResponse{Type: internal.CodeConfigError}, nil
	}

	tokenData := cfg.TokensData[cfg.AutoConnectData.ID]
	tokenInfo := &pb.TokenInfoResponse{
		Type:               internal.CodeSuccess,
		Token:              tokenData.Token,
		ExpiresAt:          tokenData.TokenExpiry,
		TrustedPassToken:   tokenData.TrustedPassToken,
		TrustedPassOwnerId: tokenData.TrustedPassOwnerID,
	}

	return tokenInfo, nil
}
