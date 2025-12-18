package events

import "github.com/NordSecurity/nordvpn-linux/config"

type Analytics struct {
	EnableErr error
	DisablErr error
	InitErr   error

	State          config.AnalyticsConsent
	InitCalled     bool
	InitCalledWith config.AnalyticsConsent
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

func (a *Analytics) Init(consent config.AnalyticsConsent) error {
	a.InitCalled = true
	a.InitCalledWith = consent
	if a.InitErr != nil {
		return a.InitErr
	}
	return nil
}
