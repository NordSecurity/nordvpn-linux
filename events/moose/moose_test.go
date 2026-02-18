//go:build moose

package moose

import (
	"math"
	moose "moose/events"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

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
		name                            string
		currentConsentState             config.AnalyticsConsent
		currentOptInState               bool
		newConsentState                 config.AnalyticsConsent
		consentErrCode                  uint32
		expectedEssentialAnalyticsState bool
		shouldFail                      bool
	}{
		{
			name:                            "undefined to enabled success",
			currentConsentState:             config.ConsentUndefined,
			currentOptInState:               false,
			newConsentState:                 config.ConsentGranted,
			expectedEssentialAnalyticsState: true,
		},
		{
			name:                            "undefined to disabled success",
			currentConsentState:             config.ConsentUndefined,
			currentOptInState:               false,
			newConsentState:                 config.ConsentDenied,
			expectedEssentialAnalyticsState: false,
		},
		{
			name:                            "enabled to disabled success",
			currentConsentState:             config.ConsentGranted,
			currentOptInState:               true,
			newConsentState:                 config.ConsentDenied,
			expectedEssentialAnalyticsState: false,
		},
		{
			name:                            "disabled to enabled success",
			currentConsentState:             config.ConsentDenied,
			currentOptInState:               true,
			newConsentState:                 config.ConsentGranted,
			expectedEssentialAnalyticsState: true,
		},
		{
			name:                            "undefined to enabled failure to consent",
			currentConsentState:             config.ConsentUndefined,
			currentOptInState:               false,
			newConsentState:                 config.ConsentGranted,
			consentErrCode:                  1,
			expectedEssentialAnalyticsState: false,
		},
		{
			name:                            "undefined to disabled failure to consent",
			currentConsentState:             config.ConsentUndefined,
			currentOptInState:               false,
			newConsentState:                 config.ConsentDenied,
			consentErrCode:                  1,
			expectedEssentialAnalyticsState: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			consentFunc := func(enable bool) uint32 {
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
				canSendAllEvents: atomic.Bool{},
			}

			s.canSendAllEvents.Store(test.currentOptInState)
			err := s.changeConsentState(test.newConsentState)

			if test.consentErrCode != 0 || test.shouldFail {
				assert.Assert(t, err != nil)
			}

			assert.Equal(t, test.expectedEssentialAnalyticsState, s.canSendAllEvents.Load(), "Unexpected consent state saved.")
		})
	}
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
