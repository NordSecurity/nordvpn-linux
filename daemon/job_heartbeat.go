package daemon

import (
	"time"

	"github.com/NordSecurity/nordvpn-linux/events"
)

// JobHeartBeat sends heart beats.
func JobHeartBeat(
	publisher events.Publisher[time.Duration],
	period time.Duration,
) func() {
	return func() {
		publisher.Publish(period)
	}
}
