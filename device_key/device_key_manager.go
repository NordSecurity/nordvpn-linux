package devicekey

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	cmesh "github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
	sysinfo "github.com/NordSecurity/nordvpn-linux/sysinfo"

	"github.com/google/uuid"
)

// KeyGenerator for use in device key generation.
type KeyGenerator interface {
	// Private returns base64 encoded private key
	Private() string
	// Public expects base64 encoded private key and returns base64 encoded public key
	Public(string) string
}

// DelayFunc blocks the app for a duration of time
type DelayFunc func(duration time.Duration)

type MeshnetDeviceKeyManager interface {
	// RegisterMeshnet checks if the device key is registered for meshnet and registers it if it isn't.
	CheckAndRegisterMeshnet() bool
	// ForceRegisterMeshnet registers device key for meshnet.
	ForceRegisterMeshnet() error
}

type DedicatedServersKeyManager interface {
	CheckAndRegisterDedicatedServers() bool
}

// DeviceKeyManagerImpl does both registration checks and registration, if it's not done.
type DeviceKeyManagerImpl struct {
	configManager       config.Manager
	keyGenerator        KeyGenerator
	meshnetRegistery    cmesh.Registry
	dedicatedServersAPI core.DedicatedServersAPI
	mu                  sync.Mutex
	delayFunc           DelayFunc
}

// NewDeviceKeyManager is a default constructor for RegisteringChecker.
func NewDeviceKeyManager(
	configManager config.Manager,
	keyGenerator KeyGenerator,
	meshnetRegistery cmesh.Registry,
	dedicatedServersAPI core.DedicatedServersAPI,
) *DeviceKeyManagerImpl {
	return &DeviceKeyManagerImpl{
		configManager:       configManager,
		keyGenerator:        keyGenerator,
		meshnetRegistery:    meshnetRegistery,
		dedicatedServersAPI: dedicatedServersAPI,
		delayFunc:           time.Sleep}
}

func isMeshnetRegistrationInfoCorrect(cfg config.Config) bool {
	return cfg.DeviceKey != "" &&
		cfg.MeshDevice != nil &&
		cfg.MeshDevice.ID != uuid.Nil &&
		cfg.MeshDevice.Address.IsValid()
}

func isDedicatedServersRegistrationInfoCorrect(cfg config.Config) bool {
	return cfg.DeviceKey != "" && cfg.DeviceUUID != uuid.Nil
}

func invalidateKeyData(cfg config.Config) config.Config {
	cfg.MeshDevice = nil
	cfg.DeviceUUID = uuid.Nil
	return cfg
}

// getDeviceKey returns device key from cfg or generates a new key if it doesn't exist. Returns true if key was not
// found in the provided config.
func (d *DeviceKeyManagerImpl) getDeviceKey(cfg config.Config) (string, bool) {
	if cfg.DeviceKey == "" {
		return d.keyGenerator.Private(), true
	}

	return cfg.DeviceKey, false
}

// RegisterMeshnet registers the device key for meshnet if it isn't registered and returns true if it is successfully
// registered.
//
// Thread-safe.
func (d *DeviceKeyManagerImpl) CheckAndRegisterMeshnet() bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	var cfg config.Config
	if err := d.configManager.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return false
	}

	if isMeshnetRegistrationInfoCorrect(cfg) && cfg.DeviceKey != "" {
		return true
	}

	newConfig, err := d.registerKey(cfg, d.registerMeshnet)
	if err != nil {
		log.Println(internal.ErrorPrefix, "failed to register new device key: %s", err)
		return false
	}

	if err := d.configManager.SaveWith(keyConfig(newConfig)); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return false
	}

	return isMeshnetRegistrationInfoCorrect(newConfig)
}

// ForceRegisterMeshnet registers the device key for meshnet.
//
// Thread-safe.
func (d *DeviceKeyManagerImpl) ForceRegisterMeshnet() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	var cfg config.Config
	if err := d.configManager.Load(&cfg); err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	newConfig, err := d.registerKey(cfg, d.registerMeshnet)
	if err != nil {
		return fmt.Errorf("registering device key: %w", err)
	}

	if err := d.configManager.SaveWith(keyConfig(newConfig)); err != nil {
		return err
	}

	if !isMeshnetRegistrationInfoCorrect(newConfig) {
		return fmt.Errorf("meshnet registration failure")
	}

	return nil
}

type registerFunc func(deviceKey string,
	distroName string,
	isKeyNew bool,
	cfg config.Config) (config.Config, error)

// CheckAndRegisterDedicatedServers checks if device key is registered for the dedicated servers and registers it if it
// isn't. Return true if the key was successfully registered.
//
// Thread-safe.
func (d *DeviceKeyManagerImpl) CheckAndRegisterDedicatedServers() bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	var cfg config.Config
	if err := d.configManager.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return false
	}

	if isDedicatedServersRegistrationInfoCorrect(cfg) {
		return true
	}

	newConfig, err := d.registerKey(cfg, d.registerDedicateServer)
	if err != nil {
		log.Println(internal.ErrorPrefix, "failed to register device key for dedicated servers:", err)
		return false
	}

	if err := d.configManager.SaveWith(keyConfig(newConfig)); err != nil {
		log.Println(internal.ErrorPrefix, "failed to save dedicated servers config:", err)
		return false
	}

	return isDedicatedServersRegistrationInfoCorrect(newConfig)
}

// registerKey reads the device key from local config(or generates it if it doesn't exist) and registers it using the
// provided registerFunc.
func (d *DeviceKeyManagerImpl) registerKey(cfg config.Config,
	registerFunc registerFunc) (config.Config, error) {
	deviceKey, newKey := d.getDeviceKey(cfg)
	if newKey {
		cfg = invalidateKeyData(cfg)
	}

	distroName, err := sysinfo.GetHostOSName()
	if err != nil {
		return cfg, err
	}

	cfg, err = registerFunc(deviceKey, distroName, newKey, cfg)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

func (d *DeviceKeyManagerImpl) registerMeshnet(deviceKey string,
	distroName string,
	isKeyNew bool,
	cfg config.Config) (config.Config, error) {
	token := cfg.TokensData[cfg.AutoConnectData.ID].Token
	peer, err := d.meshnetRegistery.Register(token, cmesh.Machine{
		HardwareID:      cfg.MachineID,
		PublicKey:       d.keyGenerator.Public(deviceKey),
		OS:              cmesh.OperatingSystem{Name: "linux", Distro: distroName},
		SupportsRouting: true,
	})
	if errors.Is(err, core.ErrConflict) {
		// We try to keep the same keys as long as possible, but if relogin with different account happens
		// then they have to be regenerated. There's no way to check if the current mesh device data
		// belongs to this account or not, so handling this on registering error is the best approach.
		deviceKey = d.keyGenerator.Private()
		cfg = invalidateKeyData(cfg)
		isKeyNew = true

		token := cfg.TokensData[cfg.AutoConnectData.ID].Token
		peer, err = d.meshnetRegistery.Register(token, cmesh.Machine{
			HardwareID:      cfg.MachineID,
			PublicKey:       d.keyGenerator.Public(deviceKey),
			OS:              cmesh.OperatingSystem{Name: "linux", Distro: distroName},
			SupportsRouting: true,
		})
	}
	if err != nil {
		return cfg, fmt.Errorf("registering meshnet: %w", err)
	}

	cfg.MeshDevice = peer
	cfg.DeviceKey = deviceKey

	if isKeyNew {
		// There is a delay in the backend between registering a new key and when that key is recognized, so we need to wait
		// some time, otherwise connection will fail.
		const delayAfterNewKey time.Duration = time.Second * 5
		d.delayFunc(delayAfterNewKey)
	}

	return cfg, nil
}

func (d *DeviceKeyManagerImpl) registerDedicateServer(deviceKey string,
	distroName string,
	isKeyNew bool,
	cfg config.Config) (config.Config, error) {
	if isKeyNew {
		cfg = invalidateKeyData(cfg)
	}

	resp, err := d.dedicatedServersAPI.RegisterDevice(core.DevicesRequest{
		HardwareIdentifier: cfg.MachineID.String(),
		PublicKey:          d.keyGenerator.Public(deviceKey),
		Os:                 "linux",
		Type:               "pc",
		Name:               fmt.Sprintf("Linux %s", distroName),
	})
	if errors.Is(err, core.ErrConflict) {
		// We try to keep the same keys as long as possible, but if relogin with different account happens
		// then they have to be regenerated. There's no way to check if the current mesh device data
		// belongs to this account or not, so handling this on registering error is the best approach.
		deviceKey = d.keyGenerator.Private()
		cfg = invalidateKeyData(cfg)

		uuid, uuidParseErr := uuid.Parse(resp.UUID)
		if uuidParseErr != nil {
			return cfg, fmt.Errorf("parsing UUID: %w", err)
		}

		resp, err = d.dedicatedServersAPI.UpdateDevice(uuid, core.UpdateDeviceRequest{
			PublicKey: deviceKey,
			Name:      fmt.Sprintf("Linux %s", distroName)})
	}

	if err != nil {
		return cfg, fmt.Errorf("registering device in the backend: %w", err)
	}

	uuid, err := uuid.Parse(resp.UUID)
	if err != nil {
		return cfg, fmt.Errorf("parsing UUID: %w", err)
	}

	cfg.DeviceUUID = uuid
	cfg.DeviceKey = deviceKey
	return cfg, nil
}

// keyConfig returns a config SaveFunc that saves device key, meshnet device data and device UUID.
func keyConfig(cfg config.Config) config.SaveFunc {
	return func(c config.Config) config.Config {
		c.DeviceKey = cfg.DeviceKey
		c.MeshDevice = cfg.MeshDevice
		c.DeviceUUID = cfg.DeviceUUID
		return c
	}
}
