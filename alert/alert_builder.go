package alert

import "github.com/NordSecurity/nordvpn-linux/internal"

type AlertBuilder struct {
	send  AlertSender
	alert Alert
}

type AlertSender func(Alert) bool

func NewAlertBuilder(send AlertSender, body string) *AlertBuilder {
	return &AlertBuilder{
		send: send,
		alert: Alert{
			Summary: internal.AppName,
			Body:    body,
			Urgency: UrgencyNormal,
			OnShown: nil,
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

func (b *AlertBuilder) OnShown(onShown func()) *AlertBuilder {
	b.alert.OnShown = onShown
	return b
}

func (b *AlertBuilder) Show() {
	if b.send(b.alert) && b.alert.OnShown != nil {
		b.alert.OnShown()
	}
}
