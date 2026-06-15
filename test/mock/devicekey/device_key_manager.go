package devicekey

import devicekey "github.com/NordSecurity/nordvpn-linux/device_key"

type MockDeviceKeyManager struct {
	DedicatedServerRegistrationData       *devicekey.DedicatedServersConnectionData
	DedicatedServerForcedRegistrationData *devicekey.DedicatedServersConnectionData

	WasKeyRegistered        bool
	WasKeyForceRegistered   bool
	WasDeviceKeyInvalidated bool
}

func (m *MockDeviceKeyManager) CheckAndRegisterDedicatedServers() *devicekey.DedicatedServersConnectionData {
	if m.DedicatedServerRegistrationData == nil {
		return nil
	}
	m.WasKeyRegistered = true
	return m.DedicatedServerRegistrationData
}

func (m *MockDeviceKeyManager) ForceRegisterDedicatedServers() *devicekey.DedicatedServersConnectionData {
	if m.DedicatedServerForcedRegistrationData == nil {
		return nil
	}
	m.WasKeyForceRegistered = true
	return m.DedicatedServerForcedRegistrationData
}

func (m *MockDeviceKeyManager) InvalidateDeviceKeyData() error {
	m.WasDeviceKeyInvalidated = true
	return nil
}
