package session

import (
	"errors"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
)

type ncCredentialsConfig struct {
	Username  string
	Password  string
	Endpoint  string
	ExpiresAt time.Time
}

type ncCredentialsSession struct {
	cm config.Manager
}

// TODO: needs global caching
func (s *ncCredentialsSession) get() (ncCredentialsConfig, error) {

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		return ncCredentialsConfig{}, err
	}

	data, ok := cfg.TokensData[cfg.AutoConnectData.ID]
	if !ok {
		return ncCredentialsConfig{}, errors.New("non existing data")
	}

	return ncCredentialsConfig{
		Username:  data.NCData.Username,
		Password:  data.NCData.Password,
		Endpoint:  data.NCData.Endpoint,
		ExpiresAt: data.NCData.ExpirationDate,
	}, nil
}

// TODO: needs global caching
func (s *ncCredentialsSession) set(cfg ncCredentialsConfig) error {
	err := s.cm.SaveWith(func(c config.Config) config.Config {
		data := c.TokensData[c.AutoConnectData.ID]
		data.NCData.Username = cfg.Username
		data.NCData.Password = cfg.Password
		data.NCData.Endpoint = cfg.Endpoint
		data.NCData.ExpirationDate = cfg.ExpiresAt
		c.TokensData[c.AutoConnectData.ID] = data
		return c
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *ncCredentialsSession) reset() {
	s.cm.SaveWith(func(c config.Config) config.Config {
		data := c.TokensData[c.AutoConnectData.ID]
		data.NCData.Username = ""
		data.NCData.Password = ""
		data.NCData.Endpoint = ""
		data.NCData.ExpirationDate = time.Time{}
		c.TokensData[c.AutoConnectData.ID] = data
		return c
	})
}

// SetUsername
func (s *ncCredentialsSession) SetUsername(value string) error {
	cfg, err := s.get()
	if err != nil {
		return err
	}
	cfg.Username = value
	return s.set(cfg)
}

// SetPassword
func (s *ncCredentialsSession) SetPassword(value string) error {
	cfg, err := s.get()
	if err != nil {
		return err
	}
	cfg.Password = value
	return s.set(cfg)
}

// SetEndpoint
func (s *ncCredentialsSession) SetEndpoint(value string) error {
	cfg, err := s.get()
	if err != nil {
		return err
	}
	cfg.Endpoint = value
	return s.set(cfg)
}

// SetExpiry
func (s *ncCredentialsSession) SetExpiry(value time.Time) error {
	cfg, err := s.get()
	if err != nil {
		return err
	}
	cfg.ExpiresAt = value
	return s.set(cfg)
}

// GetUsername
// implements CredentialsBasedSession
func (s *ncCredentialsSession) GetUsername() string {
	cfg, err := s.get()
	if err != nil {
		return ""
	}
	return cfg.Username
}

// GetPassword
// implements CredentialsBasedSession
func (s *ncCredentialsSession) GetPassword() string {
	cfg, err := s.get()
	if err != nil {
		return ""
	}
	return cfg.Password
}

// GetEndpoint
func (s *ncCredentialsSession) GetEndpoint() string {
	cfg, err := s.get()
	if err != nil {
		return ""
	}
	return cfg.Endpoint
}

// GetExpiry
// implements SessionExpiryProvider
func (s *ncCredentialsSession) GetExpiry() time.Time {
	cfg, err := s.get()
	if err != nil {
		return time.Time{}
	}
	return cfg.ExpiresAt
}

// IsExpired
// implements ExpirableSession
func (s *ncCredentialsSession) IsExpired() bool {
	cfg, err := s.get()
	if err != nil {
		return true
	}
	return time.Now().After(cfg.ExpiresAt)
}

func newNCCredentialsSession(confman config.Manager) *ncCredentialsSession {
	return &ncCredentialsSession{cm: confman}
}
