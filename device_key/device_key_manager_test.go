package devicekey

import (
	"net/netip"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const privateKey = "0001"

type mockDedicatedServersAPI struct {
	deviceUUID              string
	updateDeviceResponse    core.DevicesResponse
	registerDeviceErr       error
	wasRegisterDeviceCalled bool
	wasUpdateDeviceCalled   bool
}

// getUnsetWasCalled returns wasCalled and sets it to false
func (m *mockDedicatedServersAPI) getUnsetWasCalled() bool {
	wasCalled := m.wasRegisterDeviceCalled
	m.wasRegisterDeviceCalled = false
	return wasCalled
}

func (m *mockDedicatedServersAPI) RegisterDevice(req core.DevicesRequest) (core.DevicesResponse, error) {
	m.wasRegisterDeviceCalled = true
	return core.DevicesResponse{
		UUID:               m.deviceUUID,
		HardwareIdentifier: req.HardwareIdentifier,
		PublicKey:          req.PublicKey,
		OS:                 req.Os,
		Type:               req.Type,
		Name:               req.Name,
	}, m.registerDeviceErr
}

func (m *mockDedicatedServersAPI) UpdateDevice(uuid uuid.UUID, req core.UpdateDeviceRequest) (core.DevicesResponse, error) {
	m.wasUpdateDeviceCalled = true
	return core.DevicesResponse{
		UUID:               m.updateDeviceResponse.UUID,
		HardwareIdentifier: m.updateDeviceResponse.PublicKey,
		PublicKey:          m.updateDeviceResponse.PublicKey,
		OS:                 m.updateDeviceResponse.OS,
		Type:               m.updateDeviceResponse.Type,
		Name:               m.updateDeviceResponse.Name,
	}, nil
}

func (m *mockDedicatedServersAPI) DedicatedServers() (core.DedicatedServers, error) {
	return core.DedicatedServers{}, nil
}

func (m *mockDedicatedServersAPI) Connect(string, core.ConnectRequest) (core.ConnectResponse, error) {
	return core.ConnectResponse{}, nil
}

type generator struct {
	KeyGenerator
	privateKey string
}

func (*generator) Public(string) string { return "0001" }
func (g *generator) Private() string {
	return g.privateKey
}

type registry struct {
	mesh.Registry
	registerErrors []error
	wasCalled      bool
}

// getUnsetWasCalled returns wasCalled and sets it to false
func (m *registry) getUnsetWasCalled() bool {
	wasCalled := m.wasCalled
	m.wasCalled = false
	return wasCalled
}

const registryUUID = "00000000-0000-0000-0000-000000000001"
const registryIP = "0.0.0.1"

func (r *registry) popRegisterErr() error {
	if len(r.registerErrors) == 0 {
		return nil
	}

	err := r.registerErrors[len(r.registerErrors)-1]
	r.registerErrors = r.registerErrors[:len(r.registerErrors)-1]
	return err
}

func (r *registry) Register(token string, self mesh.Machine) (*mesh.Machine, error) {
	if registerError := r.popRegisterErr(); registerError != nil {
		return nil, registerError
	}

	return &mesh.Machine{
		ID:      uuid.MustParse(registryUUID),
		Address: netip.MustParseAddr(registryIP),
	}, nil
}

type delayChecker struct {
	called bool
}

func (d *delayChecker) Delay(duration time.Duration) {
	d.called = true
}

func TestForceRegister_NotYetRegistered(t *testing.T) {
	category.Set(t, category.Unit)

	delayChecker := delayChecker{}

	cm := &mock.ConfigManager{}
	rc := NewDeviceKeyManager(cm, &generator{privateKey: privateKey}, &registry{}, &mockDedicatedServersAPI{})
	rc.delayFunc = delayChecker.Delay

	err := rc.ForceRegisterMeshnet()
	assert.NoError(t, err)
	assert.Equal(t, privateKey, cm.Cfg.DeviceKey)
	assert.Equal(t, registryUUID, cm.Cfg.MeshDevice.ID.String())
	assert.Equal(t, registryIP, cm.Cfg.MeshDevice.Address.String())
	assert.True(t, delayChecker.called, "App did not block after registering a new mesh key.")
}

func TestForceRegister_AlreadyRegistered(t *testing.T) {
	category.Set(t, category.Unit)

	delayChecker := delayChecker{}

	cm := &mock.ConfigManager{
		Cfg: &config.Config{
			DeviceKey: "0002",
			MeshDevice: &mesh.Machine{
				ID:      uuid.MustParse("00000000-0000-0000-0000-000000000002"),
				Address: netip.MustParseAddr("0.0.0.2"),
			},
		},
	}
	rc := NewDeviceKeyManager(cm, &generator{privateKey: privateKey}, &registry{}, &mockDedicatedServersAPI{})
	rc.delayFunc = delayChecker.Delay

	err := rc.ForceRegisterMeshnet()
	assert.NoError(t, err)
	assert.NotEqual(t, privateKey, cm.Cfg.DeviceKey) // Existing private key should be kept
	assert.Equal(t, registryUUID, cm.Cfg.MeshDevice.ID.String())
	assert.Equal(t, registryIP, cm.Cfg.MeshDevice.Address.String())
	assert.False(t, delayChecker.called, "App blocked when no new mesh key was registered.")
}

func TestCheckAndRegisterMeshnet_NotYetRegistered(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name          string
		cfg           *config.Config
		newPrivateKey bool
	}{
		{
			name:          "empty config",
			cfg:           &config.Config{},
			newPrivateKey: true,
		},
		{
			name: "no private key",
			cfg: &config.Config{
				MeshDevice: &mesh.Machine{
					ID:      uuid.MustParse("00000000-0000-0000-0000-000000000002"),
					Address: netip.MustParseAddr("0.0.0.2"),
				}},
			newPrivateKey: true,
		},
		{
			name: "no MeshDevice",
			cfg: &config.Config{
				DeviceKey: "0002",
			},
		},
		{
			name: "no ID",
			cfg: &config.Config{
				DeviceKey: "0002",
				MeshDevice: &mesh.Machine{
					Address: netip.MustParseAddr("0.0.0.2"),
				}},
		},
		{
			name: "no address",
			cfg: &config.Config{
				DeviceKey: "0002",
				MeshDevice: &mesh.Machine{
					ID: uuid.MustParse("00000000-0000-0000-0000-000000000002"),
				}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cm := &mock.ConfigManager{Cfg: test.cfg}
			rc := NewDeviceKeyManager(cm, &generator{privateKey: privateKey}, &registry{}, &mockDedicatedServersAPI{})
			delayChecker := delayChecker{}
			rc.delayFunc = delayChecker.Delay
			ok := rc.CheckAndRegisterMeshnet()
			assert.True(t, ok)
			assert.Equal(t, test.newPrivateKey, privateKey == cm.Cfg.DeviceKey)
			assert.Equal(t, registryUUID, cm.Cfg.MeshDevice.ID.String())
			assert.Equal(t, registryIP, cm.Cfg.MeshDevice.Address.String())
			if test.newPrivateKey {
				assert.True(t, delayChecker.called, "App did not block after registering a new mesh key.")
			}
		})
	}
}

func TestCheckAndRegisterMeshnet_AlreadyRegistered(t *testing.T) {
	category.Set(t, category.Unit)

	cm := &mock.ConfigManager{
		Cfg: &config.Config{
			DeviceKey: "0002",
			MeshDevice: &mesh.Machine{
				ID:      uuid.MustParse("00000000-0000-0000-0000-000000000002"),
				Address: netip.MustParseAddr("0.0.0.2"),
			},
		},
	}
	delayChecker := delayChecker{}
	rc := NewDeviceKeyManager(cm, &generator{privateKey: privateKey}, &registry{}, &mockDedicatedServersAPI{})
	rc.delayFunc = delayChecker.Delay
	ok := rc.CheckAndRegisterMeshnet()
	assert.True(t, ok)
	// Registration should not be done, values should not change
	assert.NotEqual(t, privateKey, cm.Cfg.DeviceKey)
	assert.NotEqual(t, registryUUID, cm.Cfg.MeshDevice.ID.String())
	assert.NotEqual(t, registryIP, cm.Cfg.MeshDevice.Address.String())
	assert.False(t, delayChecker.called, "App blocked when no new mesh key was registered.")
}

func TestForceRegisterMeshnet_ConflictHandling(t *testing.T) {
	category.Set(t, category.Unit)

	const newPrivateKey = "0002"

	cm := &mock.ConfigManager{Cfg: &config.Config{
		DeviceKey: privateKey,
	}}
	delayChecker := delayChecker{}
	registryMock := &registry{}
	registryMock.registerErrors = []error{core.ErrConflict}
	rc := NewDeviceKeyManager(cm, &generator{privateKey: newPrivateKey}, registryMock, &mockDedicatedServersAPI{})
	rc.delayFunc = delayChecker.Delay

	err := rc.ForceRegisterMeshnet()
	assert.NoError(t, err)
	// Registration should not be done, values should not change
	assert.Equal(t, newPrivateKey, cm.Cfg.DeviceKey)
	assert.Equal(t, registryUUID, cm.Cfg.MeshDevice.ID.String())
	assert.Equal(t, registryIP, cm.Cfg.MeshDevice.Address.String())
	assert.True(t, delayChecker.called, "App did not block when new mesh key was registered.")
}

func TestCheckAndRegisterDedicatedServers(t *testing.T) {
	category.Set(t, category.Unit)

	deviceID := uuid.MustParse("a559a0a1-ed1d-4fe2-8173-258afcc68538")

	tests := []struct {
		name                string
		cfgDeviceKey        string
		cfgDeviceID         uuid.UUID
		apiDeviceID         string
		generatedPrivateKey string
		shouldCallAPI       bool
		expectedDeviceKey   string
		expectedDeviceID    uuid.UUID
	}{
		{
			name:                "not registered",
			apiDeviceID:         deviceID.String(),
			shouldCallAPI:       true,
			generatedPrivateKey: privateKey,
			expectedDeviceKey:   privateKey,
			expectedDeviceID:    deviceID,
		},
		{
			name:              "already registered",
			cfgDeviceKey:      privateKey,
			cfgDeviceID:       deviceID,
			shouldCallAPI:     false,
			expectedDeviceKey: privateKey,
			expectedDeviceID:  deviceID,
		},
		{
			name:              "key exists but device id doesn't",
			cfgDeviceKey:      privateKey,
			apiDeviceID:       deviceID.String(),
			shouldCallAPI:     true,
			expectedDeviceKey: privateKey,
			expectedDeviceID:  deviceID,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cm := &mock.ConfigManager{
				Cfg: &config.Config{
					DeviceKey:  test.cfgDeviceKey,
					DeviceUUID: test.cfgDeviceID,
				},
			}

			mockDedicatedServer := mockDedicatedServersAPI{deviceUUID: test.apiDeviceID}

			rc := NewDeviceKeyManager(cm,
				&generator{privateKey: test.generatedPrivateKey},
				&registry{},
				&mockDedicatedServer)

			deviceData := rc.CheckAndRegisterDedicatedServers()
			assert.NotNil(t, deviceData)
			assert.Equal(t, test.shouldCallAPI, mockDedicatedServer.wasRegisterDeviceCalled)
			assert.Equal(t, test.expectedDeviceKey, cm.Cfg.DeviceKey)
			assert.Equal(t, test.expectedDeviceID, cm.Cfg.DeviceUUID)
		})
	}
}

func TestCheckAndRegisterDedicatedServers_ConflictHandling(t *testing.T) {
	category.Set(t, category.Unit)

	deviceUUID := uuid.MustParse("6f85c8df-d585-421f-b82a-b77eed19b35f")
	newPrivateKey := "00002"

	configManagerMock := mock.ConfigManager{Cfg: &config.Config{
		DeviceKey: privateKey,
		MeshDevice: &mesh.Machine{
			ID:      uuid.MustParse("00000000-0000-0000-0000-000000000002"),
			Address: netip.MustParseAddr("0.0.0.2"),
		},
	}}
	keyGeneratorMock := generator{privateKey: newPrivateKey}

	dedicatedServersAPIMock := mockDedicatedServersAPI{
		deviceUUID: deviceUUID.String(),
		updateDeviceResponse: core.DevicesResponse{
			UUID: deviceUUID.String(),
		},
		registerDeviceErr: core.ErrConflict}
	meshRegistryMock := registry{}

	rc := NewDeviceKeyManager(&configManagerMock, &keyGeneratorMock, &meshRegistryMock, &dedicatedServersAPIMock)
	deviceData := rc.CheckAndRegisterDedicatedServers()
	assert.NotNil(t, deviceData)
	assert.True(t, dedicatedServersAPIMock.wasRegisterDeviceCalled)
	assert.True(t, dedicatedServersAPIMock.wasUpdateDeviceCalled)
	assert.Equal(t, newPrivateKey, configManagerMock.Cfg.DeviceKey,
		"New device key should be generated when devices API returns a conflict(409).")
	assert.Equal(t, deviceUUID, configManagerMock.Cfg.DeviceUUID, "Device UUID was not updated.")
	assert.Nil(t, configManagerMock.Cfg.MeshDevice,
		"Meshnet config was not invalidated after generating new private key.")
}

func TestDeviceKeyManager_DedicatedServersMeshnetInteraction(t *testing.T) {
	category.Set(t, category.Unit)

	deviceUUID := uuid.MustParse("6f85c8df-d585-421f-b82a-b77eed19b35f")

	configManagerMock := mock.ConfigManager{}
	keyGeneratorMock := generator{privateKey: privateKey}
	dedicatedServersAPIMock := mockDedicatedServersAPI{deviceUUID: deviceUUID.String()}
	meshRegistryMock := registry{}

	rc := NewDeviceKeyManager(&configManagerMock, &keyGeneratorMock, &meshRegistryMock, &dedicatedServersAPIMock)
	delayChecker := delayChecker{}
	rc.delayFunc = delayChecker.Delay

	err := rc.ForceRegisterMeshnet()

	assert.NoError(t, err)
	assert.Equal(t, privateKey, configManagerMock.Cfg.DeviceKey)
	assert.Equal(t, registryUUID, configManagerMock.Cfg.MeshDevice.ID.String())
	assert.Equal(t, registryIP, configManagerMock.Cfg.MeshDevice.Address.String())
	assert.True(t, delayChecker.called, "App did not block after registering a new mesh key.")

	// CheckAndRegisterDedicatedServers called for the first time, it register the key for dedicated servers and set the
	// DeviceUUID but not touch the DeviceKey iteself.
	deviceData := rc.CheckAndRegisterDedicatedServers()
	assert.NotNil(t, deviceData, "Key not registered for dedicated servers when expected.")
	assert.True(t, dedicatedServersAPIMock.getUnsetWasCalled(), "Dedicated servers api was not called when expected.")
	assert.Equal(t, privateKey, configManagerMock.Cfg.DeviceKey)
	assert.Equal(t, deviceUUID, configManagerMock.Cfg.DeviceUUID)

	// CheckAndRegisterDedicatedServers called for the second time, it should not do anything since the key is already
	// registered.
	deviceData = rc.CheckAndRegisterDedicatedServers()
	assert.NotNil(t, deviceData, "Key not registered for dedicated servers when expected.")
	assert.False(t, dedicatedServersAPIMock.getUnsetWasCalled(), "Dedicated servers api was called when not expected.")

	// CheckAndRegisterMeshnet called for the second time, it should not do anything since the key is already
	// registered.
	registered := rc.CheckAndRegisterMeshnet()
	assert.True(t, registered, "Key not registered for meshnet when expected.")
	assert.False(t, meshRegistryMock.getUnsetWasCalled(), "Meshnet api was called when not expected.")

	// ForceRegisterMeshnet called for the second time. Since it will fail with a conflict error, it should force a key
	// regeneration and invalidation of all of the previous key related data.
	meshRegistryMock.registerErrors = []error{core.ErrConflict}
	newKey := "0002"
	keyGeneratorMock.privateKey = newKey
	err = rc.ForceRegisterMeshnet()
	assert.NoError(t, err)
	assert.Equal(t, newKey, configManagerMock.Cfg.DeviceKey)
	assert.Equal(t, registryUUID, configManagerMock.Cfg.MeshDevice.ID.String())
	assert.Equal(t, registryIP, configManagerMock.Cfg.MeshDevice.Address.String())
	assert.Equal(t, uuid.Nil, configManagerMock.Cfg.DeviceUUID,
		"DeviceID was not invalidated after regenerating DeviceKey.")
	assert.True(t, delayChecker.called, "App did not block after registering a new mesh key.")

	// CheckAndRegisterDedicatedServers called after the key was regenerated, it should register the new key.
	deviceData = rc.CheckAndRegisterDedicatedServers()
	assert.NotNil(t, deviceData, "Key not registered for dedicated servers when expected.")
	assert.True(t, dedicatedServersAPIMock.getUnsetWasCalled(), "Dedicated servers api was not called when expected.")
	assert.Equal(t, newKey, configManagerMock.Cfg.DeviceKey)
	assert.Equal(t, deviceUUID, configManagerMock.Cfg.DeviceUUID)
}
