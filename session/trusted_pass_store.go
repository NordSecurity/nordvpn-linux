package session

import (
	"errors"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

type trustedPassConfig struct {
	Token     string
	OwnerID   string
	ExpiresAt time.Time
}

type trustedPassSession struct {
	cm config.Manager
}

// TODO: needs global caching
func (s *trustedPassSession) get() (trustedPassConfig, error) {
	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		return trustedPassConfig{}, err
	}

	data, ok := cfg.TokensData[cfg.AutoConnectData.ID]
	if !ok {
		return trustedPassConfig{}, errors.New("non existing data")
	}

	expiryTime, err := time.Parse(internal.ServerDateFormat, data.TrustedPassTokenExpiry)
	if err != nil {
		// if we don't have valid time, just make it already expired
		expiryTime = time.Now().Add(-1 * time.Second)
	}

	return trustedPassConfig{
		Token:     data.TrustedPassToken,
		OwnerID:   data.TrustedPassOwnerID,
		ExpiresAt: expiryTime,
	}, nil
}

// TODO: needs global caching
func (s *trustedPassSession) set(cfg trustedPassConfig) error {
	err := s.cm.SaveWith(func(c config.Config) config.Config {
		data := c.TokensData[c.AutoConnectData.ID]
		data.TrustedPassToken = cfg.Token
		data.TrustedPassOwnerID = cfg.OwnerID
		data.TrustedPassTokenExpiry = cfg.ExpiresAt.Format(internal.ServerDateFormat)
		c.TokensData[c.AutoConnectData.ID] = data
		return c
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *TrustedPassSessionStore) SetToken(value string) error {
	cfg, err := s.session.get()
	if err != nil {
		return err
	}
	cfg.Token = value
	return s.session.set(cfg)
}

func (s *TrustedPassSessionStore) SetOwnerID(value string) error {
	cfg, err := s.session.get()
	if err != nil {
		return err
	}
	cfg.OwnerID = value
	return s.session.set(cfg)
}

func (s *TrustedPassSessionStore) SetExpiry(value time.Time) error {
	cfg, err := s.session.get()
	if err != nil {
		return err
	}
	cfg.ExpiresAt = value
	return s.session.set(cfg)
}

// implements SessionTokenProvider
func (s *TrustedPassSessionStore) GetToken() string {
	cfg, err := s.session.get()
	if err != nil {
		return ""
	}
	return cfg.Token
}

// implements SessionOwnerProvider
func (s *TrustedPassSessionStore) GetOwnerID() string {
	cfg, err := s.session.get()
	if err != nil {
		return ""
	}
	return cfg.OwnerID
}

// implements SessionExpiryProvider
func (s *TrustedPassSessionStore) GetExpiry() time.Time {
	cfg, err := s.session.get()
	if err != nil {
		return time.Time{}
	}
	return cfg.ExpiresAt
}

// implements ExpirableSession
func (s *TrustedPassSessionStore) IsExpired() bool {
	cfg, err := s.session.get()
	if err != nil {
		return true
	}
	return time.Now().After(cfg.ExpiresAt)
}

func newTrustedPassSession(confman config.Manager) *trustedPassSession {
	return &trustedPassSession{cm: confman}
}
