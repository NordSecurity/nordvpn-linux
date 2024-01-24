package daemon

import (
	"log"

	"github.com/NordSecurity/nordvpn-linux/nc"
)

// StartNC tries to start notification client and logs any errors if they occur. This is just a convenience wrapper, we
// always start notification client in another goroutine, so we cannot handle the errors directly in the caller. Prefix
// will be prepended to the error log.
func StartNC(prefix string, ncClient nc.NotificationClient) {
	if err := ncClient.Start(); err != nil {
		log.Printf("%s starting notification client: %s", prefix, err)
	}
}
