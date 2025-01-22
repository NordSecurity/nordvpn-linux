//go:build moose

package moose

import (
	"math"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/core"
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
