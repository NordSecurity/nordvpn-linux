package nc

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
)

const credentialsValidityPeriod = 24 * time.Hour

type CredentialsGetter struct {
	api core.CredentialsAPI
	cm  config.Manager
}

func NewCredsFetcher(api core.CredentialsAPI, cm config.Manager) CredentialsGetter {
	return CredentialsGetter{
		api: api,
		cm:  cm,
	}
}

func areNCCredentialsValid(ncData config.NCData) bool {
	return !ncData.IsUserIDEmpty() &&
		ncData.Endpoint != "" &&
		ncData.Username != "" &&
		ncData.Password != "" &&
		!ncData.ExpirationDate.IsZero()
}

var ErrInvalidCredentials = fmt.Errorf("stored credentials are not valid")

// GetCredentialsFromConfig loads credentials from local config
func (cf *CredentialsGetter) GetCredentialsFromConfig() (config.NCData, error) {
	var cfg config.Config
	if err := cf.cm.Load(&cfg); err != nil {
		return config.NCData{}, fmt.Errorf("reading cfg: %w", err)
	}
	userID := cfg.AutoConnectData.ID
	tokenData := cfg.TokensData[userID]
	ncData := tokenData.NCData

	if !areNCCredentialsValid(ncData) {
		return config.NCData{}, ErrInvalidCredentials
	}

	return ncData, nil
}

// GetCredentials fetches credentials from core API and saves them in local config
func (cf *CredentialsGetter) GetCredentialsFromAPI() (config.NCData, error) {
	var cfg config.Config
	if err := cf.cm.Load(&cfg); err != nil {
		return config.NCData{}, fmt.Errorf("reading cfg: %w", err)
	}
	userID := cfg.AutoConnectData.ID
	tokenData := cfg.TokensData[userID]
	ncData := tokenData.NCData

	if ncData.IsUserIDEmpty() {
		ncData.UserID = uuid.New()
		err := cf.cm.SaveWith(func(c config.Config) config.Config {
			user := c.TokensData[userID]
			user.NCData = ncData
			c.TokensData[userID] = user

			return c
		})
		if err != nil {
			return ncData, fmt.Errorf("saving new generated user id: %w", err)
		}
	}

	resp, err := cf.api.NotificationCredentials(ncData.UserID.String())
	if err != nil {
		return config.NCData{}, fmt.Errorf("getting NC credentials: %w", err)
	}

	ncData.Endpoint = resp.Endpoint
	ncData.Username = resp.Username
	ncData.Password = resp.Password
	ncData.ExpirationDate = time.Now().Add(credentialsValidityPeriod)

	return ncData, cf.cm.SaveWith(func(c config.Config) config.Config {
		user := c.TokensData[userID]
		user.NCData = ncData
		c.TokensData[userID] = user

		return c
	})
}

// RevokeCredentials revokes credentials
func (cf *CredentialsGetter) RevokeCredentials(purgeSession bool) (bool, error) {
	var cfg config.Config
	if err := cf.cm.Load(&cfg); err != nil {
		return false, fmt.Errorf("reading cfg: %w", err)
	}
	userID := cfg.AutoConnectData.ID
	tokenData := cfg.TokensData[userID]
	ncData := tokenData.NCData

	if !areNCCredentialsValid(ncData) {
		return false, ErrInvalidCredentials
	}

	resp, err := cf.api.NotificationCredentialsRevoke(tokenData.NCData.UserID.String(), purgeSession)
	if err != nil {
		return false, fmt.Errorf("error revoking token: %w", err)
	}
	if strings.ToLower(resp.Status) == "ok" {
		return true, nil
	} else {
		return false, fmt.Errorf("response status: %s", resp.Status)
	}
}
