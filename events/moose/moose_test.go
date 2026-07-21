//go:build moose

package moose

import (
	"math"
	moose "moose/events"
	"reflect"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"gotest.tools/v3/assert"
)

func TestNewSubscriber_AllMooseFuncsSet(t *testing.T) {
	category.Set(t, category.Unit)
	// values are not important, `mooseFuncs` field needs to be set properly
	sub := NewSubscriber("", nil, nil, nil, config.BuildTarget{}, "", "", "")

	v := reflect.ValueOf(sub.mooseFuncs)
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		value := v.Field(i)
		assert.Equal(t, value.IsNil(), false, "mooseFunc %q is not set", field.Name)
	}
}

func TestIsPaymentValid(t *testing.T) {
	category.Set(t, category.Unit)
	for _, tt := range []struct {
		name    string
		payment core.Payment
		valid   bool
	}{
		{name: "empty payment", valid: false},
		{name: "invalid status", valid: false, payment: core.Payment{Status: "invalid"}},
		{name: "done status", valid: true, payment: core.Payment{Status: "done"}},
		{name: "error status", valid: false, payment: core.Payment{Status: "error"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(
				t,
				tt.valid,
				isPaymentValid(core.PaymentResponse{Payment: tt.payment}),
			)
		})
	}
}

func TestFindPayment(t *testing.T) {
	category.Set(t, category.Unit)
	now := time.Now()
	for _, tt := range []struct {
		name     string
		payments []core.PaymentResponse
		payment  core.Payment
		ok       bool
	}{
		{
			name: "empty list",
		},
		{
			name:     "not found in invalid list",
			ok:       false,
			payments: []core.PaymentResponse{{Payment: core.Payment{}}},
		},
		{
			name:     "1 out of 1 found",
			ok:       true,
			payments: []core.PaymentResponse{{Payment: core.Payment{Status: "done"}}},
			payment:  core.Payment{Status: "done"},
		},
		{
			name: "latest valid found",
			ok:   true,
			payments: []core.PaymentResponse{
				{Payment: core.Payment{Status: "done", CreatedAt: now}},
				{Payment: core.Payment{Status: "done", CreatedAt: now.Add(-time.Second)}},
			},
			payment: core.Payment{Status: "done", CreatedAt: now},
		},
		{
			name: "latest valid found in mixture of valids and invalids",
			ok:   true,
			payments: []core.PaymentResponse{
				{Payment: core.Payment{Status: "done", CreatedAt: now.Add(-3 * time.Second)}},
				{Payment: core.Payment{Status: "invalid", CreatedAt: now}},
				{Payment: core.Payment{Status: "done", CreatedAt: now.Add(-time.Second)}},
				{Payment: core.Payment{Status: "invalid", CreatedAt: now.Add(-2 * time.Second)}},
			},
			payment: core.Payment{Status: "done", CreatedAt: now.Add(-time.Second)},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			payment, ok := findPayment(tt.payments)
			assert.Equal(t, tt.payment, payment)
			assert.Equal(t, tt.ok, ok)
		})
	}
}

func TestFindOrder(t *testing.T) {
	category.Set(t, category.Unit)
	validOrder := core.Order{ID: 123, RemoteID: 321, Status: "done"}
	for _, tt := range []struct {
		name    string
		orders  []core.Order
		payment core.Payment
		order   core.Order
		ok      bool
	}{
		{
			name: "empty orders list",
		},
		{
			name:    "invalid merchant ID",
			payment: core.Payment{Payer: core.Payer{OrderID: 123}, Subscription: core.Subscription{MerchantID: math.MaxInt32}},
			orders:  []core.Order{validOrder},
		},
		{
			name:    "merchant ID 3",
			payment: core.Payment{Payer: core.Payer{OrderID: 123}, Subscription: core.Subscription{MerchantID: 3}},
			orders:  []core.Order{validOrder},
			order:   validOrder,
			ok:      true,
		},
		{
			name:    "merchant ID 25",
			payment: core.Payment{Payer: core.Payer{OrderID: 321}, Subscription: core.Subscription{MerchantID: 25}},
			orders:  []core.Order{validOrder},
			order:   validOrder,
			ok:      true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			order, ok := findOrder(tt.payment, tt.orders)
			assert.DeepEqual(t, tt.order, order)
			assert.Equal(t, tt.ok, ok)
		})
	}
}

func TestChangeConsentState(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name                string
		currentConsentState config.AnalyticsConsent
		newConsentState     config.AnalyticsConsent
		expectedLevel       moose.UserConsent
		consentErrCode      uint32
		shouldFail          bool
	}{
		{
			name:                "undefined to enabled success",
			currentConsentState: config.ConsentUndefined,
			newConsentState:     config.ConsentGranted,
			expectedLevel:       moose.UserConsentNonEssential,
		},
		{
			name:                "undefined to disabled success",
			currentConsentState: config.ConsentUndefined,
			newConsentState:     config.ConsentDenied,
			expectedLevel:       moose.UserConsentEssential,
		},
		{
			name:                "enabled to disabled success",
			currentConsentState: config.ConsentGranted,
			newConsentState:     config.ConsentDenied,
			expectedLevel:       moose.UserConsentEssential,
		},
		{
			name:                "disabled to enabled success",
			currentConsentState: config.ConsentDenied,
			newConsentState:     config.ConsentGranted,
			expectedLevel:       moose.UserConsentNonEssential,
		},
		{
			name:                "undefined to enabled failure to consent",
			currentConsentState: config.ConsentUndefined,
			newConsentState:     config.ConsentGranted,
			expectedLevel:       moose.UserConsentNonEssential,
			consentErrCode:      1,
		},
		{
			name:                "undefined to disabled failure to consent",
			currentConsentState: config.ConsentUndefined,
			newConsentState:     config.ConsentDenied,
			expectedLevel:       moose.UserConsentEssential,
			consentErrCode:      1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var capturedLevel moose.UserConsent
			var consentCalls int
			consentFunc := func(userConsent moose.UserConsent) uint32 {
				consentCalls++
				capturedLevel = userConsent
				if test.consentErrCode != 0 {
					return test.consentErrCode
				}
				return 0
			}

			setConsentToCtx := func(val moose.NordvpnappConsentLevel) uint32 {
				return 0
			}

			configManagerMock := mock.NewMockConfigManager()
			configManagerMock.Cfg.AnalyticsConsent = test.currentConsentState
			s := &Subscriber{
				config: configManagerMock,
				mooseFuncs: mooseFunctions{
					setAppConsentLevel:       consentFunc,
					setConsentUserPreference: setConsentToCtx,
				},
			}

			err := s.changeConsentState(test.newConsentState)

			if test.consentErrCode != 0 || test.shouldFail {
				assert.Assert(t, err != nil)
			} else {
				assert.NilError(t, err)
			}

			assert.Equal(t, 1, consentCalls, "setAppConsentLevel should be called exactly once")
			assert.Equal(t, test.expectedLevel, capturedLevel, "Unexpected consent level passed to setAppConsentLevel.")
		})
	}
}

func TestAnalyticsConsentLevel(t *testing.T) {
	category.Set(t, category.Unit)

	cases := []struct {
		name  string
		input config.AnalyticsConsent
		want  moose.UserConsent
	}{
		{"granted", config.ConsentGranted, moose.UserConsentNonEssential},
		{"denied", config.ConsentDenied, moose.UserConsentEssential},
		{"undefined", config.ConsentUndefined, moose.UserConsentRejectAll},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, toAnalyticsConsentLevel(c.input))
		})
	}
}

func TestAnalyticsConsentLevel_Granted(t *testing.T) {
	category.Set(t, category.Unit)
	assert.Equal(t, moose.UserConsentNonEssential, toAnalyticsConsentLevel(config.ConsentGranted))
}

func TestAnalyticsConsentLevel_Denied(t *testing.T) {
	category.Set(t, category.Unit)
	assert.Equal(t, moose.UserConsentEssential, toAnalyticsConsentLevel(config.ConsentDenied))
}

func TestAnalyticsConsentLevel_Undefined(t *testing.T) {
	category.Set(t, category.Unit)
	assert.Equal(t, moose.UserConsentRejectAll, toAnalyticsConsentLevel(config.ConsentUndefined))
}

func TestGetTokenRenewDate(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		cfg      *config.Config
		expected string
	}{
		{
			name:     "nil config returns empty string",
			cfg:      nil,
			expected: "",
		},
		{
			name:     "nil TokensData returns empty string",
			cfg:      &config.Config{TokensData: nil},
			expected: "",
		},
		{
			name: "missing user ID returns empty string",
			cfg: &config.Config{
				AutoConnectData: config.AutoConnectData{ID: 123},
				TokensData:      map[int64]config.TokenData{},
			},
			expected: "",
		},
		{
			name: "returns TokenRenewDate for current user",
			cfg: &config.Config{
				AutoConnectData: config.AutoConnectData{ID: 123},
				TokensData: map[int64]config.TokenData{
					123: {TokenRenewDate: "2024-01-15 10:30:00"},
				},
			},
			expected: "2024-01-15 10:30:00",
		},
		{
			name: "returns empty string when TokenRenewDate is empty",
			cfg: &config.Config{
				AutoConnectData: config.AutoConnectData{ID: 123},
				TokensData: map[int64]config.TokenData{
					123: {TokenRenewDate: ""},
				},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTokenRenewDate(tt.cfg)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHandleTokenRenewDateChange(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name              string
		prevConfig        *config.Config
		currConfig        *config.Config
		expectSetCalled   bool
		expectedTimestamp int32
		expectError       bool
	}{
		{
			name:            "no action when current date is empty",
			prevConfig:      nil,
			currConfig:      &config.Config{TokensData: map[int64]config.TokenData{}},
			expectSetCalled: false,
		},
		{
			name: "no change when dates are the same",
			prevConfig: &config.Config{
				AutoConnectData: config.AutoConnectData{ID: 123},
				TokensData: map[int64]config.TokenData{
					123: {TokenRenewDate: "2024-01-15 10:30:00"},
				},
			},
			currConfig: &config.Config{
				AutoConnectData: config.AutoConnectData{ID: 123},
				TokensData: map[int64]config.TokenData{
					123: {TokenRenewDate: "2024-01-15 10:30:00"},
				},
			},
			expectSetCalled: false,
		},
		{
			name: "calls set when date changes",
			prevConfig: &config.Config{
				AutoConnectData: config.AutoConnectData{ID: 123},
				TokensData: map[int64]config.TokenData{
					123: {TokenRenewDate: "2024-01-15 10:30:00"},
				},
			},
			currConfig: &config.Config{
				AutoConnectData: config.AutoConnectData{ID: 123},
				TokensData: map[int64]config.TokenData{
					123: {TokenRenewDate: "2024-01-16 11:00:00"},
				},
			},
			expectSetCalled:   true,
			expectedTimestamp: int32(time.Date(2024, 1, 16, 11, 0, 0, 0, time.UTC).Unix()),
		},
		{
			name:       "calls set when date is set for first time (prev nil)",
			prevConfig: nil,
			currConfig: &config.Config{
				AutoConnectData: config.AutoConnectData{ID: 123},
				TokensData: map[int64]config.TokenData{
					123: {TokenRenewDate: "2024-01-16 11:00:00"},
				},
			},
			expectSetCalled:   true,
			expectedTimestamp: int32(time.Date(2024, 1, 16, 11, 0, 0, 0, time.UTC).Unix()),
		},
		{
			name: "calls set when date is set for first time (prev empty)",
			prevConfig: &config.Config{
				AutoConnectData: config.AutoConnectData{ID: 123},
				TokensData: map[int64]config.TokenData{
					123: {TokenRenewDate: ""},
				},
			},
			currConfig: &config.Config{
				AutoConnectData: config.AutoConnectData{ID: 123},
				TokensData: map[int64]config.TokenData{
					123: {TokenRenewDate: "2024-01-16 11:00:00"},
				},
			},
			expectSetCalled:   true,
			expectedTimestamp: int32(time.Date(2024, 1, 16, 11, 0, 0, 0, time.UTC).Unix()),
		},
		{
			name:       "returns error for invalid date format",
			prevConfig: nil,
			currConfig: &config.Config{
				AutoConnectData: config.AutoConnectData{ID: 123},
				TokensData: map[int64]config.TokenData{
					123: {TokenRenewDate: "invalid-date"},
				},
			},
			expectSetCalled: false,
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setCalled := false
			var capturedTimestamp int32

			s := &Subscriber{
				mooseFuncs: mooseFunctions{
					setTokenRenewDateCurrentState: func(timestamp int32) uint32 {
						setCalled = true
						capturedTimestamp = timestamp
						return 0
					},
				},
			}

			err := s.handleTokenRenewDateChange(tt.prevConfig, tt.currConfig)

			if tt.expectError {
				assert.Assert(t, err != nil, "expected error but got nil")
			} else {
				assert.NilError(t, err)
			}

			assert.Equal(t, tt.expectSetCalled, setCalled, "set call expectation mismatch")

			if tt.expectSetCalled {
				assert.Equal(t, tt.expectedTimestamp, capturedTimestamp, "timestamp mismatch")
			}
		})
	}
}

func TestNotifyThreatProtectionLite_CallsUserPreferenceSetter(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name                      string
		enabled                   bool
		mooseUserPrefErrCode      uint32
		mooseSetCustomDNSMetaErr  uint32
		mooseSetCustomDNSValueErr uint32
		expectNotifyErr           bool
		expectPrefCalled          bool
		expectCustomDNSCalled     bool
	}{
		{
			name:                      "returns err if setting user preference fails (when connected so current-state update is skipped)",
			enabled:                   true,
			mooseUserPrefErrCode:      12,
			mooseSetCustomDNSMetaErr:  0,
			mooseSetCustomDNSValueErr: 0,
			expectNotifyErr:           true,
			expectPrefCalled:          true,
			expectCustomDNSCalled:     false,
		},
		{
			name:                  "returns nil on success (when connected so current-state update is skipped)",
			enabled:               false,
			mooseUserPrefErrCode:  0,
			expectNotifyErr:       false,
			expectPrefCalled:      true,
			expectCustomDNSCalled: false, // TP Lite disabled, so custom DNS is not touched
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefCalled := false
			customDNSMetaCalled := false
			customDNSValueCalled := false
			var gotPref bool

			s := &Subscriber{
				mooseFuncs: mooseFunctions{
					setTPLiteUserPreference: func(v bool) uint32 {
						prefCalled = true
						gotPref = v
						return tt.mooseUserPrefErrCode
					},
					setCustomDNSMeta: func(meta string) uint32 {
						customDNSMetaCalled = true
						return tt.mooseSetCustomDNSMetaErr
					},
					setCustomDNSValue: func(enabled bool) uint32 {
						customDNSValueCalled = true
						return tt.mooseSetCustomDNSValueErr
					},
				},
			}

			// make sure we don't hit the "Current State" call path:
			// `NotifyThreatProtectionLite` only sets Current State when connectionStartTime.IsZero().
			s.connectionStartTime = time.Now()

			err := s.NotifyThreatProtectionLite(tt.enabled)

			assert.Equal(t, tt.expectPrefCalled, prefCalled)
			if tt.expectPrefCalled {
				assert.Equal(t, tt.enabled, gotPref)
			}

			assert.Equal(t, tt.expectCustomDNSCalled, customDNSMetaCalled)
			assert.Equal(t, tt.expectCustomDNSCalled, customDNSValueCalled)

			if tt.expectNotifyErr {
				assert.Assert(t, err != nil)
			} else {
				assert.NilError(t, err)
			}
		})
	}
}

func TestSetCustomDNS(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name                 string
		dnsIPs               []string
		mooseMetaErrCode     uint32
		mooseValueErrCode    uint32
		expectErr            bool
		expectMetaCalled     bool
		expectValueCalled    bool
		expectedMetaContent  string
		expectedValueContent bool
	}{
		{
			name:                 "no DNS IPs - success",
			dnsIPs:               []string{},
			mooseMetaErrCode:     0,
			mooseValueErrCode:    0,
			expectErr:            false,
			expectMetaCalled:     true,
			expectValueCalled:    true,
			expectedMetaContent:  `{"count":0}`,
			expectedValueContent: false,
		},
		{
			name:                 "single DNS IP - success",
			dnsIPs:               []string{"1.1.1.1"},
			mooseMetaErrCode:     0,
			mooseValueErrCode:    0,
			expectErr:            false,
			expectMetaCalled:     true,
			expectValueCalled:    true,
			expectedMetaContent:  `{"count":1}`,
			expectedValueContent: true,
		},
		{
			name:                 "multiple DNS IPs - success",
			dnsIPs:               []string{"1.1.1.1", "8.8.8.8", "8.8.4.4"},
			mooseMetaErrCode:     0,
			mooseValueErrCode:    0,
			expectErr:            false,
			expectMetaCalled:     true,
			expectValueCalled:    true,
			expectedMetaContent:  `{"count":3}`,
			expectedValueContent: true,
		},
		{
			name:              "meta setter fails - propagates error",
			dnsIPs:            []string{"1.1.1.1"},
			mooseMetaErrCode:  1,
			mooseValueErrCode: 0,
			expectErr:         true,
			expectMetaCalled:  true,
			expectValueCalled: false, // value setter not called if meta setter fails
		},
		{
			name:                "value setter fails - propagates error",
			dnsIPs:              []string{"1.1.1.1"},
			mooseMetaErrCode:    0,
			mooseValueErrCode:   1,
			expectErr:           true,
			expectMetaCalled:    true,
			expectValueCalled:   true,
			expectedMetaContent: `{"count":1}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metaCalled := false
			valueCalled := false
			var gotMeta string
			var gotValue bool

			s := &Subscriber{
				mooseFuncs: mooseFunctions{
					setCustomDNSMeta: func(meta string) uint32 {
						metaCalled = true
						gotMeta = meta
						return tt.mooseMetaErrCode
					},
					setCustomDNSValue: func(enabled bool) uint32 {
						valueCalled = true
						gotValue = enabled
						return tt.mooseValueErrCode
					},
				},
			}

			data := events.DataDNS{Ips: tt.dnsIPs}
			err := s.setCustomDNS(data)

			assert.Equal(t, tt.expectMetaCalled, metaCalled)
			if tt.expectMetaCalled && tt.mooseMetaErrCode == 0 {
				assert.Equal(t, tt.expectedMetaContent, gotMeta)
			}

			assert.Equal(t, tt.expectValueCalled, valueCalled)
			if tt.expectValueCalled && tt.mooseValueErrCode == 0 {
				assert.Equal(t, tt.expectedValueContent, gotValue)
			}

			if tt.expectErr {
				assert.Assert(t, err != nil)
			} else {
				assert.NilError(t, err)
			}
		})
	}
}

func TestNotifyDNS(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name                   string
		dnsIPs                 []string
		mooseMetaErrCode       uint32
		mooseValueErrCode      uint32
		mooseTPLiteUserPrefErr uint32
		mooseTPLiteCurrentErr  uint32
		expectErr              bool
		expectMetaCalled       bool
		expectValueCalled      bool
		expectTPLiteCalled     bool
		expectedTPLiteValue    bool
	}{
		{
			name:               "no DNS IPs - custom DNS disabled, TP Lite not touched",
			dnsIPs:             []string{},
			mooseMetaErrCode:   0,
			mooseValueErrCode:  0,
			expectErr:          false,
			expectMetaCalled:   true,
			expectValueCalled:  true,
			expectTPLiteCalled: false, // TP Lite only touched when custom DNS is enabled
		},
		{
			name:                   "single DNS IP - custom DNS enabled, TP Lite disabled (not connected)",
			dnsIPs:                 []string{"1.1.1.1"},
			mooseMetaErrCode:       0,
			mooseValueErrCode:      0,
			mooseTPLiteUserPrefErr: 0,
			mooseTPLiteCurrentErr:  0,
			expectErr:              false,
			expectMetaCalled:       true,
			expectValueCalled:      true,
			expectTPLiteCalled:     true,
			expectedTPLiteValue:    false,
		},
		{
			name:                   "multiple DNS IPs - custom DNS enabled, TP Lite disabled (not connected)",
			dnsIPs:                 []string{"1.1.1.1", "8.8.8.8"},
			mooseMetaErrCode:       0,
			mooseValueErrCode:      0,
			mooseTPLiteUserPrefErr: 0,
			mooseTPLiteCurrentErr:  0,
			expectErr:              false,
			expectMetaCalled:       true,
			expectValueCalled:      true,
			expectTPLiteCalled:     true,
			expectedTPLiteValue:    false,
		},
		{
			name:               "custom DNS meta setter fails - propagates error, TP Lite not touched",
			dnsIPs:             []string{"1.1.1.1"},
			mooseMetaErrCode:   1,
			mooseValueErrCode:  0,
			expectErr:          true,
			expectMetaCalled:   true,
			expectValueCalled:  false,
			expectTPLiteCalled: false, // setCustomDNS fails early
		},
		{
			name:               "custom DNS value setter fails - propagates error, TP Lite not touched",
			dnsIPs:             []string{"1.1.1.1"},
			mooseMetaErrCode:   0,
			mooseValueErrCode:  1,
			expectErr:          true,
			expectMetaCalled:   true,
			expectValueCalled:  true,
			expectTPLiteCalled: false, // setCustomDNS fails, so setTPLite not called
		},
		{
			name:                   "TP Lite user pref setter fails - propagates error (not connected)",
			dnsIPs:                 []string{"1.1.1.1"},
			mooseMetaErrCode:       0,
			mooseValueErrCode:      0,
			mooseTPLiteUserPrefErr: 12,
			mooseTPLiteCurrentErr:  0,
			expectErr:              true,
			expectMetaCalled:       true,
			expectValueCalled:      true,
			expectTPLiteCalled:     true,
		},
		{
			name:                   "TP Lite current setter fails - propagates error (not connected)",
			dnsIPs:                 []string{"1.1.1.1"},
			mooseMetaErrCode:       0,
			mooseValueErrCode:      0,
			mooseTPLiteUserPrefErr: 0,
			mooseTPLiteCurrentErr:  1,
			expectErr:              true,
			expectMetaCalled:       true,
			expectValueCalled:      true,
			expectTPLiteCalled:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metaCalled := false
			valueCalled := false
			tpLiteUserPrefCalled := false
			tpLiteCurrentCalled := false
			var gotTPLiteValue bool

			s := &Subscriber{
				mooseFuncs: mooseFunctions{
					setCustomDNSMeta: func(meta string) uint32 {
						metaCalled = true
						return tt.mooseMetaErrCode
					},
					setCustomDNSValue: func(enabled bool) uint32 {
						valueCalled = true
						return tt.mooseValueErrCode
					},
					setTPLiteUserPreference: func(v bool) uint32 {
						tpLiteUserPrefCalled = true
						return tt.mooseTPLiteUserPrefErr
					},
					setTPLiteCurrentState: func(v bool) uint32 {
						tpLiteCurrentCalled = true
						gotTPLiteValue = v
						return tt.mooseTPLiteCurrentErr
					},
				},
			}

			// Ensure we're in "not connected" state so TP Lite current state is set
			s.connectionStartTime = time.Time{}

			data := events.DataDNS{Ips: tt.dnsIPs}
			err := s.NotifyDNS(data)

			assert.Equal(t, tt.expectMetaCalled, metaCalled)
			assert.Equal(t, tt.expectValueCalled, valueCalled)
			assert.Equal(t, tt.expectTPLiteCalled, tpLiteUserPrefCalled)
			assert.Equal(t, tt.expectTPLiteCalled, tpLiteCurrentCalled)

			if tt.expectTPLiteCalled && tt.mooseTPLiteCurrentErr == 0 {
				assert.Equal(t, tt.expectedTPLiteValue, gotTPLiteValue)
			}

			if tt.expectErr {
				assert.Assert(t, err != nil)
			} else {
				assert.NilError(t, err)
			}
		})
	}
}

func TestHasSensitiveServerGroup(t *testing.T) {
	category.Set(t, category.Unit)

	cases := []struct {
		name   string
		groups []config.ServerGroup
		want   bool
	}{
		{
			name: "contains_dedicated_ip",
			groups: []config.ServerGroup{
				config.ServerGroup_DEDICATED_IP,
				config.ServerGroup_STANDARD_VPN_SERVERS,
			},
			want: true,
		},
		{
			name:   "only_dedicated_ip",
			groups: []config.ServerGroup{config.ServerGroup_DEDICATED_IP},
			want:   true,
		},
		{
			name: "no_sensitive_groups",
			groups: []config.ServerGroup{
				config.ServerGroup_STANDARD_VPN_SERVERS,
				config.ServerGroup_P2P,
			},
			want: false,
		},
		{
			name:   "empty_slice",
			groups: []config.ServerGroup{},
			want:   false,
		},
		{
			name:   "nil_slice",
			groups: nil,
			want:   false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, hasSensitiveServerGroup(c.groups))
		})
	}
}

func TestSetDedicatedServerServiceStatus(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name              string
		services          core.ServicesResponse
		DSEnabled         bool
		mooseErrCode      uint32
		expectErr         bool
		expectCalled      bool
		expectedActive    bool
		expectedDSEnabled bool
	}{
		{
			name: "service ID 33 present - sets true",
			services: core.ServicesResponse{
				{Service: core.Service{ID: auth.DedicatedServersServiceID}, ExpiresAt: "2050-06-04 00:00:00"},
			},
			mooseErrCode:   0,
			expectCalled:   true,
			expectedActive: true,
		},
		{
			name: "service ID 33 among others - sets true",
			services: core.ServicesResponse{
				{Service: core.Service{ID: 1}, ExpiresAt: "2050-06-04 00:00:00"},
				{Service: core.Service{ID: auth.DedicatedServersServiceID}, ExpiresAt: "2050-06-04 00:00:00"},
				{Service: core.Service{ID: 11}, ExpiresAt: "2050-06-04 00:00:00"},
			},
			mooseErrCode:   0,
			expectCalled:   true,
			expectedActive: true,
		},
		{
			name: "service ID 33 expired - sets false",
			services: core.ServicesResponse{
				{Service: core.Service{ID: auth.DedicatedServersServiceID}, ExpiresAt: "2023-08-22 00:00:00"},
			},
			mooseErrCode:   0,
			expectCalled:   true,
			expectedActive: false,
		},
		{
			name: "no service ID 33 - sets false",
			services: core.ServicesResponse{
				{Service: core.Service{ID: 1}, ExpiresAt: "2050-06-04 00:00:00"},
				{Service: core.Service{ID: 11}, ExpiresAt: "2050-06-04 00:00:00"},
			},
			mooseErrCode:   0,
			expectCalled:   true,
			expectedActive: false,
		},
		{
			name:           "empty services - sets false",
			services:       core.ServicesResponse{},
			mooseErrCode:   0,
			expectCalled:   true,
			expectedActive: false,
		},
		{
			name:           "nil services - sets false",
			services:       nil,
			mooseErrCode:   0,
			expectCalled:   true,
			expectedActive: false,
		},
		{
			name: "moose setter fails - propagates error",
			services: core.ServicesResponse{
				{Service: core.Service{ID: auth.DedicatedServersServiceID}, ExpiresAt: "2050-06-04 00:00:00"},
			},
			mooseErrCode:      7,
			expectErr:         true,
			expectCalled:      true,
			DSEnabled:         true,
			expectedDSEnabled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			var gotValue bool
			hasDS := false

			s := &Subscriber{
				mooseFuncs: mooseFunctions{
					setDSIsActive: func(active bool) uint32 {
						called = true
						gotValue = active
						return tt.mooseErrCode
					},
					setDSEnabled: func(bool) uint32 {
						hasDS = tt.DSEnabled
						return 0
					},
				},
			}

			hasDSService := hasDedicatedServerService(tt.services)
			err := s.setDedicatedServerServiceStatus(hasDSService, hasDS)

			assert.Equal(t, tt.expectCalled, called)
			if called && tt.mooseErrCode == 0 {
				assert.Equal(t, tt.expectedActive, gotValue)
			}
			if tt.expectErr {
				assert.Assert(t, err != nil)
			} else {
				assert.NilError(t, err)
			}
		})
	}
}

func TestNotifyConnect_DedicatedIP_StripsBodyAndUnsetsContext(t *testing.T) {
	category.Set(t, category.Unit)
	sub := NewSubscriber("", nil, nil, nil, config.BuildTarget{}, "", "", "")

	var capturedDomain, capturedUUID string
	var calls []string
	sub.mooseFuncs.sendConnect = func(
		_ moose.EventParams,
		_ moose.TargetConnectionParams,
		additional moose.TargetConnectionAdditionalParams,
		_ moose.ConnectionParams,
		_ moose.NordvpnappOptBool,
		_ int32,
		uuid string,
		_ *string,
	) uint32 {
		capturedDomain = additional.TargetServerDomain
		capturedUUID = uuid
		calls = append(calls, "sendConnect")
		return 0
	}
	sub.mooseFuncs.unsetServerDomainValue = func() uint32 {
		calls = append(calls, "unsetServerDomainValue")
		return 0
	}
	sub.mooseFuncs.unsetRecommendationUuid = func() uint32 {
		calls = append(calls, "unsetRecommendationUuid")
		return 0
	}
	sub.mooseFuncs.setServerGroupValue = func(_ moose.NordvpnappServerGroup) uint32 {
		calls = append(calls, "setServerGroupValue")
		return 0
	}

	err := sub.NotifyConnect(events.DataConnect{
		ServerGroups: []config.ServerGroup{
			config.ServerGroup_DEDICATED_IP,
			config.ServerGroup_STANDARD_VPN_SERVERS,
		},
		TargetServerDomain: "dip-1234.nordvpn.com",
		RecommendationUUID: "rec-abc",
		EventStatus:        events.StatusAttempt,
	})

	assert.NilError(t, err)
	assert.Equal(t, "", capturedDomain)
	assert.Equal(t, "", capturedUUID)
	assert.DeepEqual(t, []string{"unsetServerDomainValue", "unsetRecommendationUuid", "sendConnect"}, calls)
}

func TestNotifyConnect_DedicatedIPByHostname_StripsBodyAndUnsetsContext(t *testing.T) {
	category.Set(t, category.Unit)
	sub := NewSubscriber("", nil, nil, nil, config.BuildTarget{}, "", "", "")

	var capturedDomain, capturedUUID string
	var calls []string
	sub.mooseFuncs.sendConnect = func(
		_ moose.EventParams,
		_ moose.TargetConnectionParams,
		additional moose.TargetConnectionAdditionalParams,
		_ moose.ConnectionParams,
		_ moose.NordvpnappOptBool,
		_ int32,
		uuid string,
		_ *string,
	) uint32 {
		capturedDomain = additional.TargetServerDomain
		capturedUUID = uuid
		calls = append(calls, "sendConnect")
		return 0
	}
	sub.mooseFuncs.unsetServerDomainValue = func() uint32 {
		calls = append(calls, "unsetServerDomainValue")
		return 0
	}
	sub.mooseFuncs.unsetRecommendationUuid = func() uint32 {
		calls = append(calls, "unsetRecommendationUuid")
		return 0
	}
	sub.mooseFuncs.setServerGroupValue = func(_ moose.NordvpnappServerGroup) uint32 {
		calls = append(calls, "setServerGroupValue")
		return 0
	}

	err := sub.NotifyConnect(events.DataConnect{
		ServerGroups: []config.ServerGroup{
			config.ServerGroup_DEDICATED_IP,
			config.ServerGroup_STANDARD_VPN_SERVERS,
		},
		TargetServerDomain: "dip-1234.nordvpn.com",
		RecommendationUUID: "rec-abc",
		EventStatus:        events.StatusAttempt,
	})

	assert.NilError(t, err)
	assert.Equal(t, "", capturedDomain)
	assert.Equal(t, "", capturedUUID)
	assert.DeepEqual(t, []string{"unsetServerDomainValue", "unsetRecommendationUuid", "sendConnect"}, calls)
}

func TestNotifyConnect_Success_DedicatedIPByHostname_ContextValueIsUserTarget(t *testing.T) {
	category.Set(t, category.Unit)
	sub := NewSubscriber("", nil, nil, nil, config.BuildTarget{}, "", "", "")
	noopDisconnectAmbientMooseFuncs(sub)

	var capturedGroup moose.NordvpnappServerGroup
	var groupCalls int
	sub.mooseFuncs.sendConnect = func(
		_ moose.EventParams,
		_ moose.TargetConnectionParams,
		_ moose.TargetConnectionAdditionalParams,
		_ moose.ConnectionParams,
		_ moose.NordvpnappOptBool,
		_ int32,
		_ string,
		_ *string,
	) uint32 {
		return 0
	}
	sub.mooseFuncs.setTPLiteCurrentState = func(_ bool) uint32 { return 0 }
	sub.mooseFuncs.unsetServerDomainValue = func() uint32 { return 0 }
	sub.mooseFuncs.unsetRecommendationUuid = func() uint32 { return 0 }
	sub.mooseFuncs.setServerGroupValue = func(group moose.NordvpnappServerGroup) uint32 {
		capturedGroup = group
		groupCalls++
		return 0
	}

	err := sub.NotifyConnect(events.DataConnect{
		TargetServerGroupID: config.ServerGroup_STANDARD_VPN_SERVERS,
		ServerGroups: []config.ServerGroup{
			config.ServerGroup_DEDICATED_IP,
			config.ServerGroup_STANDARD_VPN_SERVERS,
		},
		TargetServerDomain: "dip-9999.nordvpn.com",
		RecommendationUUID: "rec-dip-uuid",
		EventStatus:        events.StatusSuccess,
	})

	assert.NilError(t, err)
	assert.Equal(t, moose.NordvpnappServerGroupStandard, capturedGroup)
	assert.Equal(t, 1, groupCalls)
}

func TestNotifyConnect_Success_EmptyTargetServerGroup_UnsetsServerGroupContext(t *testing.T) {
	category.Set(t, category.Unit)
	sub := NewSubscriber("", nil, nil, nil, config.BuildTarget{}, "", "", "")
	noopDisconnectAmbientMooseFuncs(sub)

	var setCalls, unsetCalls int
	sub.mooseFuncs.sendConnect = func(
		_ moose.EventParams,
		_ moose.TargetConnectionParams,
		_ moose.TargetConnectionAdditionalParams,
		_ moose.ConnectionParams,
		_ moose.NordvpnappOptBool,
		_ int32,
		_ string,
		_ *string,
	) uint32 {
		return 0
	}
	sub.mooseFuncs.setTPLiteCurrentState = func(_ bool) uint32 { return 0 }
	sub.mooseFuncs.setServerGroupValue = func(_ moose.NordvpnappServerGroup) uint32 {
		setCalls++
		return 0
	}
	sub.mooseFuncs.unsetServerGroupValue = func() uint32 {
		unsetCalls++
		return 0
	}

	err := sub.NotifyConnect(events.DataConnect{
		EventStatus: events.StatusSuccess,
	})

	assert.NilError(t, err)
	assert.Equal(t, 0, setCalls)
	assert.Equal(t, 1, unsetCalls)
}

func TestNotifyConnect_StandardVPN_PreservesBodyAndSkipsUnsets(t *testing.T) {
	category.Set(t, category.Unit)
	sub := NewSubscriber("", nil, nil, nil, config.BuildTarget{}, "", "", "")

	var capturedDomain, capturedUUID string
	var domainUnsets, uuidUnsets, groupSetCalls int
	sub.mooseFuncs.sendConnect = func(
		_ moose.EventParams,
		_ moose.TargetConnectionParams,
		additional moose.TargetConnectionAdditionalParams,
		_ moose.ConnectionParams,
		_ moose.NordvpnappOptBool,
		_ int32,
		uuid string,
		_ *string,
	) uint32 {
		capturedDomain = additional.TargetServerDomain
		capturedUUID = uuid
		return 0
	}
	sub.mooseFuncs.unsetServerDomainValue = func() uint32 {
		domainUnsets++
		return 0
	}
	sub.mooseFuncs.unsetRecommendationUuid = func() uint32 {
		uuidUnsets++
		return 0
	}
	sub.mooseFuncs.setServerGroupValue = func(_ moose.NordvpnappServerGroup) uint32 {
		groupSetCalls++
		return 0
	}

	err := sub.NotifyConnect(events.DataConnect{
		TargetServerDomain: "us-1.nordvpn.com",
		RecommendationUUID: "rec-def",
		EventStatus:        events.StatusAttempt,
	})

	assert.NilError(t, err)
	assert.Equal(t, "us-1.nordvpn.com", capturedDomain)
	assert.Equal(t, "rec-def", capturedUUID)
	assert.Equal(t, 0, domainUnsets)
	assert.Equal(t, 0, uuidUnsets)
	assert.Equal(t, 0, groupSetCalls)
}

func TestNotifyConnect_DedicatedIP_UnsetContinuesAfterFirstError(t *testing.T) {
	category.Set(t, category.Unit)
	sub := NewSubscriber("", nil, nil, nil, config.BuildTarget{}, "", "", "")

	sub.mooseFuncs.sendConnect = func(
		_ moose.EventParams,
		_ moose.TargetConnectionParams,
		_ moose.TargetConnectionAdditionalParams,
		_ moose.ConnectionParams,
		_ moose.NordvpnappOptBool,
		_ int32,
		_ string,
		_ *string,
	) uint32 {
		return 0
	}
	sub.mooseFuncs.unsetServerDomainValue = func() uint32 {
		return 7 // moose: context set error — logged, must not block second unset
	}
	var uuidUnsets int
	sub.mooseFuncs.unsetRecommendationUuid = func() uint32 {
		uuidUnsets++
		return 0
	}
	sub.mooseFuncs.setServerGroupValue = func(_ moose.NordvpnappServerGroup) uint32 { return 0 }

	err := sub.NotifyConnect(events.DataConnect{
		ServerGroups: []config.ServerGroup{config.ServerGroup_DEDICATED_IP},
		EventStatus:  events.StatusAttempt,
	})

	assert.NilError(t, err)
	assert.Equal(t, 1, uuidUnsets)
}

func TestNotifyConnect_MeshnetPeerWithSensitiveGroup_DoesNotSetFlag(t *testing.T) {
	category.Set(t, category.Unit)
	sub := NewSubscriber("", nil, nil, nil, config.BuildTarget{}, "", "", "")

	_ = sub.NotifyConnect(events.DataConnect{
		IsMeshnetPeer: true,
		ServerGroups:  []config.ServerGroup{config.ServerGroup_DEDICATED_IP},
		EventStatus:   events.StatusAttempt,
	})

	assert.Equal(t, false, sub.connectionToSensitiveServerGroup)
}

func TestNotifyConnect_Success_InvokesPostConnectContextSetters(t *testing.T) {
	category.Set(t, category.Unit)
	sub := NewSubscriber("", nil, nil, nil, config.BuildTarget{}, "", "", "")

	sub.mooseFuncs.sendConnect = func(
		_ moose.EventParams,
		_ moose.TargetConnectionParams,
		_ moose.TargetConnectionAdditionalParams,
		_ moose.ConnectionParams,
		_ moose.NordvpnappOptBool,
		_ int32,
		_ string,
		_ *string,
	) uint32 {
		return 0
	}

	var tpLiteCalls, isOnVpnCalls, countryCalls, groupCalls int
	var capturedTPLite, capturedIsOnVpn bool
	var capturedCountry string
	var capturedGroup moose.NordvpnappServerGroup
	sub.mooseFuncs.setTPLiteCurrentState = func(enabled bool) uint32 {
		capturedTPLite = enabled
		tpLiteCalls++
		return 0
	}
	sub.mooseFuncs.setIsOnVpnValue = func(onVpn bool) uint32 {
		capturedIsOnVpn = onVpn
		isOnVpnCalls++
		return 0
	}
	sub.mooseFuncs.setServerCountryValue = func(country string) uint32 {
		capturedCountry = country
		countryCalls++
		return 0
	}
	sub.mooseFuncs.setServerGroupValue = func(group moose.NordvpnappServerGroup) uint32 {
		capturedGroup = group
		groupCalls++
		return 0
	}

	err := sub.NotifyConnect(events.DataConnect{
		TargetServerGroupID:     config.ServerGroup_STANDARD_VPN_SERVERS,
		TargetServerCountryCode: "us",
		ThreatProtectionLite:    false,
		EventStatus:             events.StatusSuccess,
	})

	assert.NilError(t, err)
	assert.Equal(t, 1, tpLiteCalls)
	assert.Equal(t, 1, isOnVpnCalls)
	assert.Equal(t, 1, countryCalls)
	assert.Equal(t, 1, groupCalls)
	assert.Equal(t, false, capturedTPLite)
	assert.Equal(t, true, capturedIsOnVpn)
	assert.Equal(t, "us", capturedCountry)
	assert.Equal(t, moose.NordvpnappServerGroupStandard, capturedGroup)
}

func TestNotifyConnect_MeshnetPeer_PreservesPriorSensitiveFlag(t *testing.T) {
	category.Set(t, category.Unit)
	sub := NewSubscriber("", nil, nil, nil, config.BuildTarget{}, "", "", "")

	// simulate a prior server connect to a sensitive group that set the flag.
	sub.connectionToSensitiveServerGroup = true

	_ = sub.NotifyConnect(events.DataConnect{
		IsMeshnetPeer: true,
		EventStatus:   events.StatusAttempt,
	})

	assert.Equal(t, true, sub.connectionToSensitiveServerGroup)
}

func noopDisconnectAmbientMooseFuncs(sub *Subscriber) {
	sub.mooseFuncs.unsetTPLiteCurrentState = func() uint32 { return 0 }
	sub.mooseFuncs.setServerCountryValue = func(_ string) uint32 { return 0 }
	sub.mooseFuncs.setServerGroupValue = func(_ moose.NordvpnappServerGroup) uint32 { return 0 }
	sub.mooseFuncs.unsetServerGroupValue = func() uint32 { return 0 }
	sub.mooseFuncs.setIsOnVpnValue = func(_ bool) uint32 { return 0 }
}

func TestNotifyDisconnect_AfterSensitiveConnect_SkipsRecommendationUuidContext(t *testing.T) {
	category.Set(t, category.Unit)
	sub := NewSubscriber("", nil, nil, nil, config.BuildTarget{}, "", "", "")
	noopDisconnectAmbientMooseFuncs(sub)

	// simulate the state left by a prior connect to a sensitive group.
	sub.connectionToSensitiveServerGroup = true

	var sendDisconnectCalls, setCalls, unsetCalls int
	sub.mooseFuncs.sendDisconnect = func(
		_ moose.EventParams,
		_ moose.TargetConnectionParams,
		_ moose.ConnectionParams,
		_ int32,
		_ int32,
		_ *string,
	) uint32 {
		sendDisconnectCalls++
		return 0
	}
	sub.mooseFuncs.setRecommendationUuid = func(_ string) uint32 {
		setCalls++
		return 0
	}
	sub.mooseFuncs.unsetRecommendationUuid = func() uint32 {
		unsetCalls++
		return 0
	}

	err := sub.NotifyDisconnect(events.DataDisconnect{
		RecommendationUUID: "rec-dip-uuid",
		EventStatus:        events.StatusSuccess,
	})

	assert.NilError(t, err)
	assert.Equal(t, 1, sendDisconnectCalls)
	assert.Equal(t, 0, setCalls)
	assert.Equal(t, 0, unsetCalls)
	assert.Equal(t, false, sub.connectionToSensitiveServerGroup)
}

func TestNotifyDisconnect_AfterNonSensitiveConnect_SetsAndUnsetsRecommendationUuidContext(t *testing.T) {
	category.Set(t, category.Unit)
	sub := NewSubscriber("", nil, nil, nil, config.BuildTarget{}, "", "", "")
	noopDisconnectAmbientMooseFuncs(sub)

	var capturedUUID string
	var sendDisconnectCalls, setCalls, unsetCalls int
	sub.mooseFuncs.sendDisconnect = func(
		_ moose.EventParams,
		_ moose.TargetConnectionParams,
		_ moose.ConnectionParams,
		_ int32,
		_ int32,
		_ *string,
	) uint32 {
		sendDisconnectCalls++
		return 0
	}
	sub.mooseFuncs.setRecommendationUuid = func(uuid string) uint32 {
		capturedUUID = uuid
		setCalls++
		return 0
	}
	sub.mooseFuncs.unsetRecommendationUuid = func() uint32 {
		unsetCalls++
		return 0
	}

	err := sub.NotifyDisconnect(events.DataDisconnect{
		RecommendationUUID: "rec-standard-uuid",
		EventStatus:        events.StatusSuccess,
	})

	assert.NilError(t, err)
	assert.Equal(t, 1, sendDisconnectCalls)
	assert.Equal(t, 1, setCalls)
	assert.Equal(t, "rec-standard-uuid", capturedUUID)
	assert.Equal(t, 1, unsetCalls)
	assert.Equal(t, false, sub.connectionToSensitiveServerGroup)
}

func TestNotifyDisconnect_EmptyRecommendationUuid_SkipsBothCalls(t *testing.T) {
	category.Set(t, category.Unit)
	sub := NewSubscriber("", nil, nil, nil, config.BuildTarget{}, "", "", "")
	noopDisconnectAmbientMooseFuncs(sub)

	var sendDisconnectCalls, setCalls, unsetCalls int
	sub.mooseFuncs.sendDisconnect = func(
		_ moose.EventParams,
		_ moose.TargetConnectionParams,
		_ moose.ConnectionParams,
		_ int32,
		_ int32,
		_ *string,
	) uint32 {
		sendDisconnectCalls++
		return 0
	}
	sub.mooseFuncs.setRecommendationUuid = func(_ string) uint32 {
		setCalls++
		return 0
	}
	sub.mooseFuncs.unsetRecommendationUuid = func() uint32 {
		unsetCalls++
		return 0
	}

	err := sub.NotifyDisconnect(events.DataDisconnect{
		RecommendationUUID: "",
		EventStatus:        events.StatusSuccess,
	})

	assert.NilError(t, err)
	assert.Equal(t, 1, sendDisconnectCalls)
	assert.Equal(t, 0, setCalls)
	assert.Equal(t, 0, unsetCalls)
	assert.Equal(t, false, sub.connectionToSensitiveServerGroup)
}

func TestNotifyConnect_Success_DedicatedIP_SetsGroupAndKeepsDependentsEmpty(t *testing.T) {
	category.Set(t, category.Unit)
	sub := NewSubscriber("", nil, nil, nil, config.BuildTarget{}, "", "", "")
	noopDisconnectAmbientMooseFuncs(sub)

	var capturedGroup moose.NordvpnappServerGroup
	var groupCalls, domainUnsets, uuidUnsets int
	var groupCallOrder, sendConnectCallOrder int
	var nextCall int
	sub.mooseFuncs.sendConnect = func(
		_ moose.EventParams,
		_ moose.TargetConnectionParams,
		_ moose.TargetConnectionAdditionalParams,
		_ moose.ConnectionParams,
		_ moose.NordvpnappOptBool,
		_ int32,
		_ string,
		_ *string,
	) uint32 {
		nextCall++
		sendConnectCallOrder = nextCall
		return 0
	}
	sub.mooseFuncs.setTPLiteCurrentState = func(_ bool) uint32 { return 0 }
	sub.mooseFuncs.setServerGroupValue = func(group moose.NordvpnappServerGroup) uint32 {
		nextCall++
		groupCallOrder = nextCall
		capturedGroup = group
		groupCalls++
		return 0
	}
	sub.mooseFuncs.unsetServerDomainValue = func() uint32 {
		domainUnsets++
		return 0
	}
	sub.mooseFuncs.unsetRecommendationUuid = func() uint32 {
		uuidUnsets++
		return 0
	}

	err := sub.NotifyConnect(events.DataConnect{
		TargetServerGroupID: config.ServerGroup_DEDICATED_IP,
		ServerGroups: []config.ServerGroup{
			config.ServerGroup_DEDICATED_IP,
			config.ServerGroup_STANDARD_VPN_SERVERS,
		},
		TargetServerDomain: "dip-9999.nordvpn.com",
		RecommendationUUID: "rec-dip-uuid",
		EventStatus:        events.StatusSuccess,
	})

	assert.NilError(t, err)
	assert.Equal(t, moose.NordvpnappServerGroupDedicatedIp, capturedGroup)
	assert.Equal(t, 1, groupCalls)
	assert.Equal(t, true, domainUnsets >= 1)
	assert.Equal(t, true, uuidUnsets >= 1)
	assert.Equal(t, true, groupCallOrder < sendConnectCallOrder)
}

func TestNotifyConnect_Failure_DoesNotWriteServerGroupContext(t *testing.T) {
	category.Set(t, category.Unit)
	sub := NewSubscriber("", nil, nil, nil, config.BuildTarget{}, "", "", "")

	var groupSetCalls, domainUnsets, uuidUnsets int
	sub.mooseFuncs.sendConnect = func(
		_ moose.EventParams,
		_ moose.TargetConnectionParams,
		_ moose.TargetConnectionAdditionalParams,
		_ moose.ConnectionParams,
		_ moose.NordvpnappOptBool,
		_ int32,
		_ string,
		_ *string,
	) uint32 {
		return 0
	}
	sub.mooseFuncs.setServerGroupValue = func(_ moose.NordvpnappServerGroup) uint32 {
		groupSetCalls++
		return 0
	}
	sub.mooseFuncs.unsetServerDomainValue = func() uint32 {
		domainUnsets++
		return 0
	}
	sub.mooseFuncs.unsetRecommendationUuid = func() uint32 {
		uuidUnsets++
		return 0
	}

	err := sub.NotifyConnect(events.DataConnect{
		ServerGroups: []config.ServerGroup{config.ServerGroup_DEDICATED_IP},
		EventStatus:  events.StatusFailure,
	})

	assert.NilError(t, err)
	assert.Equal(t, 0, groupSetCalls)
	assert.Equal(t, 1, domainUnsets)
	assert.Equal(t, 1, uuidUnsets)
}

func TestNotifyDisconnect_UnsetsServerGroupValueAfterEvent(t *testing.T) {
	category.Set(t, category.Unit)
	sub := NewSubscriber("", nil, nil, nil, config.BuildTarget{}, "", "", "")
	noopDisconnectAmbientMooseFuncs(sub)

	var groupUnsetOrder, sendDisconnectCallOrder int
	var nextCall int
	var groupUnsets int

	sub.mooseFuncs.unsetServerGroupValue = func() uint32 {
		nextCall++
		groupUnsetOrder = nextCall
		groupUnsets++
		return 0
	}
	sub.mooseFuncs.sendDisconnect = func(
		_ moose.EventParams,
		_ moose.TargetConnectionParams,
		_ moose.ConnectionParams,
		_ int32,
		_ int32,
		_ *string,
	) uint32 {
		nextCall++
		sendDisconnectCallOrder = nextCall
		return 0
	}

	err := sub.NotifyDisconnect(events.DataDisconnect{
		EventStatus: events.StatusSuccess,
	})

	assert.NilError(t, err)
	assert.Equal(t, 1, groupUnsets)
	assert.Equal(t, true, sendDisconnectCallOrder < groupUnsetOrder)
}
func TestNotifyDedicatedServerStatus(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name         string
		status       string
		mooseErrCode uint32
		expectValue  bool
		expectErr    bool
	}{
		{
			name:        "running - sets true",
			status:      "running",
			expectValue: true,
		},
		{
			name:        "provisioning - sets true",
			status:      "provisioning",
			expectValue: true,
		},
		{
			name:        "new - sets false",
			status:      "new",
			expectValue: false,
		},
		{
			name:        "empty - sets false",
			status:      "",
			expectValue: false,
		},
		{
			name:         "setter fails - propagates error",
			status:       "running",
			mooseErrCode: 7,
			expectErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotValue bool

			s := &Subscriber{
				mooseFuncs: mooseFunctions{
					setDSEnabled: func(v bool) uint32 {
						gotValue = v
						return tt.mooseErrCode
					},
				},
			}

			err := s.NotifyDedicatedServerStatus(
				events.DataDedicatedServerStatus{Status: tt.status},
			)

			if tt.mooseErrCode == 0 {
				assert.Equal(t, tt.expectValue, gotValue)
			}
			if tt.expectErr {
				assert.Assert(t, err != nil)
			} else {
				assert.NilError(t, err)
			}
		})
	}
}

func TestVPNConnReasonToMoose(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		name              string
		trigger           events.VPNConnectionReason
		wantMooseTrigger  moose.NordvpnappVpnConnectionTrigger
		wantExceptionCode int32
		wantEventTrigger  moose.NordvpnappEventTrigger
	}{
		{
			name:              "server maintenance is app-triggered with ServerMaintenance trigger + 1000076",
			trigger:           events.VPNConnectionReasonServerMaintenance,
			wantMooseTrigger:  moose.NordvpnappVpnConnectionTriggerServerMaintenance,
			wantExceptionCode: 1000076,
			wantEventTrigger:  moose.NordvpnappEventTriggerApp,
		},
		{
			name:              "auto-connect is app-triggered with AutoConnectUserSetting trigger + -1",
			trigger:           events.VPNConnectionReasonAutoConnect,
			wantMooseTrigger:  moose.NordvpnappVpnConnectionTriggerAutoConnectUserSetting,
			wantExceptionCode: -1,
			wantEventTrigger:  moose.NordvpnappEventTriggerApp,
		},
		{
			name:              "none is user-triggered with None trigger + -1",
			trigger:           events.VPNConnectionReasonNone,
			wantMooseTrigger:  moose.NordvpnappVpnConnectionTriggerNone,
			wantExceptionCode: -1,
			wantEventTrigger:  moose.NordvpnappEventTriggerUser,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := vpnConnReasonToMoose(tt.trigger)
			assert.Equal(t, tt.wantMooseTrigger, got.trigger)
			assert.Equal(t, tt.wantExceptionCode, got.exceptionCode)
			assert.Equal(t, tt.wantEventTrigger, got.eventTrigger)
		})
	}
}

func TestUiItemType(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		name     string
		itemType string
		want     moose.NordvpnappUserInterfaceItemType
	}{
		{
			name:     "textbox maps to text box",
			itemType: "textbox",
			want:     moose.NordvpnappUserInterfaceItemTypeTextBox,
		},
		{
			name:     "click maps to button",
			itemType: "click",
			want:     moose.NordvpnappUserInterfaceItemTypeButton,
		},
		{
			name:     "empty maps to button",
			itemType: "",
			want:     moose.NordvpnappUserInterfaceItemTypeButton,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, uiItemType(tt.itemType))
		})
	}
}

func TestNotifyConnect_VPNConnReason(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		name                  string
		trigger               events.VPNConnectionReason
		wantEventTrigger      moose.NordvpnappEventTrigger
		wantConnectionTrigger moose.NordvpnappVpnConnectionTrigger
	}{
		{
			name:                  "server maintenance reconnect is app-triggered",
			trigger:               events.VPNConnectionReasonServerMaintenance,
			wantEventTrigger:      moose.NordvpnappEventTriggerApp,
			wantConnectionTrigger: moose.NordvpnappVpnConnectionTriggerServerMaintenance,
		},
		{
			name:                  "user connect is user-triggered",
			trigger:               events.VPNConnectionReasonNone,
			wantEventTrigger:      moose.NordvpnappEventTriggerUser,
			wantConnectionTrigger: moose.NordvpnappVpnConnectionTriggerNone,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := NewSubscriber("", nil, nil, nil, config.BuildTarget{}, "", "", "")
			var gotEvent moose.EventParams
			var gotConn moose.ConnectionParams
			sub.mooseFuncs.sendConnect = func(
				eventParams moose.EventParams,
				_ moose.TargetConnectionParams,
				_ moose.TargetConnectionAdditionalParams,
				connectionParams moose.ConnectionParams,
				_ moose.NordvpnappOptBool,
				_ int32,
				_ string,
				_ *string,
			) uint32 {
				gotEvent = eventParams
				gotConn = connectionParams
				return 0
			}

			err := sub.NotifyConnect(events.DataConnect{
				EventStatus:   events.StatusAttempt,
				VPNConnReason: tt.trigger,
			})

			assert.NilError(t, err)
			assert.Equal(t, tt.wantEventTrigger, gotEvent.EventTrigger)
			assert.Equal(t, tt.wantConnectionTrigger, gotConn.VpnConnectionTrigger)
		})
	}
}

func TestNotifyDisconnect_VPNConnReason(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		name              string
		trigger           events.VPNConnectionReason
		wantEventTrigger  moose.NordvpnappEventTrigger
		wantExceptionCode int32
	}{
		{
			name:              "server maintenance disconnect is app-triggered with code 1000076",
			trigger:           events.VPNConnectionReasonServerMaintenance,
			wantEventTrigger:  moose.NordvpnappEventTriggerApp,
			wantExceptionCode: 1000076,
		},
		{
			name:              "user disconnect is user-triggered with no exception code",
			trigger:           events.VPNConnectionReasonNone,
			wantEventTrigger:  moose.NordvpnappEventTriggerUser,
			wantExceptionCode: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := NewSubscriber("", nil, nil, nil, config.BuildTarget{}, "", "", "")
			var gotEvent moose.EventParams
			var gotCode int32
			sub.mooseFuncs.sendDisconnect = func(
				eventParams moose.EventParams,
				_ moose.TargetConnectionParams,
				_ moose.ConnectionParams,
				_ int32,
				exceptionCode int32,
				_ *string,
			) uint32 {
				gotEvent = eventParams
				gotCode = exceptionCode
				return 0
			}
			// NotifyDisconnect also updates context state after sending; no-op those so the test
			// does not depend on the native moose context.
			sub.mooseFuncs.unsetTPLiteCurrentState = func() uint32 { return 0 }
			sub.mooseFuncs.setServerCountryValue = func(_ string) uint32 { return 0 }
			sub.mooseFuncs.unsetServerGroupValue = func() uint32 { return 0 }
			sub.mooseFuncs.setIsOnVpnValue = func(_ bool) uint32 { return 0 }

			err := sub.NotifyDisconnect(events.DataDisconnect{
				EventStatus:   events.StatusSuccess,
				VPNConnReason: tt.trigger,
			})

			assert.NilError(t, err)
			assert.Equal(t, tt.wantEventTrigger, gotEvent.EventTrigger)
			assert.Equal(t, tt.wantExceptionCode, gotCode)
		})
	}
}
