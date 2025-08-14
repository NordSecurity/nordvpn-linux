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

	layout := internal.ServerDateFormat
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

// dipServicesToProtobuf converts internal DedicatedIPService structure to the protobuf generated structure.
func dipServicesToProtobuf(dipServices []auth.DedicatedIPService) []*pb.DedidcatedIPService {
	dipServicesProtobuf := []*pb.DedidcatedIPService{}
	for _, dipService := range dipServices {
		dipServicesProtobuf = append(dipServicesProtobuf, &pb.DedidcatedIPService{
			ServerIds:            dipService.ServerIDs,
			DedicatedIpExpiresAt: dipService.ExpiresAt,
		})
	}

	return dipServicesProtobuf
}

// AccountInfo returns user account information.
func (r *RPC) AccountInfo(ctx context.Context, req *pb.AccountRequest) (*pb.AccountResponse, error) {
	r.credentialsAPI.CurrentUser() // remove me
	if ok, err := r.ac.IsLoggedIn(); !ok {
		if errors.Is(err, core.ErrUnauthorized) {
			return &pb.AccountResponse{Type: internal.CodeRevokedAccessToken}, nil
		}
		return nil, internal.ErrNotLoggedIn
	}

	if accountInfo, ok := r.dm.GetAccountData(req.Full); ok {
		return accountInfo, nil
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
		log.Println(internal.ErrorPrefix, "getting dedicated ip services:", err)
		if errors.Is(err, core.ErrUnauthorized) {
			return &pb.AccountResponse{Type: internal.CodeExpiredAccessToken}, nil
		} else {
			return &pb.AccountResponse{Type: internal.CodeTokenRenewError}, nil
		}
	}

	if len(dipServices) < 1 {
		accountInfo.DedicatedIpStatus = internal.CodeNoService
	}

	dedicatedIPExpirationDate := ""
	if len(dipServices) != 0 {
		accountInfo.DedicatedIpServices = dipServicesToProtobuf(dipServices)
		dedicatedIPExpirationDate, err = findLatestDIPExpirationData(dipServices)
		if err != nil {
			log.Println(internal.ErrorPrefix, "getting latest dedicated ip expiration date:", err)
			return &pb.AccountResponse{Type: internal.CodeTokenRenewError}, nil
		}
	}

	// get user's current mfa status
	accountInfo.MfaStatus = pb.TriState_DISABLED
	mfaStatus, err := r.ac.IsMFAEnabled()
	if err != nil {
		accountInfo.MfaStatus = pb.TriState_UNKNOWN
	} else if mfaStatus {
		accountInfo.MfaStatus = pb.TriState_ENABLED
	}

	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		return &pb.AccountResponse{
			Type: internal.CodeConfigError,
		}, nil
	}

	tokenData := cfg.TokensData[cfg.AutoConnectData.ID]
	accountInfo.ExpiresAt = tokenData.ServiceExpiry
	accountInfo.LastDedicatedIpExpiresAt = dedicatedIPExpirationDate

	currentUser, err := r.credentialsAPI.CurrentUser()
	if err != nil {
		log.Println(internal.ErrorPrefix, "retrieving user:", err)
		switch {
		case errors.Is(err, core.ErrUnauthorized):
			if err := r.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID, r.events.User.Logout)); err != nil {
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

	r.events.Service.AccountCheck.Publish(struct{}{})
	r.dm.SetAccountData(accountInfo)

	return accountInfo, nil
}
