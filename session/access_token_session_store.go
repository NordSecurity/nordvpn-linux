package session

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/google/uuid"
)

type AccessTokenResponse struct {
	Token      string
	RenewToken string
	ExpiresAt  string
}

type AccessTokenRenewalAPICall func(token string, idempotencyKey uuid.UUID) (*AccessTokenResponse, error)

type AccessTokenExternalValidator func(token string) error

type accessTokenConfig struct {
	Token      string
	RenewToken string
	ExpiresAt  time.Time
}

type AccessTokenSessionStore struct {
	cfgManager         config.Manager
	errHandlerRegistry *internal.ErrorHandlingRegistry[error]
	renewAPICall       AccessTokenRenewalAPICall
	externalValidator  AccessTokenExternalValidator
}

func NewAccessTokenSessionStore(
	cfgManager config.Manager,
	errorHandlingRegistry *internal.ErrorHandlingRegistry[error],
	renewAPICall AccessTokenRenewalAPICall,
	externalValidator AccessTokenExternalValidator,
) *AccessTokenSessionStore {
	return &AccessTokenSessionStore{
		cfgManager:         cfgManager,
		errHandlerRegistry: errorHandlingRegistry,
		renewAPICall:       renewAPICall,
		externalValidator:  externalValidator,
	}
}

func (s *AccessTokenSessionStore) Renew() error {
	cfg, err := s.getConfig()
	if err != nil {
		return err
	}

	// Handle manual access tokens
	if cfg.ExpiresAt.Equal(ManualAccessTokenExpiryDate) {
		if err := s.Validate(cfg); err != nil {
			_ = s.Invalidate(err)
			return ErrAccessTokenRevoked
		}
		return nil
	}

	// Check if token needs renewal
	if err := s.Validate(cfg); err == nil {
		return nil
	}

	// Token is expired, proceed with renewal
	var fullCfg config.Config
	if err := s.cfgManager.Load(&fullCfg); err != nil {
		return err
	}

	uid := fullCfg.AutoConnectData.ID
	data, ok := fullCfg.TokensData[uid]
	if !ok {
		return errors.New("no token data")
	}

	if err := s.renewToken(uid, data); err != nil {
		log.Printf("[auth] %s Renewing token for uid(%v): %s\n", internal.ErrorPrefix, uid, err)
		return err
	}

	return nil
}

// Validate checks if the access token is valid
func (s *AccessTokenSessionStore) Validate(cfg accessTokenConfig) error {
	if cfg.ExpiresAt.Equal(ManualAccessTokenExpiryDate) {
		if !internal.AccessTokenFormatValidator(cfg.Token) {
			return fmt.Errorf("invalid access token format: %w", ErrAccessTokenExpired)
		}

		if s.externalValidator != nil {
			return s.externalValidator(cfg.Token)
		}

		return nil
	}

	return ValidateExpiry(cfg.ExpiresAt)
}

func (s *AccessTokenSessionStore) Invalidate(reason error) error {
	handlers := s.errHandlerRegistry.GetHandlers(reason)
	if len(handlers) == 0 {
		return fmt.Errorf("invalidating session: %w", reason)
	}

	for _, handler := range handlers {
		handler(reason)
	}
	return nil
}

func (s *AccessTokenSessionStore) renewToken(uid int64, data config.TokenData) error {
	if s.renewAPICall == nil {
		return errors.New("renewal API call not configured")
	}

	if err := s.tryUpdateIdempotencyKey(uid, &data); err != nil {
		return err
	}

	resp, err := s.renewAPICall(data.Token, *data.IdempotencyKey)
	if err != nil {
		return s.Invalidate(err)
	}

	if resp == nil {
		return errors.New("renewal API returned nil response")
	}

	expTime, err := time.Parse(internal.ServerDateFormat, resp.ExpiresAt)
	if err != nil {
		return fmt.Errorf("parsing expiry time: %w", err)
	}

	// Use config's SaveWith for atomic bulk update
	err = s.cfgManager.SaveWith(func(c config.Config) config.Config {
		data := c.TokensData[uid]
		data.Token = resp.Token
		data.RenewToken = resp.RenewToken
		data.TokenExpiry = expTime.Format(internal.ServerDateFormat)
		c.TokensData[uid] = data
		return c
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *AccessTokenSessionStore) tryUpdateIdempotencyKey(uid int64, data *config.TokenData) error {
	if data.IdempotencyKey != nil {
		return nil
	}

	key := uuid.New()
	data.IdempotencyKey = &key
	err := s.cfgManager.SaveWith(func(c config.Config) config.Config {
		user := c.TokensData[uid]
		user.IdempotencyKey = data.IdempotencyKey
		c.TokensData[uid] = user
		return c
	})

	if err != nil {
		return fmt.Errorf("saving idempotency key: %w", err)
	}

	return nil
}

func (s *AccessTokenSessionStore) getConfig() (accessTokenConfig, error) {
	var cfg config.Config
	if err := s.cfgManager.Load(&cfg); err != nil {
		return accessTokenConfig{}, err
	}

	data, ok := cfg.TokensData[cfg.AutoConnectData.ID]
	if !ok {
		return accessTokenConfig{}, errors.New("non existing data")
	}

	expiryTime, err := time.Parse(internal.ServerDateFormat, data.TokenExpiry)
	if err != nil {
		expiryTime = time.Now().Add(-1 * time.Second)
	}

	return accessTokenConfig{
		Token:      data.Token,
		RenewToken: data.RenewToken,
		ExpiresAt:  expiryTime,
	}, nil
}

func (s *AccessTokenSessionStore) setConfig(cfg accessTokenConfig) error {
	err := s.cfgManager.SaveWith(func(c config.Config) config.Config {
		data := c.TokensData[c.AutoConnectData.ID]
		data.Token = cfg.Token
		data.RenewToken = cfg.RenewToken
		data.TokenExpiry = cfg.ExpiresAt.Format(internal.ServerDateFormat)
		c.TokensData[c.AutoConnectData.ID] = data
		return c
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *AccessTokenSessionStore) reset() {
	s.cfgManager.SaveWith(func(c config.Config) config.Config {
		data := c.TokensData[c.AutoConnectData.ID]
		data.Token = ""
		data.RenewToken = ""
		data.TokenExpiry = ""
		c.TokensData[c.AutoConnectData.ID] = data
		return c
	})
}

func (s *AccessTokenSessionStore) SetToken(value string) error {
	cfg, err := s.getConfig()
	if err != nil {
		return err
	}
	cfg.Token = value
	return s.setConfig(cfg)
}

func (s *AccessTokenSessionStore) SetRenewToken(value string) error {
	cfg, err := s.getConfig()
	if err != nil {
		return err
	}
	cfg.RenewToken = value
	return s.setConfig(cfg)
}

func (s *AccessTokenSessionStore) SetExpiry(value time.Time) error {
	cfg, err := s.getConfig()
	if err != nil {
		return err
	}
	cfg.ExpiresAt = value
	return s.setConfig(cfg)
}

func (s *AccessTokenSessionStore) GetToken() string {
	cfg, err := s.getConfig()
	if err != nil {
		return ""
	}
	return cfg.Token
}

func (s *AccessTokenSessionStore) GetRenewalToken() string {
	cfg, err := s.getConfig()
	if err != nil {
		return ""
	}
	return cfg.RenewToken
}

func (s *AccessTokenSessionStore) GetExpiry() time.Time {
	cfg, err := s.getConfig()
	if err != nil {
		return time.Time{}
	}
	return cfg.ExpiresAt
}

func (s *AccessTokenSessionStore) IsExpired() bool {
	cfg, err := s.getConfig()
	if err != nil {
		return true
	}
	return time.Now().After(cfg.ExpiresAt)
}
