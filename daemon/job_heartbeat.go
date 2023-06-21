package daemon

// JobHeartBeat sends heart beats.
func JobHeartBeat(
	timePeriod int,
	events *Events,
) func() {
	return func() {
		events.Service.HeartBeat.Publish(timePeriod)
	}
}
