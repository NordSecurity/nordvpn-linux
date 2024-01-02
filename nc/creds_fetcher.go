package nc

import (
	"fmt"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
)

type TimeSource interface {
	GetDurationSinceTimestamp(int64) time.Duration
}

type RealTime struct{}

func (r RealTime) GetDurationSinceTimestamp(timestamp int64) time.Duration {
	return time.Since(time.Unix(timestamp, 0))
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
		ncData.Password != "" ||
		ncData.IssuedTimestamp != 0
}

// GetCredentials loads credentials from local config and falls back to fetching them from the core api if they are outdated
// or invalid.
func (cf *CredentialsGetter) GetCredentials() (config.NCData, error) {
	var cfg config.Config
	if err := cf.cm.Load(&cfg); err != nil {
		return config.NCData{}, fmt.Errorf("reading cfg: %w", err)
	}
	userID := cfg.AutoConnectData.ID
	tokenData := cfg.TokensData[userID]
	ncData := tokenData.NCData

	const tokenValidityPeriod time.Duration = 86400 * time.Second
	elapsedSinceTokenIssued := cf.timeSource.GetDurationSinceTimestamp(ncData.IssuedTimestamp)

	if elapsedSinceTokenIssued < tokenValidityPeriod && areNCCredentialsValid(ncData) {
		return ncData, nil
	}

	credentials, err := core.GetNCCredentials(cf.api, tokenData.Token, ncData.UserID)
	if err != nil {
		return config.NCData{}, fmt.Errorf("getting NC credentials: %w", err)
	}

	return credentials, cf.cm.SaveWith(func(c config.Config) config.Config {
		user := c.TokensData[userID]
		user.NCData = credentials
		c.TokensData[userID] = user

		return c
	})
}
