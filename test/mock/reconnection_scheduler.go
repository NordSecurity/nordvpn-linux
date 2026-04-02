package mock

import "time"

type PauseSchedulerMock struct {
	ConnectionScheduled bool
	PauseDuration       time.Duration
}

func (p *PauseSchedulerMock) ScheduleReconnection(duration time.Duration) {
	p.ConnectionScheduled = true
	p.PauseDuration = duration
}

func (p *PauseSchedulerMock) CancelReconnection() {
	p.ConnectionScheduled = false
	p.PauseDuration = 0
}
