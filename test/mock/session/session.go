package session

type MockSessionStore struct {
	RenewErr      error
	InvalidateErr error
}

func (m *MockSessionStore) Renew() error {
	return m.RenewErr
}

func (m *MockSessionStore) Invalidate(reason error) error {
	return m.InvalidateErr
}
