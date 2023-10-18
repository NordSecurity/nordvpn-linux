package mock

import (
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
)

type ConfigManager struct {
	Cfg     *config.Config
	SaveErr error
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
	return nil
}

func (m *ConfigManager) Load(c *config.Config) error {
	if m.Cfg == nil {
		m.Cfg = &config.Config{}
	}
	if m.Cfg.MeshDevice == nil {
		m.Cfg.MeshDevice = &mesh.Machine{}
	}
	*c = *m.Cfg
	return nil
}

func (m *ConfigManager) Reset() error {
	*m.Cfg = config.Config{}
	return nil
}
