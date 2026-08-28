package alert

import "github.com/NordSecurity/nordvpn-linux/log"

type NoopNotifier struct{}

func (NoopNotifier) Alert(body string) *AlertBuilder {
	return NewAlertBuilder(func(a Alert) bool {
		log.Warnf("notifier unavailable, dropping alert: %s", a)
		return false
	}, body)
}

func (NoopNotifier) Mute() {}

func (NoopNotifier) Unmute() {}

func (NoopNotifier) Close() error { return nil }
