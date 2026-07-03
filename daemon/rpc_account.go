package daemon

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
	"google.golang.org/protobuf/proto"
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

// setDedicatedIPServerData sets Dedicated IP and Dedicated Server related fields in the accountInfo
func (r *RPC) setDedicatedIPServerData(accountInfo *pb.AccountResponse) (*pb.AccountResponse, error) {
	dipServices, err := r.ac.GetDedicatedIPServices()
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			return &pb.AccountResponse{Type: internal.CodeExpiredAccessToken},
				fmt.Errorf("getting dedicated ip services: %w", err)
		}
		return &pb.AccountResponse{Type: internal.CodeTokenRenewError},
			fmt.Errorf("getting dedicated ip services: %w", err)
	}

	accountInfo.DedicatedIpStatus = internal.CodeSuccess

	if len(dipServices) < 1 {
		accountInfo.DedicatedIpStatus = internal.CodeNoService
	}

	dedicatedIPExpirationDate := ""
	if len(dipServices) != 0 {
		accountInfo.DedicatedIpServices = dipServicesToProtobuf(dipServices)
		dedicatedIPExpirationDate, err = findLatestDIPExpirationData(dipServices)
		if err != nil {
			return &pb.AccountResponse{Type: internal.CodeTokenRenewError},
				fmt.Errorf("getting latest dedicated ip expiration date: %w", err)
		}
	}

	accountInfo.LastDedicatedIpExpiresAt = dedicatedIPExpirationDate

	dedicatedServerService, err := r.ac.GetDedicatedServerService()
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			return &pb.AccountResponse{Type: internal.CodeExpiredAccessToken},
				fmt.Errorf("getting dedicated server service: %w", err)
		}
		return &pb.AccountResponse{Type: internal.CodeTokenRenewError},
			fmt.Errorf("getting dedicated server service: %w", err)
	}

	accountInfo.DedicatedServerStatus = internal.CodeNoService
	if dedicatedServerService.Active {
		accountInfo.DedicatedServerStatus = internal.CodeSuccess
		accountInfo.DedicatedServersServiceExpiresAt = dedicatedServerService.ExpiresAt
	}

	return accountInfo, nil
}

// AccountInfo returns user account information.
func (r *RPC) AccountInfo(ctx context.Context, req *pb.AccountRequest) (*pb.AccountResponse, error) {
	if ok, err := r.ac.IsLoggedIn(); !ok {
		if errors.Is(err, core.ErrUnauthorized) {
			return &pb.AccountResponse{Type: internal.CodeRevokedAccessToken}, nil
		}
		return nil, internal.ErrNotLoggedIn
	}

	if accountInfo, ok := r.dm.GetAccountData(req.Full); ok {
		// because Dedicated IP and Dedicated Servers service data is using it's own caching that is updated with NC
		// events, we can always try to fetch it to keep the returned information more accurate
		// because account info is provided by reference, it is cloned so that it is not overwritten in case of errors.
		dedicatedServerIPAccountInfo := proto.Clone(accountInfo).(*pb.AccountResponse)
		var err error
		dedicatedServerIPAccountInfo, err = r.setDedicatedIPServerData(dedicatedServerIPAccountInfo)
		if err != nil {
			log.Error("getting dedicated servers/ip service data:", err)
			return accountInfo, nil
		}
		return dedicatedServerIPAccountInfo, nil
	}

	accountInfo := &pb.AccountResponse{}

	vpnExpired, err := r.ac.IsVPNExpired()
	if err != nil {
		log.Error("checking VPN expiration:", err)
		return &pb.AccountResponse{Type: internal.CodeTokenRenewError}, nil
	} else if vpnExpired {
		accountInfo.Type = internal.CodeNoService
	} else {
		accountInfo.Type = internal.CodeSuccess
	}

	accountInfo, err = r.setDedicatedIPServerData(accountInfo)
	if err != nil {
		log.Error("getting dedicated servers/ip service data:", err)
		return accountInfo, nil
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
	accountInfo.SubscriptionExpiresAt = tokenData.ServiceExpiry

	currentUser, err := r.credentialsAPI.CurrentUser()
	if err != nil {
		log.Error("retrieving user:", err)
		switch {
		case errors.Is(err, core.ErrUnauthorized):
			if err := r.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID, r.events.User.Logout, events.ReasonUnauthorized)); err != nil {
				return &pb.AccountResponse{
					Type: internal.CodeConfigError,
				}, nil
			}
			return nil, internal.ErrNotLoggedIn
		}
		return nil, internal.ErrUnhandled
	}

	accountInfo.CreatedOn = currentUser.CreatedOn
	accountInfo.Email = currentUser.Email
	if currentUser.Username != currentUser.Email {
		accountInfo.Username = currentUser.Username
	}

	r.events.Service.AccountCheck.Publish(struct{}{})
	r.dm.SetAccountData(accountInfo)

	return accountInfo, nil
}
