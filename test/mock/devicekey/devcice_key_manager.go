package devicekey

type MockDeviceKeyManager struct {
	CheckAndRegisterDedicatedServersStatus bool

	WasKeyRegistered        bool
	WasDeviceKeyInvalidated bool
}

func (m *MockDeviceKeyManager) CheckAndRegisterDedicatedServers() bool {
	if !m.CheckAndRegisterDedicatedServersStatus {
		return m.CheckAndRegisterDedicatedServersStatus
	}
	m.WasKeyRegistered = true
	return m.CheckAndRegisterDedicatedServersStatus
}

func (m *MockDeviceKeyManager) InvalidateDeviceKeyData() error {
	m.WasDeviceKeyInvalidated = true
	return nil
}
