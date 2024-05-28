package service

type MockNorduserCombinedService struct {
	Enabled   []uint32
	EnableErr error

	Disabled    []uint32
	DeisableErr error
}

func NewMockNorduserCombinedService() MockNorduserCombinedService {
	return MockNorduserCombinedService{
		Enabled:  []uint32{},
		Disabled: []uint32{},
	}
}

func (m *MockNorduserCombinedService) Enable(uid uint32, _ uint32, _ string) error {
	if m.EnableErr != nil {
		return m.EnableErr
	}

	m.Enabled = append(m.Enabled, uid)
	return nil
}

func (m *MockNorduserCombinedService) Disable(uid uint32) error {
	if m.DeisableErr != nil {
		return m.DeisableErr
	}

	m.Disabled = append(m.Disabled, uid)

	return nil
}

func (m *MockNorduserCombinedService) Stop(uint32, bool) error { return nil }

func (m *MockNorduserCombinedService) Restart(uint32) error { return nil }

func (m *MockNorduserCombinedService) StopAll() {}

func (m *MockNorduserCombinedService) DisableAll() {}
