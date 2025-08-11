package mock

import (
	"encoding/json"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
)

type ConfigManager struct {
	mu      sync.RWMutex
	Cfg     *config.Config
	SaveErr error
	LoadErr error
	Saved   bool
}

func NewMockConfigManager() *ConfigManager {
	m := ConfigManager{}
	m.Cfg = &config.Config{
		TokensData: map[int64]config.TokenData{
			0: {
				Token:              "",
				TokenExpiry:        "",
				RenewToken:         "",
				NordLynxPrivateKey: "",
				OpenVPNUsername:    "",
				OpenVPNPassword:    "",
			},
		},
		AutoConnectData: config.AutoConnectData{
			Allowlist: config.Allowlist{
				Ports: config.Ports{
					TCP: config.PortSet{},
					UDP: config.PortSet{},
				},
			},
		},
	}
	m.Cfg.MeshDevice = &mesh.Machine{}

	return &m
}

// makeDeepCopy creates a complete deep copy of the config using JSON marshaling
// to prevent race conditions when accessing maps and pointers.
func makeDeepCopy(src config.Config) (config.Config, error) {
	var dst config.Config

	// Marshal the source config to JSON
	data, err := json.Marshal(src)
	if err != nil {
		return dst, err
	}

	// Unmarshal back to create a deep copy
	err = json.Unmarshal(data, &dst)
	if err != nil {
		return dst, err
	}

	return dst, nil
}

func (m *ConfigManager) SaveWith(fn config.SaveFunc) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.SaveErr != nil {
		return m.SaveErr
	}

	if m.Cfg == nil {
		m.Cfg = &config.Config{}
	}

	copyCfg, err := makeDeepCopy(*m.Cfg)
	if err != nil {
		return err
	}

	updatedCfg := fn(copyCfg)
	*m.Cfg = updatedCfg

	m.Saved = true
	return nil
}

func (m *ConfigManager) Load(c *config.Config) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.LoadErr != nil {
		return m.LoadErr
	}

	if m.Cfg == nil {
		m.Cfg = &config.Config{}
	}

	copyCfg, err := makeDeepCopy(*m.Cfg)
	if err != nil {
		return err
	}

	*c = copyCfg
	return nil
}

func (m *ConfigManager) Reset(bool, bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	*m.Cfg = config.Config{}
	return nil
}
