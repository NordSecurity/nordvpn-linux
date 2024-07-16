package daemon

import (
	"github.com/NordSecurity/nordvpn-linux/daemon/events"
)

// JobHeartBeat sends heart beats.
func JobHeartBeat(
	timePeriod int,
	events *events.Events,
) func() {
	return func() {
		events.Service.HeartBeat.Publish(timePeriod)
	}
}
