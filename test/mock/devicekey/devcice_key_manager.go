package devicekey

import devicekey "github.com/NordSecurity/nordvpn-linux/device_key"

type MockDeviceKeyManager struct {
	DedicatedServerRegistrationData *devicekey.DedicatedServersRegistrationData

	WasKeyRegistered        bool
	WasDeviceKeyInvalidated bool
}

func (m *MockDeviceKeyManager) CheckAndRegisterDedicatedServers() *devicekey.DedicatedServersRegistrationData {
	if m.DedicatedServerRegistrationData == nil {
		return nil
	}
	m.WasKeyRegistered = true
	return m.DedicatedServerRegistrationData
}

func (m *MockDeviceKeyManager) InvalidateDeviceKeyData() error {
	m.WasDeviceKeyInvalidated = true
	return nil
}
