package mock

import (
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
)

type ConfigManager struct {
	Cfg     *config.Config
	SaveErr error
	LoadErr error
	Saved   bool
}

func NewMockConfigManager() *ConfigManager {
	m := ConfigManager{}
	m.Cfg = &config.Config{}
	m.Cfg.MeshDevice = &mesh.Machine{}

	return &m
}

func (m *ConfigManager) SaveWith(fn config.SaveFunc) error {
	if m.SaveErr != nil {
		return m.SaveErr
	}

	if m.Cfg == nil {
		m.Cfg = &config.Config{}
	}
	cfg := fn(*m.Cfg)
	*m.Cfg = cfg
	m.Saved = true
	return nil
}

func (m *ConfigManager) Load(c *config.Config) error {
	if m.LoadErr != nil {
		return m.LoadErr
	}
	if m.Cfg == nil {
		m.Cfg = &config.Config{}
	}

	*c = *m.Cfg
	return nil
}

func (m *ConfigManager) Reset(bool) error {
	*m.Cfg = config.Config{}
	return nil
}
