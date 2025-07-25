package session

import (
	"errors"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

type accessTokenConfig struct {
	Token      string
	RenewToken string
	ExpiresAt  time.Time
}

type accessTokenSession struct {
	cm config.Manager
}

// TODO: needs global caching
func (s *accessTokenSession) get() (accessTokenConfig, error) {
	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		return accessTokenConfig{}, err
	}

	data, ok := cfg.TokensData[cfg.AutoConnectData.ID]
	if !ok {
		return accessTokenConfig{}, errors.New("non existing data")
	}

	// must contain valid data
	expiryTime, _ := time.Parse(internal.ServerDateFormat, data.TokenExpiry)

	return accessTokenConfig{
		Token:      data.Token,
		RenewToken: data.RenewToken,
		ExpiresAt:  expiryTime,
	}, nil
}

// TODO: needs global caching
func (s *accessTokenSession) set(cfg accessTokenConfig) error {
	err := s.cm.SaveWith(func(c config.Config) config.Config {
		data := c.TokensData[c.AutoConnectData.ID]
		data.Token = cfg.Token
		data.RenewToken = cfg.RenewToken
		data.TokenExpiry = cfg.ExpiresAt.String()
		c.TokensData[c.AutoConnectData.ID] = data
		return c
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *AccessTokenSessionStore) SetToken(value string) error {
	cfg, err := s.session.get()
	if err != nil {
		return err
	}
	cfg.Token = value
	return s.session.set(cfg)
}

func (s *AccessTokenSessionStore) SetRenewToken(value string) error {
	cfg, err := s.session.get()
	if err != nil {
		return err
	}
	cfg.RenewToken = value
	return s.session.set(cfg)
}

func (s *AccessTokenSessionStore) SetExpiry(value time.Time) error {
	cfg, err := s.session.get()
	if err != nil {
		return err
	}
	cfg.ExpiresAt = value
	return s.session.set(cfg)
}

// implements SessionTokenProvider
func (s *AccessTokenSessionStore) GetToken() string {
	cfg, err := s.session.get()
	if err != nil {
		return ""
	}
	return cfg.Token
}

// implements SessionRenewalTokenProvider
func (s *AccessTokenSessionStore) GetRenewalToken() string {
	cfg, err := s.session.get()
	if err != nil {
		return ""
	}
	return cfg.RenewToken
}

// implements SessionExpiryProvider
func (s *AccessTokenSessionStore) GetExpiry() time.Time {
	cfg, err := s.session.get()
	if err != nil {
		return time.Time{}
	}
	return cfg.ExpiresAt
}

// implements ExpirableSession
func (s *AccessTokenSessionStore) IsExpired() bool {
	cfg, err := s.session.get()
	if err != nil {
		return true
	}
	return time.Now().After(cfg.ExpiresAt)
}

func newAccessTokenSession(confman config.Manager) *accessTokenSession {
	return &accessTokenSession{cm: confman}
}
