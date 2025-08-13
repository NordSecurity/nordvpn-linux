/*
Package auth is responsible for user authentication.
*/
package auth

import (
	"fmt"
	"sync"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	daemonevents "github.com/NordSecurity/nordvpn-linux/daemon/events"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/session"
)

type DedicatedIPService struct {
	ExpiresAt string
	// ServerID will be set to NoServerSelected if server was not selected by the user
	ServerIDs []int64
}

// Checker provides information about current authentication.
type Checker interface {
	// IsLoggedIn returns true when the user is logged in.
	IsLoggedIn() (bool, error)
	// IsMFAEnabled returns true if Multifactor Authentication is enabled.
	IsMFAEnabled() (bool, error)
	// IsVPNExpired is used to check whether the user is allowed to use VPN
	IsVPNExpired() (bool, error)
	// GetDedicatedIPServices returns all available server IDs, if server is not selected by the user it will set
	// ServerID for that service to NoServerSelected
	GetDedicatedIPServices() ([]DedicatedIPService, error)
}

const (
	VPNServiceID         = 1
	DedicatedIPServiceID = 11
)

type systemTimeExpirationChecker struct{}

// isTokenExpired reports whether the token is expired or not.
func (systemTimeExpirationChecker) IsExpired(expiryTime string) bool {
	if expiryTime == "" {
		return true
	}

	expiry, err := time.Parse(internal.ServerDateFormat, expiryTime)
	if err != nil {
		return true
	}

	return time.Now().After(expiry)
}

// NewTokenExpirationChecker
func NewTokenExpirationChecker() core.ExpirationChecker {
	return &systemTimeExpirationChecker{}
}

// RenewingChecker does both authentication checks and renewals in case of expiration.
type RenewingChecker struct {
	cm                  config.Manager
	creds               core.CredentialsAPI
	expChecker          core.ExpirationChecker
	mfaPub              events.Publisher[bool]
	logoutPub           events.Publisher[events.DataAuthorization]
	errPub              events.Publisher[error]
	mu                  sync.Mutex
	accountUpdateEvents *daemonevents.AccountUpdateEvents
	sessionStores       []session.SessionStore
}

// NewRenewingChecker is a default constructor for RenewingChecker.
func NewRenewingChecker(cm config.Manager,
	creds core.CredentialsAPI,
	mfaPub events.Publisher[bool],
	logoutPub events.Publisher[events.DataAuthorization],
	errPub events.Publisher[error],
	accountUpdateEvents *daemonevents.AccountUpdateEvents,
	sessionStores ...session.SessionStore,
) *RenewingChecker {
	return &RenewingChecker{
		cm:                  cm,
		creds:               creds,
		expChecker:          systemTimeExpirationChecker{},
		mfaPub:              mfaPub,
		logoutPub:           logoutPub,
		errPub:              errPub,
		accountUpdateEvents: accountUpdateEvents,
		sessionStores:       sessionStores,
	}
}

// IsLoggedIn reports user login status.
//
// Thread safe.
func (r *RenewingChecker) IsLoggedIn() (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, store := range r.sessionStores {
		if err := store.Renew(); err != nil {
			return false, fmt.Errorf("renewing session: %w", err)
		}
	}

	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		return false, err
	}

	return cfg.AutoConnectData.ID != 0 && len(cfg.TokensData) > 0, nil
}

// IsMFAEnabled checks if user account has MFA turned on.
//
// Thread safe.
func (r *RenewingChecker) IsMFAEnabled() (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.isMFAEnabled()
}

func (r *RenewingChecker) isMFAEnabled() (bool, error) {
	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		extraErr := fmt.Errorf("checking MFA status, loading config: %w", err)
		r.errPub.Publish(extraErr)
		return false, extraErr
	}

	resp, err := r.creds.MultifactorAuthStatus()
	if err != nil {
		extraErr := fmt.Errorf("querying MFA status: %w", err)
		r.errPub.Publish(extraErr)
		return false, extraErr
	}

	// inform subscribers
	r.mfaPub.Publish(resp.Status == internal.MFAEnabledStatusName)

	return resp.Status == internal.MFAEnabledStatusName, nil
}

// IsVPNExpired is used to check whether the user is allowed to use VPN
func (r *RenewingChecker) IsVPNExpired() (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		return true, fmt.Errorf("loading config: %w", err)
	}

	data, ok := cfg.TokensData[cfg.AutoConnectData.ID]
	if !ok {
		return true, fmt.Errorf("there is no token stored for user id: %v", cfg.AutoConnectData.ID)
	}

	if r.expChecker.IsExpired(data.ServiceExpiry) {
		if err := r.fetchSaveServices(cfg.AutoConnectData.ID, &data); err != nil {
			return true, fmt.Errorf("updating service expiry token: %w", err)
		}
	}

	return r.expChecker.IsExpired(data.ServiceExpiry), nil
}

func (r *RenewingChecker) GetDedicatedIPServices() ([]DedicatedIPService, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	services, err := r.fetchServices()
	if err != nil {
		return nil, fmt.Errorf("fetching available services: %w", err)
	}

	dipServices := []DedicatedIPService{}
	for _, service := range services {
		if service.Service.ID == DedicatedIPServiceID && !r.expChecker.IsExpired(service.ExpiresAt) {
			serverIDs := []int64{}
			for _, server := range service.Details.Servers {
				serverIDs = append(serverIDs, server.ID)
			}
			dipServices = append(dipServices,
				DedicatedIPService{ExpiresAt: service.ExpiresAt, ServerIDs: serverIDs})
		}
	}

	return dipServices, nil
}

// fetchSaveServices fetches services and updates data appropriately
func (r *RenewingChecker) fetchSaveServices(userID int64, data *config.TokenData) error {
	services, err := r.creds.Services()
	if err != nil {
		return err
	}

	var vpnExpiry string
	for _, service := range services {
		if service.Service.ID == VPNServiceID {
			vpnExpiry = service.ExpiresAt
			break
		}
	}

	if vpnExpiry == "" {
		return fmt.Errorf("vpn service not found for user %d", userID)
	}

	data.ServiceExpiry = vpnExpiry

	if err := r.cm.SaveWith(saveVpnExpirationDate(userID, *data)); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}
	r.accountUpdateEvents.SubscriptionUpdate.Publish(&pb.AccountModification{
		ExpiresAt: &data.ServiceExpiry,
	})

	return nil
}

func (r *RenewingChecker) fetchServices() ([]core.ServiceData, error) {
	services, err := r.creds.Services()
	if err != nil {
		return nil, fmt.Errorf("fetching available services: %w", err)
	}
	return services, nil
}

func saveVpnExpirationDate(userID int64, data config.TokenData) config.SaveFunc {
	return func(c config.Config) config.Config {
		user := c.TokensData[userID]
		user.ServiceExpiry = data.ServiceExpiry
		c.TokensData[userID] = user
		return c
	}
}

// Logout the user.
func Logout(user int64, logoutPub events.Publisher[events.DataAuthorization]) config.SaveFunc {
	return func(c config.Config) config.Config {
		if logoutPub != nil {
			// register stats instant logout with status success
			logoutPub.Publish(events.DataAuthorization{
				DurationMs: -1, EventTrigger: events.TriggerApp, EventStatus: events.StatusSuccess})
		}
		delete(c.TokensData, user)
		return c
	}
}
