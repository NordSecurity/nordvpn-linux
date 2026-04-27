package devicekey

type MockDeviceKeyInvalidator struct {
	WasDeviceKeyInvalidated bool
}

func (m *MockDeviceKeyInvalidator) InvalidateDeviceKeyData() error {
	m.WasDeviceKeyInvalidated = true
	return nil
}
