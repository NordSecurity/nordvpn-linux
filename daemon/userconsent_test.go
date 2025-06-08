package daemon

import (
	"errors"
	"strings"
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
		code     countryCode
		expected consentMode
	}{
		{countryCode("us"), consentModeStandard},
		{countryCode("ca"), consentModeStandard},
		{countryCode("jp"), consentModeStandard},
		{countryCode("au"), consentModeStandard},
		{countryCode("fr"), consentModeGDPR},
		{countryCode("zz"), consentModeGDPR},
		{countryCode(""), consentModeGDPR},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			mode := modeForCountryCode(tt.code)
			assert.Equal(t, mode, tt.expected)
		})
	}
}

func TestModeForCountryCode_CaseInsensitive(t *testing.T) {
	category.Set(t, category.Unit)

	codes := []string{"US", "Us", "uS", "us"}
	for _, codeStr := range codes {
		t.Run(codeStr, func(t *testing.T) {
			cc := countryCode(strings.ToLower(codeStr))
			mode := modeForCountryCode(cc)
			assert.Equal(t, mode, consentModeStandard)
		})
	}
}

func TestIsConsentFlowCompleted(t *testing.T) {
	category.Set(t, category.Unit)

	enabled := true
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
			name:     "AnalyticsConsent nil -> false",
			manager:  &mock.ConfigManager{Cfg: &config.Config{AnalyticsConsent: nil}},
			expected: false,
		},
		{
			name:     "AnalyticsConsent non-nil -> true",
			manager:  &mock.ConfigManager{Cfg: &config.Config{AnalyticsConsent: &enabled}},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			consentChecker := NewConsentChecker(tt.manager, &insights.InsightsMock{}, &auth.AuthCheckerMock{})
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
	cm := &mock.ConfigManager{Cfg: &config.Config{AnalyticsConsent: nil}}
	acc := &AnalyticsConsentChecker{cm: cm}
	assert.False(t, cm.Saved)
	assert.Nil(t, cm.Cfg.AnalyticsConsent)

	err := acc.setConsentTrue()

	assert.NoError(t, err)
	assert.True(t, cm.Saved)
	assert.NotNil(t, cm.Cfg.AnalyticsConsent)
	assert.True(t, *cm.Cfg.AnalyticsConsent)
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

	truePtr := true
	tests := []struct {
		name               string
		initialConfig      config.Config
		loadErr            error
		apiInsights        core.InsightsAPI
		apiErr             error
		authLoggedIn       bool
		expectedSaved      bool
		expectedConsentSet bool
		expectedConsentVal bool
		expectedTokensLen  int
		expectedAutoConnID int64
	}{
		{
			name:               "consent already completed",
			initialConfig:      config.Config{AnalyticsConsent: &truePtr},
			expectedSaved:      false,
			expectedConsentSet: true,
			expectedConsentVal: true,
			expectedTokensLen:  0,
			expectedAutoConnID: 0,
		},
		{
			name:               "standard country -> set consent",
			initialConfig:      config.Config{AnalyticsConsent: nil},
			apiInsights:        &insights.InsightsMock{InsightsResult: &core.Insights{CountryCode: "ca"}},
			expectedSaved:      true,
			expectedConsentSet: true,
			expectedConsentVal: true,
			expectedTokensLen:  0,
			expectedAutoConnID: 0,
		},
		{
			name: "GDPR country & logged in -> do light logout",
			initialConfig: config.Config{
				AnalyticsConsent: nil,
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
			initialConfig:      config.Config{AnalyticsConsent: nil},
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
				assert.Equal(t, *cm.Cfg.AnalyticsConsent, tt.expectedConsentVal)
			} else {
				assert.Nil(t, cm.Cfg.AnalyticsConsent)
			}
			assert.Equal(t, len(cm.Cfg.TokensData), tt.expectedTokensLen)
			assert.Equal(t, cm.Cfg.AutoConnectData.ID, tt.expectedAutoConnID)
		})
	}
}
