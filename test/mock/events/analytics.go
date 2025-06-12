package events

import "github.com/NordSecurity/nordvpn-linux/config"

type Analytics struct {
	EnableErr error
	DisablErr error

	State config.AnalyticsConsent
}

func NewAnalytics(currentState config.AnalyticsConsent) Analytics {
	return Analytics{
		State: currentState,
	}
}

func (a *Analytics) Enable() error {
	if a.EnableErr != nil {
		return a.EnableErr
	}

	a.State = config.ConsentGranted

	return nil
}

func (a *Analytics) Disable() error {
	if a.DisablErr != nil {
		return a.DisablErr
	}

	a.State = config.ConsentDenied

	return nil
}
