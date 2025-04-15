package daemon

import (
	"time"

	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/events"
)

// JobHeartBeat sends heart beats.
func JobHeartBeat(
	authChecker auth.Checker,
	publisher events.Publisher[time.Duration],
	period time.Duration,
) func() {
	return func() {
		if authChecker.IsLoggedIn() {
			publisher.Publish(period)
		}
	}
}
