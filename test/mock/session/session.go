package session

type MockSessionStore struct {
	RenewErr            error
	InvalidateErr       error
	RenewCallCount      int
	InvalidateCallCount int
}

func (m *MockSessionStore) Renew() error {
	m.RenewCallCount++
	return m.RenewErr
}

func (m *MockSessionStore) Invalidate(reason error) error {
	m.InvalidateCallCount++
	return m.InvalidateErr
}

type MockAccessTokenSessionStore struct {
	RenewFunc      func() error
	InvalidateFunc func(reason error) error
	GetTokenFunc   func() string

	RenewCallCount      int
	InvalidateCallCount int
	GetTokenCallCount   int
}

func (m *MockAccessTokenSessionStore) Renew() error {
	m.RenewCallCount++
	return m.RenewFunc()
}

func (m *MockAccessTokenSessionStore) Invalidate(reason error) error {
	m.InvalidateCallCount++
	return m.InvalidateFunc(reason)
}

func (m *MockAccessTokenSessionStore) GetToken() string {
	m.GetTokenCallCount++
	return m.GetTokenFunc()
}
