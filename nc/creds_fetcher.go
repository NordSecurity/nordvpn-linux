package nc

import (
	"fmt"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/google/uuid"
)

type TimeSource interface {
	GetDurationSinceTimestamp(time.Time) time.Duration
}

type RealTime struct{}

func (r RealTime) GetDurationSinceTimestamp(timestamp time.Time) time.Duration {
	return time.Since(timestamp)
}

type CredentialsGetter struct {
	api        core.CredentialsAPI
	cm         config.Manager
	timeSource TimeSource
}

func NewCredsFetcher(api core.CredentialsAPI, cm config.Manager, time TimeSource) CredentialsGetter {
	return CredentialsGetter{
		api:        api,
		cm:         cm,
		timeSource: time,
	}
}

func areNCCredentialsValid(ncData config.NCData) bool {
	return !ncData.IsUserIDEmpty() ||
		ncData.Endpoint != "" ||
		ncData.Username != "" ||
		ncData.Password != ""
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
	}

	resp, err := cf.api.NotificationCredentials(tokenData.Token, ncData.UserID.String())
	if err != nil {
		return config.NCData{}, fmt.Errorf("getting NC credentials: %w", err)
	}

	ncData.Endpoint = resp.Endpoint
	ncData.Username = resp.Username
	ncData.Password = resp.Password

	return ncData, cf.cm.SaveWith(func(c config.Config) config.Config {
		user := c.TokensData[userID]
		user.NCData = ncData
		c.TokensData[userID] = user

		return c
	})
}
