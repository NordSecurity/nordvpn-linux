//go:build firebase

package libtelio

import (
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/config/remote"
)

type TelioRemoteConfigFetcher struct {
	rc remote.RemoteConfigGetter
}

func (c *TelioRemoteConfigFetcher) IsAvailable() bool {
	return true
}

func (c *TelioRemoteConfigFetcher) Fetch(appVer string) (string, error) {
	if c.rc == nil {
		return "", fmt.Errorf("missing remote config")
	}
	return c.rc.GetTelioConfig(appVer)
}
