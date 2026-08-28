package alert

import (
	"time"

	"github.com/NordSecurity/nordvpn-linux/alert"
)

type Alert struct {
	ID uint32
	alert.Alert
}

type NotifierMock struct {
	Alerts   []Alert
	NextID   uint32
	UpdateCh chan struct{}
}

func (nm *NotifierMock) Wait() {
	select {
	case <-time.After(1 * time.Second):
	case <-nm.UpdateCh:
	}
}

func (nm *NotifierMock) Alert(body string) *alert.AlertBuilder {
	return alert.NewAlertBuilder(nm.record, body)
}

func (nm *NotifierMock) record(alert alert.Alert) bool {
	alertID := nm.NextID
	nm.Alerts = append(nm.Alerts, Alert{ID: alertID, Alert: alert})
	nm.NextID++

	if nm.UpdateCh != nil {
		select {
		case nm.UpdateCh <- struct{}{}:
		default:
			<-nm.UpdateCh
			nm.UpdateCh <- struct{}{}
		}
	}

	return true
}

func (nm *NotifierMock) Mute() {}

func (nm *NotifierMock) Unmute() {}

func (nm *NotifierMock) Close() error {
	return nil
}

func (nm *NotifierMock) GetLastNotification() Alert {
	return nm.Alerts[len(nm.Alerts)-1]
}
