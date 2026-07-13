package daemon

import (
	"github.com/NordSecurity/nordvpn-linux/log"
	"github.com/NordSecurity/nordvpn-linux/nc"
)

// StartNC tries to start notification client and logs any errors if they occur.
// This is just a convenience wrapper, we always start notification client in
// another goroutine, so we cannot handle the errors directly in the caller.
func StartNC(ncClient nc.NotificationClient) {
	if err := ncClient.Start(); err != nil {
		log.NC.Errorf("starting notification client: %s", err)
	}
}
