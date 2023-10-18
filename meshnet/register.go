package meshnet

import (
	"fmt"
	"log"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/config"
	cmesh "github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/distro"
	"github.com/NordSecurity/nordvpn-linux/internal"

	"github.com/google/uuid"
)

// Checker provides information about meshnet.
type Checker interface {
	// IsRegistrationInfoCorrect returns true when device has been registered to meshnet.
	IsRegistrationInfoCorrect() bool
	// Register the device
	Register() error
}

// RegisteringChecker does both registration checks and registration, if it's not done.
type RegisteringChecker struct {
	cm  config.Manager
	gen KeyGenerator
	reg cmesh.Registry
	mu  sync.Mutex
}

// NewRegisteringChecker is a default constructor for RegisteringChecker.
func NewRegisteringChecker(
	cm config.Manager,
	gen KeyGenerator,
	reg cmesh.Registry,
) *RegisteringChecker {
	return &RegisteringChecker{cm: cm, gen: gen, reg: reg}
}

func isRegistrationInfoCorrect(cfg config.Config) bool {
	return cfg.MeshDevice != nil &&
		cfg.MeshPrivateKey != "" &&
		cfg.MeshDevice.ID != uuid.Nil &&
		cfg.MeshDevice.Address.IsValid()
}

// IsRegistrationInfoCorrect reports meshnet device registration status.
//
// Thread-safe.
func (r *RegisteringChecker) IsRegistrationInfoCorrect() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return false
	}

	if isRegistrationInfoCorrect(cfg) {
		return true
	}

	if err := r.register(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return false
	}

	if err := r.cm.SaveWith(meshConfig(cfg.MeshDevice, cfg.MeshPrivateKey)); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return false
	}

	return isRegistrationInfoCorrect(cfg)
}

// Register registers the device in API, even if it was already registered
func (r *RegisteringChecker) Register() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		return err
	}

	if err := r.register(&cfg); err != nil {
		return err
	}

	if err := r.cm.SaveWith(meshConfig(cfg.MeshDevice, cfg.MeshPrivateKey)); err != nil {
		return err
	}

	if !isRegistrationInfoCorrect(cfg) {
		return fmt.Errorf("meshnet registration failure")
	}

	return nil
}

func (r *RegisteringChecker) register(cfg *config.Config) error {
	privateKey := cfg.MeshPrivateKey
	if privateKey == "" {
		privateKey = r.gen.Private()
	}
	token := cfg.TokensData[cfg.AutoConnectData.ID].Token
	distroName, err := distro.ReleaseName()
	if err != nil {
		return err
	}

	peer, err := r.reg.Register(token, cmesh.Machine{
		HardwareID:      cfg.MachineID,
		PublicKey:       r.gen.Public(privateKey),
		OS:              cmesh.OperatingSystem{Name: "linux", Distro: distroName},
		SupportsRouting: true,
	})
	if err != nil {
		return err
	}

	cfg.MeshDevice = peer
	cfg.MeshPrivateKey = privateKey
	return nil
}

func meshConfig(peer *cmesh.Machine, key string) config.SaveFunc {
	return func(c config.Config) config.Config {
		c.MeshDevice = peer
		c.MeshPrivateKey = key
		return c
	}
}
