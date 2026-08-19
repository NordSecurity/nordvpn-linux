package alert

import "github.com/NordSecurity/nordvpn-linux/internal"

type AlertBuilder struct {
	send  func(Alert)
	alert Alert
}

func NewAlertBuilder(send func(Alert), body string) *AlertBuilder {
	return &AlertBuilder{
		send: send,
		alert: Alert{
			Summary: internal.AppName,
			Body:    body,
			Urgency: UrgencyNormal,
		},
	}
}

func (b *AlertBuilder) Summary(summary string) *AlertBuilder {
	b.alert.Summary = summary
	return b
}

func (b *AlertBuilder) Urgent() *AlertBuilder {
	b.alert.Urgency = UrgencyCritical
	return b
}

func (b *AlertBuilder) Action(key, label string, callback func()) *AlertBuilder {
	b.alert.Actions = append(b.alert.Actions, Action{
		Key:      key,
		Label:    label,
		Callback: callback,
	})
	return b
}

func (b *AlertBuilder) Show() {
	b.send(b.alert)
}
