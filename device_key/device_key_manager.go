package devicekey

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	cmesh "github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/log"
	sysinfo "github.com/NordSecurity/nordvpn-linux/sysinfo"

	"github.com/google/uuid"
)

type DeviceKeyInvalidator interface {
	// InvalidateDeviceKeyData invalidates the device key and all of it's associated data
	InvalidateDeviceKeyData() error
}

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
	// CheckAndRegisterMeshnet checks if the device key is registered for meshnet and registers it if it isn't.
	CheckAndRegisterMeshnet() bool
	// ForceRegisterMeshnet registers device key for meshnet.
	ForceRegisterMeshnet() error
}

// DedicatedServersConnectionData contains device side data necessary to connect to a dedicated server
type DedicatedServersConnectionData struct {
	// DevicePublicKey is used when making dedicated server connect checks
	DevicePublicKey string
	// DevicePrivateKey is used when connecting to a dedicated server
	DevicePrivateKey string
	// DeviceUUID is used when making dedicated server connect checks
	DeviceUUID uuid.UUID
}

type DedicatedServersKeyManager interface {
	// CheckAndRegisterDedicatedServers checks if device has been registered for private servers and registers it if it
	// isn't.
	// Returns the connection data if it is available. Returns nil if data is not available.
	CheckAndRegisterDedicatedServers() *DedicatedServersConnectionData
	ForceRegisterDedicatedServers() *DedicatedServersConnectionData
	DeviceKeyInvalidator
}

// DeviceKeyManagerImpl does both registration checks and registration, if it's not done.
type DeviceKeyManagerImpl struct {
	configManager       config.Manager
	keyGenerator        KeyGenerator
	meshnetRegistry     cmesh.Registry
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
		meshnetRegistry:     meshnetRegistery,
		dedicatedServersAPI: dedicatedServersAPI,
		delayFunc:           time.Sleep}
}

func isMeshnetRegistrationInfoCorrect(cfg config.Config) error {
	if cfg.DeviceKey == "" {
		return fmt.Errorf("device key is missing")
	}

	if cfg.MeshDevice == nil {
		return fmt.Errorf("meshnet device data missing")
	}

	if cfg.MeshDevice.ID == uuid.Nil {
		return fmt.Errorf("mesh device ID is missing")
	}

	if !cfg.MeshDevice.Address.IsValid() {
		return fmt.Errorf("mesh device address is invalid")
	}

	return nil
}

func isDedicatedServersRegistrationInfoCorrect(cfg config.Config) bool {
	return cfg.DeviceKey != "" && cfg.DeviceUUID != uuid.Nil
}

func invalidateKeyData(cfg *config.Config) *config.Config {
	cfg.MeshDevice = nil
	cfg.DeviceUUID = uuid.Nil
	return cfg
}

// getDeviceKey returns device key from cfg or generates a new key if it doesn't exist. Returns a tuple where first
// element is the device key and second element is a bool that holds true if new key got generated and false if old key
// was read from the cfg.
func (d *DeviceKeyManagerImpl) getDeviceKey(cfg config.Config) (string, bool) {
	if cfg.DeviceKey == "" {
		return d.keyGenerator.Private(), true
	}

	return cfg.DeviceKey, false
}

// CheckAndRegisterMeshnet registers the device key for meshnet if it isn't registered and returns true if it is successfully
// registered.
//
// Thread-safe.
func (d *DeviceKeyManagerImpl) CheckAndRegisterMeshnet() bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	var cfg config.Config
	if err := d.configManager.Load(&cfg); err != nil {
		log.Error(err)
		return false
	}

	if err := isMeshnetRegistrationInfoCorrect(cfg); err == nil {
		return true
	}

	newConfig, err := d.registerKey(&cfg, d.registerMeshnet)
	if err != nil {
		log.Error("failed to register new device key:", err)
		return false
	}

	if err := d.configManager.SaveWith(keyConfig(*newConfig)); err != nil {
		log.Error(err)
		return false
	}

	if err := isMeshnetRegistrationInfoCorrect(*newConfig); err != nil {
		log.Error("registration failed:", err)
		return false
	}
	return true
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

	newConfig, err := d.registerKey(&cfg, d.registerMeshnet)
	if err != nil {
		return fmt.Errorf("registering device key: %w", err)
	}

	if err := d.configManager.SaveWith(keyConfig(*newConfig)); err != nil {
		return err
	}

	return isMeshnetRegistrationInfoCorrect(*newConfig)
}

type registerFunc func(deviceKey string,
	distroName string,
	isKeyNew bool,
	cfg *config.Config) (*config.Config, error)

func (d *DeviceKeyManagerImpl) registerDedicatedServer(force bool) *DedicatedServersConnectionData {
	var cfg config.Config
	if err := d.configManager.Load(&cfg); err != nil {
		log.Error(err)
		return nil
	}

	if !force {
		if isDedicatedServersRegistrationInfoCorrect(cfg) {
			return &DedicatedServersConnectionData{
				DeviceUUID:       cfg.DeviceUUID,
				DevicePublicKey:  d.keyGenerator.Public(cfg.DeviceKey),
				DevicePrivateKey: cfg.DeviceKey,
			}
		}
	}

	newConfig, err := d.registerKey(&cfg, d.registerDedicatedServerKey)
	if err != nil {
		log.Error("failed to register device key for dedicated servers:", err)
		return nil
	}

	if err := d.configManager.SaveWith(keyConfig(*newConfig)); err != nil {
		log.Error("failed to save dedicated servers config:", err)
		return nil
	}

	if !isDedicatedServersRegistrationInfoCorrect(*newConfig) {
		return nil
	}

	return &DedicatedServersConnectionData{
		DeviceUUID:       newConfig.DeviceUUID,
		DevicePublicKey:  d.keyGenerator.Public(newConfig.DeviceKey),
		DevicePrivateKey: newConfig.DeviceKey,
	}
}

// CheckAndRegisterDedicatedServers checks if device key is registered for the dedicated servers and registers it if not.
// Returns nil if key was not successfully registered.
//
// Thread-safe.
func (d *DeviceKeyManagerImpl) CheckAndRegisterDedicatedServers() *DedicatedServersConnectionData {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.registerDedicatedServer(false)
}

// ForceRegisterDedicatedServers registers device key for dedicated servers ignoring existing registration data.
// Returns nil if key was not successfully registered.
//
// Thread-safe.
func (d *DeviceKeyManagerImpl) ForceRegisterDedicatedServers() *DedicatedServersConnectionData {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.registerDedicatedServer(true)
}

func (d *DeviceKeyManagerImpl) InvalidateDeviceKeyData() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	err := d.configManager.SaveWith(func(c config.Config) config.Config {
		c.DeviceKey = ""
		c = *invalidateKeyData(&c)
		return c
	})
	if err != nil {
		return fmt.Errorf("invalidating key data: %w", err)
	}

	return nil
}

// registerKey reads the device key from local config(or generates it if it doesn't exist) and registers it using the
// provided registerFunc.
func (d *DeviceKeyManagerImpl) registerKey(cfg *config.Config,
	registerFunc registerFunc) (*config.Config, error) {
	deviceKey, newKey := d.getDeviceKey(*cfg)
	if newKey {
		cfg = invalidateKeyData(cfg)
	}

	distroName, err := sysinfo.GetHostOSName()
	if err != nil {
		return nil, err
	}

	cfg, err = registerFunc(deviceKey, distroName, newKey, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (d *DeviceKeyManagerImpl) registerMeshnet(deviceKey string,
	distroName string,
	isKeyNew bool,
	cfg *config.Config) (*config.Config, error) {
	token := cfg.TokensData[cfg.AutoConnectData.ID].Token
	peer, err := d.meshnetRegistry.Register(token, cmesh.Machine{
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
		peer, err = d.meshnetRegistry.Register(token, cmesh.Machine{
			HardwareID:      cfg.MachineID,
			PublicKey:       d.keyGenerator.Public(deviceKey),
			OS:              cmesh.OperatingSystem{Name: "linux", Distro: distroName},
			SupportsRouting: true,
		})
	}
	if err != nil {
		return nil, fmt.Errorf("registering meshnet: %w", err)
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

func (d *DeviceKeyManagerImpl) registerDedicatedServerKey(deviceKey string,
	distroName string,
	_ bool,
	cfg *config.Config) (*config.Config, error) {
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
		// If meshnet is on we can reuse it's key. If it's off we invalidate all of the key data.
		if !cfg.Mesh {
			deviceKey = d.keyGenerator.Private()
			cfg = invalidateKeyData(cfg)
		}

		uuid, uuidParseErr := uuid.Parse(resp.UUID)
		if uuidParseErr != nil {
			return nil, fmt.Errorf("parsing UUID: %w", uuidParseErr)
		}

		resp, err = d.dedicatedServersAPI.UpdateDevice(uuid, core.UpdateDeviceRequest{
			PublicKey: d.keyGenerator.Public(deviceKey),
			Name:      fmt.Sprintf("Linux %s", distroName)})
	}

	if err != nil {
		return nil, fmt.Errorf("registering device in the backend: %w", err)
	}

	uuid, err := uuid.Parse(resp.UUID)
	if err != nil {
		return nil, fmt.Errorf("parsing UUID: %w", err)
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
