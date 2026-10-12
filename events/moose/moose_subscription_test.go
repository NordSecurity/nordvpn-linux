//go:build moose

package moose

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"gotest.tools/v3/assert"
)

func TestFormatPurchases(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		name      string
		purchases []core.Purchase
		want      string
	}{
		{
			name:      "empty input returns empty string",
			purchases: nil,
			want:      "",
		},
		{
			name:      "empty slice returns empty string",
			purchases: []core.Purchase{},
			want:      "",
		},
		{
			name: "single valid purchase",
			purchases: []core.Purchase{{
				Plans:         []core.PurchasePlan{plan(384, "web")},
				Payments:      []core.PurchasePayment{payment("2020-01-15 10:30:00", "done")},
				Subscriptions: []core.PurchaseSubscription{subscription(1, 2, "year")},
			}},
			want: "2020-01:384:2:year:web:done:100.00:USD:stripe",
		},
		{
			name: "purchase with no plans is skipped",
			purchases: []core.Purchase{{
				Plans:         nil,
				Payments:      []core.PurchasePayment{payment("2020-01-15 10:30:00", "done")},
				Subscriptions: []core.PurchaseSubscription{subscription(1, 2, "year")},
			}},
			want: "",
		},
		{
			name: "disallowed status is skipped",
			purchases: []core.Purchase{{
				Plans:         []core.PurchasePlan{plan(384, "web")},
				Payments:      []core.PurchasePayment{payment("2020-01-15 10:30:00", "waiting_for_confirmation")},
				Subscriptions: []core.PurchaseSubscription{subscription(1, 2, "year")},
			}},
			want: "",
		},
		{
			name: "payment with no matching subscription is skipped",
			purchases: []core.Purchase{{
				Plans:         []core.PurchasePlan{plan(384, "web")},
				Payments:      []core.PurchasePayment{payment("2020-01-15 10:30:00", "done")},
				Subscriptions: []core.PurchaseSubscription{subscription(999, 2, "year")},
			}},
			want: "",
		},
		{
			name: "payment with unparseable created_at is skipped",
			purchases: []core.Purchase{{
				Plans:         []core.PurchasePlan{plan(384, "web")},
				Payments:      []core.PurchasePayment{payment("not-a-date", "done")},
				Subscriptions: []core.PurchaseSubscription{subscription(1, 2, "year")},
			}},
			want: "",
		},
		{
			name: "multiple payments in one purchase produce multiple elements",
			purchases: []core.Purchase{{
				Plans: []core.PurchasePlan{plan(384, "web")},
				Payments: []core.PurchasePayment{
					payment("2020-01-15 10:30:00", "trial"),
					payment("2020-02-15 10:30:00", "done"),
				},
				Subscriptions: []core.PurchaseSubscription{subscription(1, 2, "year")},
			}},
			want: "2020-01:384:2:year:web:trial:100.00:USD:stripe," +
				"2020-02:384:2:year:web:done:100.00:USD:stripe",
		},
		{
			name: "multiple purchases are all included",
			purchases: []core.Purchase{
				{
					Plans:         []core.PurchasePlan{plan(384, "web")},
					Payments:      []core.PurchasePayment{payment("2020-01-15 10:30:00", "done")},
					Subscriptions: []core.PurchaseSubscription{subscription(1, 2, "year")},
				},
				{
					Plans:         []core.PurchasePlan{plan(864, "android")},
					Payments:      []core.PurchasePayment{payment("2022-01-15 10:30:00", "done")},
					Subscriptions: []core.PurchaseSubscription{subscription(1, 1, "month")},
				},
			},
			want: "2020-01:384:2:year:web:done:100.00:USD:stripe," +
				"2022-01:864:1:month:android:done:100.00:USD:stripe",
		},
		{
			name: "only first plan is used when several exist, not sure how to handle this",
			purchases: []core.Purchase{{
				Plans: []core.PurchasePlan{
					plan(111, "web"),
					plan(222, "android"),
				},
				Payments:      []core.PurchasePayment{payment("2020-01-15 10:30:00", "done")},
				Subscriptions: []core.PurchaseSubscription{subscription(1, 2, "year")},
			}},
			want: "2020-01:111:2:year:web:done:100.00:USD:stripe",
		},
		{
			name: "allowed status mixed with disallowed keeps only allowed",
			purchases: []core.Purchase{{
				Plans: []core.PurchasePlan{plan(384, "web")},
				Payments: []core.PurchasePayment{
					payment("2020-01-15 10:30:00", "done"),
					payment("2020-02-15 10:30:00", "expired"),
					payment("2020-03-15 10:30:00", "refunded"),
				},
				Subscriptions: []core.PurchaseSubscription{subscription(1, 2, "year")},
			}},
			want: "2020-01:384:2:year:web:done:100.00:USD:stripe," +
				"2020-03:384:2:year:web:refunded:100.00:USD:stripe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatPurchases(tt.purchases)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatPurchases_StatusFiltering(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		status  string
		allowed bool
	}{
		{"done", true},
		{"chargeback", true},
		{"refunded", true},
		{"partially_refunded", true},
		{"trial", true},
		{"waiting_for_confirmation", false},
		{"expired", false},
		{"pending", false},
		{"failed", false},
		{"cancelled", false},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			purchases := []core.Purchase{{
				Plans:         []core.PurchasePlan{plan(384, "web")},
				Payments:      []core.PurchasePayment{payment("2020-01-15 10:30:00", tt.status)},
				Subscriptions: []core.PurchaseSubscription{subscription(1, 2, "year")},
			}}

			got := formatPurchases(purchases)

			if tt.allowed {
				assert.Assert(t, got != "", "payment status %q should produce output", tt.status)
			} else {
				assert.Equal(t, got, "", "payment status %q should be filtered out", tt.status)
			}
		})
	}
}
