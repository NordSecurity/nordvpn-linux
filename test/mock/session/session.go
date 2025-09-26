package session

import "github.com/NordSecurity/nordvpn-linux/session"

type MockSessionStore struct {
	RenewErr             error
	HandleErrorErr       error
	RenewCallCount       int
	HandleErrorCallCount int
}

func (m *MockSessionStore) Renew(opts ...session.RenewalOption) error {
	m.RenewCallCount++
	return m.RenewErr
}

func (m *MockSessionStore) HandleError(reason error) error {
	m.HandleErrorCallCount++
	return m.HandleErrorErr
}

func (m *MockSessionStore) GetToken() string {
	return ""
}

type MockAccessTokenSessionStore struct {
	RenewFunc       func(opts ...session.RenewalOption) error
	HandleErrorFunc func(reason error) error
	GetTokenFunc    func() string

	RenewCallCount       int
	HandleErrorCallCount int
	GetTokenCallCount    int
}

func (m *MockAccessTokenSessionStore) Renew(opts ...session.RenewalOption) error {
	m.RenewCallCount++
	if m.RenewFunc != nil {
		return m.RenewFunc(opts...)
	}
	return nil
}

func (m *MockAccessTokenSessionStore) HandleError(reason error) error {
	m.HandleErrorCallCount++
	if m.HandleErrorFunc != nil {
		return m.HandleErrorFunc(reason)
	}
	return nil
}

func (m *MockAccessTokenSessionStore) GetToken() string {
	m.GetTokenCallCount++
	if m.GetTokenFunc != nil {
		return m.GetTokenFunc()
	}
	return ""
}
