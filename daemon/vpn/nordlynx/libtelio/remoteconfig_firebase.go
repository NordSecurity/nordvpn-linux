//go:build firebase

package libtelio

import (
	"log"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/config/remote"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

type TelioRemoteConfigFetcher struct {
	rc *remote.RConfig
	cm config.Manager
}

func (c *TelioRemoteConfigFetcher) IsAvailable() bool {
	return true
}

func (c *TelioRemoteConfigFetcher) Fetch(appVer string) (string, error) {
	if c.rc == nil {
		log.Println(internal.InfoPrefix, "Initialize firebase")
		c.rc = remote.NewRConfig(remote.UpdatePeriod, remote.NewFirebaseService("firebaseToken"), c.cm)
	}
	log.Println(internal.InfoPrefix, "Fetch libtelio remote config")
	return c.rc.GetTelioConfig(appVer)
}
