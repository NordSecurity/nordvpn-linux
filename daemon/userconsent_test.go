package daemon

import (
	"errors"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"github.com/NordSecurity/nordvpn-linux/test/mock/auth"
	"github.com/NordSecurity/nordvpn-linux/test/mock/events"
	"github.com/NordSecurity/nordvpn-linux/test/mock/insights"
	"github.com/stretchr/testify/assert"
)

func TestModeForCountryCode(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		code     core.CountryCode
		expected consentMode
	}{
		{core.NewCountryCode("us"), consentModeStandard},
		{core.NewCountryCode("ca"), consentModeStandard},
		{core.NewCountryCode("jp"), consentModeStandard},
		{core.NewCountryCode("au"), consentModeStandard},
		{core.NewCountryCode("fr"), consentModeGDPR},
		{core.NewCountryCode("zz"), consentModeGDPR},
		{core.NewCountryCode(""), consentModeGDPR},
	}

	for _, tt := range tests {
		t.Run(tt.code.String(), func(t *testing.T) {
			mode := modeForCountryCode(tt.code)
			assert.Equal(t, mode, tt.expected)
		})
	}
}

func TestIsConsentFlowCompleted(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		manager  config.Manager
		expected bool
	}{
		{
			name:     "Load error -> false",
			manager:  &mock.ConfigManager{LoadErr: errors.New("load failure")},
			expected: false,
		},
		{
			name:     "AnalyticsConsent none -> false",
			manager:  &mock.ConfigManager{Cfg: &config.Config{AnalyticsConsent: config.ConsentUndefined}},
			expected: false,
		},
		{
			name:     "AnalyticsConsent != None -> true",
			manager:  &mock.ConfigManager{Cfg: &config.Config{AnalyticsConsent: config.ConsentGranted}},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analytics := events.NewAnalytics(config.ConsentUndefined)
			consentChecker := NewConsentChecker(
				false,
				tt.manager,
				&insights.InsightsMock{},
				&auth.AuthCheckerMock{},
				&analytics,
			)
			got := consentChecker.IsConsentFlowCompleted()
			assert.Equal(t, got, tt.expected)
		})
	}
}

func TestConsentModeFromUserLocation(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name          string
		initialConfig config.Config
		loadErr       error
		apiInsights   *core.Insights
		apiErr        error
		authLoggedIn  bool
		expected      consentMode
	}{
		{
			name:     "config Load error -> GDPR",
			loadErr:  errors.New("load failure"),
			expected: consentModeGDPR,
		},
		{
			name:          "KillSwitch true -> GDPR",
			initialConfig: config.Config{KillSwitch: true},
			apiInsights:   &core.Insights{CountryCode: "us"},
			expected:      consentModeGDPR,
		},
		{
			name:          "API.Insights error -> GDPR",
			initialConfig: config.Config{KillSwitch: false},
			apiErr:        errors.New("api failure"),
			expected:      consentModeGDPR,
		},
		{
			name:          "API returns nil insights -> GDPR",
			initialConfig: config.Config{KillSwitch: false},
			apiInsights:   nil,
			expected:      consentModeGDPR,
		},
		{
			name:          "standard country US -> Standard",
			initialConfig: config.Config{KillSwitch: false},
			apiInsights:   &core.Insights{CountryCode: "US"},
			expected:      consentModeStandard,
		},
		{
			name:          "non-standard country FR -> GDPR",
			initialConfig: config.Config{KillSwitch: false},
			apiInsights:   &core.Insights{CountryCode: "FR"},
			expected:      consentModeGDPR,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := &mock.ConfigManager{Cfg: &tt.initialConfig, LoadErr: tt.loadErr}
			api := &insights.InsightsMock{InsightsResult: tt.apiInsights, Err: tt.apiErr}
			authChk := &auth.AuthCheckerMock{LoggedIn: tt.authLoggedIn}

			acc := &AnalyticsConsentChecker{cm: cm, insightsAPI: api, authChecker: authChk}
			got := acc.consentModeFromUserLocation()
			assert.Equal(t, got, tt.expected)
		})
	}
}

func TestSetConsentGranted(t *testing.T) {
	tests := []struct {
		name              string
		initialState      config.AnalyticsConsent
		enableErr         error
		saveErr           error
		expectedErr       error
		expectedState     config.AnalyticsConsent
		expectedSaved     bool
		expectedConsentIn config.AnalyticsConsent
	}{
		{
			name:              "analytics is enabled and config is saved on success",
			initialState:      config.ConsentUndefined,
			expectedErr:       nil,
			expectedState:     config.ConsentGranted,
			expectedSaved:     true,
			expectedConsentIn: config.ConsentGranted,
		},
		{
			name:          "analytics enable fails -> config is not saved",
			initialState:  config.ConsentUndefined,
			enableErr:     errors.New("enable error"),
			expectedErr:   errors.New("enable error"),
			expectedState: config.ConsentUndefined,
			expectedSaved: false,
		},
		{
			name:              "config save fails",
			initialState:      config.ConsentUndefined,
			saveErr:           errors.New("save error"),
			expectedErr:       errors.New("save error"),
			expectedState:     config.ConsentGranted,
			expectedSaved:     false,
			expectedConsentIn: config.ConsentUndefined,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analytics := &events.Analytics{
				State:     tt.initialState,
				EnableErr: tt.enableErr,
			}
			cm := mock.NewMockConfigManager()
			cm.SaveErr = tt.saveErr
			acc := &AnalyticsConsentChecker{
				analytics: analytics,
				cm:        cm,
			}

			err := acc.setConsentGranted()

			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedState, analytics.State)
			assert.Equal(t, tt.expectedSaved, cm.Saved)

			if cm.Cfg != nil {
				assert.Equal(t, tt.expectedConsentIn, cm.Cfg.AnalyticsConsent)
			}
		})
	}
}

func TestDoLightLogout(t *testing.T) {
	category.Set(t, category.Unit)
	cfg := config.Config{
		TokensData:      map[int64]config.TokenData{42: {}},
		AutoConnectData: config.AutoConnectData{ID: 42},
	}
	cm := &mock.ConfigManager{Cfg: &cfg}
	analytics := events.NewAnalytics(config.ConsentUndefined)
	acc := &AnalyticsConsentChecker{cm: cm, analytics: &analytics}
	assert.False(t, cm.Saved)
	assert.True(t, len(cm.Cfg.TokensData) > 0)
	assert.True(t, cm.Cfg.AutoConnectData.ID != 0)

	err := acc.doLightLogout()

	assert.NoError(t, err)
	assert.True(t, cm.Saved)
	assert.Zero(t, len(cm.Cfg.TokensData))
	assert.Zero(t, cm.Cfg.AutoConnectData.ID)
}

func TestPrepareDaemonIfConsentNotCompleted(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name               string
		initialConfig      config.Config
		loadErr            error
		apiInsights        core.InsightsAPI
		apiErr             error
		authLoggedIn       bool
		expectedSaved      bool
		expectedConsentSet bool
		expectedConsentVal config.AnalyticsConsent
		expectedTokensLen  int
		expectedAutoConnID int64
	}{
		{
			name:               "consent already completed",
			initialConfig:      config.Config{AnalyticsConsent: config.ConsentGranted},
			expectedSaved:      false,
			expectedConsentSet: true,
			expectedConsentVal: config.ConsentGranted,
			expectedTokensLen:  0,
			expectedAutoConnID: 0,
		},
		{
			name:               "standard country -> set consent",
			initialConfig:      config.Config{AnalyticsConsent: config.ConsentUndefined},
			apiInsights:        &insights.InsightsMock{InsightsResult: &core.Insights{CountryCode: "ca"}},
			expectedSaved:      true,
			expectedConsentSet: true,
			expectedConsentVal: config.ConsentGranted,
			expectedTokensLen:  0,
			expectedAutoConnID: 0,
		},
		{
			name: "GDPR country & logged in -> do light logout",
			initialConfig: config.Config{
				AnalyticsConsent: config.ConsentUndefined,
				TokensData:       map[int64]config.TokenData{42: {}},
				AutoConnectData:  config.AutoConnectData{ID: 42},
			},
			apiInsights:        &insights.InsightsMock{InsightsResult: &core.Insights{CountryCode: "fr"}},
			authLoggedIn:       true,
			expectedSaved:      true,
			expectedConsentSet: false,
			expectedTokensLen:  0,
			expectedAutoConnID: 0,
		},
		{
			name:               "GDPR country & not logged in",
			initialConfig:      config.Config{AnalyticsConsent: config.ConsentUndefined},
			apiInsights:        &insights.InsightsMock{InsightsResult: &core.Insights{CountryCode: "fr"}},
			authLoggedIn:       false,
			expectedSaved:      false,
			expectedConsentSet: false,
			expectedTokensLen:  0,
			expectedAutoConnID: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := &mock.ConfigManager{Cfg: &tt.initialConfig, LoadErr: tt.loadErr}
			api := tt.apiInsights
			authChk := &auth.AuthCheckerMock{LoggedIn: tt.authLoggedIn}
			analytics := events.NewAnalytics(config.ConsentUndefined)

			acc := &AnalyticsConsentChecker{cm: cm, insightsAPI: api, authChecker: authChk, analytics: &analytics}
			acc.PrepareDaemonIfConsentNotCompleted()

			assert.Equal(t, cm.Saved, tt.expectedSaved)
			if tt.expectedConsentSet {
				assert.NotNil(t, cm.Cfg.AnalyticsConsent)
				assert.Equal(t, cm.Cfg.AnalyticsConsent, tt.expectedConsentVal)
			} else {
				assert.Equal(t, cm.Cfg.AnalyticsConsent, config.ConsentUndefined)
			}
			assert.Equal(t, len(cm.Cfg.TokensData), tt.expectedTokensLen)
			assert.Equal(t, cm.Cfg.AutoConnectData.ID, tt.expectedAutoConnID)
		})
	}
}
