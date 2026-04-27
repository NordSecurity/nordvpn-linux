package devicekey

type MockDeviceKeyInvalidator struct {
	deviceKeyInvalidated bool
}

func (m *MockDeviceKeyInvalidator) InvalidateDeviceKeyData() error {
	m.deviceKeyInvalidated = true
	return nil
}
