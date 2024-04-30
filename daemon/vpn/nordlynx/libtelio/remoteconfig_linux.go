//go:build !firebase

package libtelio

import "github.com/NordSecurity/nordvpn-linux/config/remote"

type TelioRemoteConfigFetcher struct {
	rc remote.RemoteConfigGetter
}

func (c *TelioRemoteConfigFetcher) IsAvailable() bool {
	return false
}

func (c *TelioRemoteConfigFetcher) Fetch(appVer string) (string, error) {
	return "", nil
}
