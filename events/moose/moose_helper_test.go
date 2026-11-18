//go:build moose

package moose

import (
	"math"
	moose "moose/events"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"gotest.tools/v3/assert"
)

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

func TestVPNConnReasonToInternalType(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		name             string
		trigger          events.VPNConnectionReason
		wantMooseTrigger moose.NordvpnappVpnConnectionTrigger
	}{
		{
			name:             "server maintenance",
			trigger:          events.VPNConnectionReasonServerMaintenance,
			wantMooseTrigger: moose.NordvpnappVpnConnectionTriggerServerMaintenance,
		},
		{
			name:             "auto-connect",
			trigger:          events.VPNConnectionReasonAutoConnect,
			wantMooseTrigger: moose.NordvpnappVpnConnectionTriggerAutoConnectUserSetting,
		},
		{
			name:             "none",
			trigger:          events.VPNConnectionReasonNone,
			wantMooseTrigger: moose.NordvpnappVpnConnectionTriggerNone,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantMooseTrigger, vpnConnReasonToInternalType(tt.trigger))
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
