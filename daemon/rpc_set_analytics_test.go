package daemon

import (
	"context"
	"fmt"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	analyticsMock "github.com/NordSecurity/nordvpn-linux/test/mock/events"
	"github.com/stretchr/testify/assert"
)

func TestSetAnalytics(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name                  string
		initialConsentLevel   config.AnalyticsConsent
		requestedConsentLevel bool
		enableErr             error
		disableErr            error
		configLoadErr         error
		configSaveErr         error
		expectedConsentLevel  config.AnalyticsConsent
		expectedResponse      int64
	}{
		{
			name:                  "enable consent first time",
			initialConsentLevel:   config.ConsentUndefined,
			requestedConsentLevel: true,
			expectedConsentLevel:  config.ConsentGranted,
			expectedResponse:      internal.CodeSuccess,
		},
		{
			name:                  "disable consent first time",
			initialConsentLevel:   config.ConsentUndefined,
			requestedConsentLevel: false,
			expectedConsentLevel:  config.ConsentDenied,
			expectedResponse:      internal.CodeSuccess,
		},
		{
			name:                  "consent enabled to disabled",
			initialConsentLevel:   config.ConsentGranted,
			requestedConsentLevel: false,
			expectedConsentLevel:  config.ConsentDenied,
			expectedResponse:      internal.CodeSuccess,
		},
		{
			name:                  "consent disabled to enabled",
			initialConsentLevel:   config.ConsentDenied,
			requestedConsentLevel: true,
			expectedConsentLevel:  config.ConsentGranted,
			expectedResponse:      internal.CodeSuccess,
		},
		{
			name:                  "enabled to enabled",
			initialConsentLevel:   config.ConsentGranted,
			requestedConsentLevel: true,
			expectedConsentLevel:  config.ConsentGranted,
			expectedResponse:      internal.CodeNothingToDo,
		},
		{
			name:                  "disabled to disabled",
			initialConsentLevel:   config.ConsentDenied,
			requestedConsentLevel: false,
			expectedConsentLevel:  config.ConsentDenied,
			expectedResponse:      internal.CodeNothingToDo,
		},
		{
			name:                  "disabled to disabled",
			initialConsentLevel:   config.ConsentDenied,
			requestedConsentLevel: false,
			expectedConsentLevel:  config.ConsentDenied,
			expectedResponse:      internal.CodeNothingToDo,
		},
		{
			name:                  "config load error",
			initialConsentLevel:   config.ConsentUndefined,
			requestedConsentLevel: true,
			configLoadErr:         fmt.Errorf("cfg load err"),
			expectedConsentLevel:  config.ConsentUndefined,
			expectedResponse:      internal.CodeConfigError,
		},
		{
			name:                  "config save error",
			initialConsentLevel:   config.ConsentUndefined,
			requestedConsentLevel: true,
			configSaveErr:         fmt.Errorf("cfg save err"),
			expectedConsentLevel:  config.ConsentUndefined,
			expectedResponse:      internal.CodeConfigError,
		},
		{
			name:                  "enable analytics error",
			initialConsentLevel:   config.ConsentUndefined,
			requestedConsentLevel: true,
			enableErr:             fmt.Errorf("enable analytics error"),
			expectedConsentLevel:  config.ConsentUndefined,
			expectedResponse:      internal.CodeConfigError,
		},
		{
			name:                  "disable analytics error",
			initialConsentLevel:   config.ConsentUndefined,
			requestedConsentLevel: false,
			disableErr:            fmt.Errorf("disable analytics error"),
			expectedConsentLevel:  config.ConsentUndefined,
			expectedResponse:      internal.CodeConfigError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfgMock := mock.NewMockConfigManager()
			cfgMock.Cfg.AnalyticsConsent = test.initialConsentLevel
			cfgMock.LoadErr = test.configLoadErr
			cfgMock.SaveErr = test.configSaveErr

			analyticsMock := analyticsMock.NewAnalytics(test.initialConsentLevel)
			analyticsMock.EnableErr = test.enableErr
			analyticsMock.DisablErr = test.disableErr

			r := RPC{
				analytics: &analyticsMock,
				cm:        cfgMock,
			}

			response, err := r.SetAnalytics(context.Background(), &pb.SetGenericRequest{
				Enabled: test.requestedConsentLevel,
			})

			assert.NoError(t, err, "Unexpected error when making the RPC request.")
			assert.Equal(t, test.expectedResponse, response.Type, "Unexpected response to the RPC.")
			assert.Equal(t, test.expectedConsentLevel, test.expectedConsentLevel, analyticsMock.State)
		})
	}
}
