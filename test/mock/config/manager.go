package config

import "github.com/NordSecurity/nordvpn-linux/config"

type ConfigManagerMock struct {
	Cfg       config.Config
	LoadError error
}

// SaveWith updates parts of the config specified by the SaveFunc.
func (c *ConfigManagerMock) SaveWith(saveFunc config.SaveFunc) error {
	c.Cfg = saveFunc(c.Cfg)

	return nil
}

func (c *ConfigManagerMock) Load(cfg *config.Config) error {
	*cfg = c.Cfg
	return c.LoadError
}

func (*ConfigManagerMock) Reset() error {
	return nil
}
