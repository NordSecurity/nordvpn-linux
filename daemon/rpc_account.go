package daemon

import (
	"context"
	"errors"
	"log"

	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// AccountInfo returns user account information.
func (r *RPC) AccountInfo(ctx context.Context, _ *pb.Empty) (*pb.AccountResponse, error) {
	if !r.ac.IsLoggedIn() {
		return nil, internal.ErrNotLoggedIn
	}

	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		return &pb.AccountResponse{
			Type: internal.CodeConfigError,
		}, nil
	}

	tokenData := cfg.TokensData[cfg.AutoConnectData.ID]

	accountInfo := &pb.AccountResponse{
		ExpiresAt: tokenData.ServiceExpiry,
	}

	if auth.IsTokenExpired(tokenData.ServiceExpiry) {
		accountInfo.Type = internal.CodeNoVPNService
	} else {
		accountInfo.Type = internal.CodeSuccess
	}

	currentUser, err := r.credentialsAPI.CurrentUser(tokenData.Token)
	if err != nil {
		log.Println(internal.ErrorPrefix, "retrieving user:", err)
		switch {
		case errors.Is(err, core.ErrUnauthorized):
			if err := r.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID)); err != nil {
				return nil, err
			}
			return nil, internal.ErrNotLoggedIn
		}
		return nil, err
	}

	accountInfo.Email = currentUser.Email
	if currentUser.Username != currentUser.Email {
		accountInfo.Username = currentUser.Username
	}

	r.events.Service.AccountCheck.Publish(
		core.ServicesResponse{},
	)

	return accountInfo, err
}
