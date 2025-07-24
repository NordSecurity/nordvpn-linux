package session

import (
	"errors"

	"github.com/NordSecurity/nordvpn-linux/config"
)

type vpnCredentialsConfig struct {
	Username           string
	Password           string
	NordLynxPrivateKey string
}

type vpnCredentialsSession struct {
	cm config.Manager
}

// TODO: needs global caching
func (s *vpnCredentialsSession) get() (vpnCredentialsConfig, error) {

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		return vpnCredentialsConfig{}, err
	}

	data, ok := cfg.TokensData[cfg.AutoConnectData.ID]
	if !ok {
		return vpnCredentialsConfig{}, errors.New("non existing data")
	}

	return vpnCredentialsConfig{
		Username:           data.OpenVPNUsername,
		Password:           data.OpenVPNPassword,
		NordLynxPrivateKey: data.NordLynxPrivateKey,
	}, nil
}

// TODO: needs global caching
func (s *vpnCredentialsSession) set(cfg vpnCredentialsConfig) error {
	err := s.cm.SaveWith(func(c config.Config) config.Config {
		data := c.TokensData[c.AutoConnectData.ID]
		data.NordLynxPrivateKey = cfg.NordLynxPrivateKey
		data.OpenVPNUsername = cfg.Username
		data.OpenVPNPassword = cfg.Password
		c.TokensData[c.AutoConnectData.ID] = data
		return c
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *vpnCredentialsSession) reset() {
	s.cm.SaveWith(func(c config.Config) config.Config {
		data := c.TokensData[c.AutoConnectData.ID]
		data.NordLynxPrivateKey = ""
		data.OpenVPNUsername = ""
		data.OpenVPNPassword = ""
		c.TokensData[c.AutoConnectData.ID] = data
		return c
	})
}

// SetUsername
func (s *vpnCredentialsSession) SetUsername(value string) error {
	cfg, err := s.get()
	if err != nil {
		return err
	}
	cfg.Username = value
	return s.set(cfg)
}

// SetPassword
func (s *vpnCredentialsSession) SetPassword(value string) error {
	cfg, err := s.get()
	if err != nil {
		return err
	}
	cfg.Password = value
	return s.set(cfg)
}

// SetNordlynxPrivateKey
func (s *vpnCredentialsSession) SetNordlynxPrivateKey(value string) error {
	cfg, err := s.get()
	if err != nil {
		return err
	}
	cfg.NordLynxPrivateKey = value
	return s.set(cfg)
}

// GetUsername
// implements CredentialsBasedSession
func (s *vpnCredentialsSession) GetUsername() string {
	cfg, err := s.get()
	if err != nil {
		return ""
	}
	return cfg.Username
}

// GetPassword
// implements CredentialsBasedSession
func (s *vpnCredentialsSession) GetPassword() string {
	cfg, err := s.get()
	if err != nil {
		return ""
	}
	return cfg.Password
}

// GetNordlynxPrivateKey
func (s *vpnCredentialsSession) GetNordlynxPrivateKey() string {
	cfg, err := s.get()
	if err != nil {
		return ""
	}
	return cfg.NordLynxPrivateKey
}

// IsExpired
// implements ExpirableSession
func (s *vpnCredentialsSession) IsExpired() bool {
	return false
}

func newVPNCredentialsSession(confman config.Manager) *vpnCredentialsSession {
	return &vpnCredentialsSession{cm: confman}
}
