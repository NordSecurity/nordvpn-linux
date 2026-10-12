package core_test

import "github.com/NordSecurity/nordvpn-linux/core"

type MockSubscriptionAPI struct {
}

func NewMockSubscriptionAPI() *MockSubscriptionAPI {
	return &MockSubscriptionAPI{}
}

func (m *MockSubscriptionAPI) Orders() ([]core.Order, error) {
	return []core.Order{}, nil
}

func (m *MockSubscriptionAPI) Payments() ([]core.PaymentResponse, error) {
	return []core.PaymentResponse{}, nil
}

func (m *MockSubscriptionAPI) Purchases() (core.PurchasesResult, error) {
	return core.PurchasesResult{}, nil
}
