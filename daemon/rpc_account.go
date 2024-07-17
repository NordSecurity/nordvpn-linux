package daemon

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func findLatestDIPExpirationData(dipServices []auth.DedicatedIPService) (string, error) {
	if len(dipServices) == 0 {
		return "", fmt.Errorf("no dip services found")
	}

	layout := "2006-01-02 15:04:05"
	latest, err := time.Parse(layout, dipServices[0].ExpiresAt)
	if err != nil {
		return "", fmt.Errorf("failed to parse initial expiration date")
	}

	for _, dipService := range dipServices {
		current, err := time.Parse(layout, dipService.ExpiresAt)
		if err != nil {
			return "", fmt.Errorf("failed to parse expiration date")
		}

		if current.After(latest) {
			latest = current
		}
	}

	return latest.Format(layout), nil
}

// AccountInfo returns user account information.
func (r *RPC) AccountInfo(ctx context.Context, _ *pb.Empty) (*pb.AccountResponse, error) {
	if !r.ac.IsLoggedIn() {
		return nil, internal.ErrNotLoggedIn
	}

	accountInfo := &pb.AccountResponse{}

	vpnExpired, err := r.ac.IsVPNExpired()
	if err != nil {
		log.Println(internal.ErrorPrefix, "checking VPN expiration: ", err)
		return &pb.AccountResponse{Type: internal.CodeTokenRenewError}, nil
	} else if vpnExpired {
		accountInfo.Type = internal.CodeNoService
	} else {
		accountInfo.Type = internal.CodeSuccess
	}

	accountInfo.DedicatedIpStatus = internal.CodeSuccess
	dipServices, err := r.ac.GetDedicatedIPServices()
	if err != nil {
		log.Println(internal.ErrorPrefix, "getting dedicated ip services: %w", err)
		return &pb.AccountResponse{Type: internal.CodeTokenRenewError}, nil
	}

	if len(dipServices) == 0 {
		accountInfo.DedicatedIpStatus = internal.CodeNoService
	}

	dedicatedIPExpirationDate, err := findLatestDIPExpirationData(dipServices)
	if err != nil {
		log.Println(internal.ErrorPrefix, "getting latest dedicated ip expiration date: %w", err)
		return &pb.AccountResponse{Type: internal.CodeTokenRenewError}, nil
	}

	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		return &pb.AccountResponse{
			Type: internal.CodeConfigError,
		}, nil
	}

	tokenData := cfg.TokensData[cfg.AutoConnectData.ID]
	accountInfo.ExpiresAt = tokenData.ServiceExpiry
	accountInfo.DedicatedIpExpiresAt = dedicatedIPExpirationDate

	currentUser, err := r.credentialsAPI.CurrentUser(tokenData.Token)
	if err != nil {
		log.Println(internal.ErrorPrefix, "retrieving user:", err)
		switch {
		case errors.Is(err, core.ErrUnauthorized):
			if err := r.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID)); err != nil {
				return &pb.AccountResponse{
					Type: internal.CodeConfigError,
				}, nil
			}
			return nil, internal.ErrNotLoggedIn
		}
		return nil, internal.ErrUnhandled
	}

	accountInfo.Email = currentUser.Email
	if currentUser.Username != currentUser.Email {
		accountInfo.Username = currentUser.Username
	}

	r.events.Service.AccountCheck.Publish(
		core.ServicesResponse{},
	)

	return accountInfo, nil
}
