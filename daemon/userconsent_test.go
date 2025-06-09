package daemon

import (
	"errors"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"github.com/NordSecurity/nordvpn-linux/test/mock/auth"
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
			manager:  &mock.ConfigManager{Cfg: &config.Config{AnalyticsConsent: config.ConsentMode_NONE}},
			expected: false,
		},
		{
			name:     "AnalyticsConsent != None -> true",
			manager:  &mock.ConfigManager{Cfg: &config.Config{AnalyticsConsent: config.ConsentMode_ALLOWED}},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			consentChecker := NewConsentChecker(false, tt.manager, &insights.InsightsMock{}, &auth.AuthCheckerMock{})
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

func TestSetConsentTrue(t *testing.T) {
	category.Set(t, category.Unit)
	cm := &mock.ConfigManager{Cfg: &config.Config{AnalyticsConsent: config.ConsentMode_NONE}}
	acc := &AnalyticsConsentChecker{cm: cm}
	assert.False(t, cm.Saved)
	assert.Equal(t, cm.Cfg.AnalyticsConsent, config.ConsentMode_NONE)

	err := acc.setConsentAllowed()

	assert.NoError(t, err)
	assert.True(t, cm.Saved)
	assert.NotNil(t, cm.Cfg.AnalyticsConsent)
	assert.Equal(t, cm.Cfg.AnalyticsConsent, config.ConsentMode_ALLOWED)
}

func TestDoLightLogout(t *testing.T) {
	category.Set(t, category.Unit)
	cfg := config.Config{
		TokensData:      map[int64]config.TokenData{42: {}},
		AutoConnectData: config.AutoConnectData{ID: 42},
	}
	cm := &mock.ConfigManager{Cfg: &cfg}
	acc := &AnalyticsConsentChecker{cm: cm}
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
		expectedConsentVal config.ConsentMode
		expectedTokensLen  int
		expectedAutoConnID int64
	}{
		{
			name:               "consent already completed",
			initialConfig:      config.Config{AnalyticsConsent: config.ConsentMode_ALLOWED},
			expectedSaved:      false,
			expectedConsentSet: true,
			expectedConsentVal: config.ConsentMode_ALLOWED,
			expectedTokensLen:  0,
			expectedAutoConnID: 0,
		},
		{
			name:               "standard country -> set consent",
			initialConfig:      config.Config{AnalyticsConsent: config.ConsentMode_NONE},
			apiInsights:        &insights.InsightsMock{InsightsResult: &core.Insights{CountryCode: "ca"}},
			expectedSaved:      true,
			expectedConsentSet: true,
			expectedConsentVal: config.ConsentMode_ALLOWED,
			expectedTokensLen:  0,
			expectedAutoConnID: 0,
		},
		{
			name: "GDPR country & logged in -> do light logout",
			initialConfig: config.Config{
				AnalyticsConsent: config.ConsentMode_NONE,
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
			initialConfig:      config.Config{AnalyticsConsent: config.ConsentMode_NONE},
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

			acc := &AnalyticsConsentChecker{cm: cm, insightsAPI: api, authChecker: authChk}
			acc.PrepareDaemonIfConsentNotCompleted()

			assert.Equal(t, cm.Saved, tt.expectedSaved)
			if tt.expectedConsentSet {
				assert.NotNil(t, cm.Cfg.AnalyticsConsent)
				assert.Equal(t, cm.Cfg.AnalyticsConsent, tt.expectedConsentVal)
			} else {
				assert.Equal(t, cm.Cfg.AnalyticsConsent, config.ConsentMode_NONE)
			}
			assert.Equal(t, len(cm.Cfg.TokensData), tt.expectedTokensLen)
			assert.Equal(t, cm.Cfg.AutoConnectData.ID, tt.expectedAutoConnID)
		})
	}
}
