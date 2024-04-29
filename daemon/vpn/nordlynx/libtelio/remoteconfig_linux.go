//go:build !firebase

package libtelio

import "github.com/NordSecurity/nordvpn-linux/config"

type TelioRemoteConfigFetcher struct {
	cm config.Manager
}

func (c *TelioRemoteConfigFetcher) IsAvailable() bool {
	return false
}

func (c *TelioRemoteConfigFetcher) Fetch(firebaseToken string, appVer string) (string, error) {
	return "", nil
}
