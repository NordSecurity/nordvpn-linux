package devicekey

type MockDeviceKeyManager struct {
	WasDeviceKeyInvalidated bool
}

func (m *MockDeviceKeyManager) CheckAndRegisterDedicatedServers() bool {
	return false
}

func (m *MockDeviceKeyManager) InvalidateDeviceKeyData() error {
	m.WasDeviceKeyInvalidated = true
	return nil
}
