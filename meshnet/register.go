package meshnet

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	cmesh "github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/distro"
	"github.com/NordSecurity/nordvpn-linux/internal"

	"github.com/google/uuid"
)

// DelayFunc blocks the app for a duration of time
type DelayFunc func(duration time.Duration)

// Checker provides information about meshnet.
type Checker interface {
	// IsRegistrationInfoCorrect returns true when device has been registered to meshnet.
	IsRegistrationInfoCorrect() bool
	// Register the device
	Register() error
}

// RegisteringChecker does both registration checks and registration, if it's not done.
type RegisteringChecker struct {
	cm        config.Manager
	gen       KeyGenerator
	reg       cmesh.Registry
	mu        sync.Mutex
	delayFunc DelayFunc
}

// NewRegisteringChecker is a default constructor for RegisteringChecker.
func NewRegisteringChecker(
	cm config.Manager,
	gen KeyGenerator,
	reg cmesh.Registry,
) *RegisteringChecker {
	return &RegisteringChecker{cm: cm, gen: gen, reg: reg, delayFunc: time.Sleep}
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
	newKey := false

	privateKey := cfg.MeshPrivateKey
	if privateKey == "" {
		newKey = true
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
	if errors.Is(err, core.ErrConflict) {
		// We try to keep the same keys as long as possible, but if relogin with different account happens
		// then they have to be regenerated. There's no way to check if the current mesh device data
		// belongs to this account or not, so handling this on registering error is the best approach.
		privateKey = r.gen.Private()
		peer, err = r.reg.Register(token, cmesh.Machine{
			HardwareID:      cfg.MachineID,
			PublicKey:       r.gen.Public(privateKey),
			OS:              cmesh.OperatingSystem{Name: "linux", Distro: distroName},
			SupportsRouting: true,
		})
		newKey = true
	}
	if err != nil {
		return err
	}

	if newKey {
		// There is a delay in the backend between registering a new key and when that key is recognized, so we need to wait
		// some time, otherwise connection will fail.
		const delayAfterNewKey time.Duration = time.Second * 5
		r.delayFunc(delayAfterNewKey)
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
