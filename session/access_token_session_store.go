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

// AccessTokenResponse represents the response from the access token renewal API
type AccessTokenResponse struct {
	Token      string
	RenewToken string
	ExpiresAt  string
}

// AccessTokenRenewalAPICall is a function type for renewing access tokens
type AccessTokenRenewalAPICall func(token string, idempotencyKey uuid.UUID) (*AccessTokenResponse, error)

// AccessTokenExternalValidator is a function type for external validation of access tokens
type AccessTokenExternalValidator func(token string) error

type accessTokenConfig struct {
	Token      string
	RenewToken string
	ExpiresAt  time.Time
}

// AccessTokenSessionStore manages access token-based sessions
type AccessTokenSessionStore struct {
	cfgManager         config.Manager
	errHandlerRegistry *internal.ErrorHandlingRegistry[error]
	renewAPICall       AccessTokenRenewalAPICall
	externalValidator  AccessTokenExternalValidator
}

// NewAccessTokenSessionStore creates a new AccessTokenSessionStore instance
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

// Renew checks if the access token needs renewal and renews it if necessary
func (s *AccessTokenSessionStore) Renew() error {
	// Check if token needs renewal
	if err := s.Validate(); err == nil {
		return nil
	}

	// Token is invalid or expired, proceed with renewal
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
func (s *AccessTokenSessionStore) Validate() error {
	cfg, err := s.getConfig()
	if err != nil {
		return err
	}

	if err := ValidateAccessTokenFormat(cfg.Token); err != nil {
		return fmt.Errorf("invalid access token format: %w", err)
	}

	if err := ValidateExpiry(cfg.ExpiresAt); err != nil {
		return fmt.Errorf("validating access token: %w", err)
	}

	// Run external validation if available
	if s.externalValidator != nil {
		if err := s.externalValidator(cfg.Token); err != nil {
			return err
		}
	}

	return nil
}

// HandleError processes errors that occur during session operations by dispatching
// them to registered error handlers. If no handlers are registered for the given
// error type, it returns the error wrapped with additional context.
func (s *AccessTokenSessionStore) HandleError(reason error) error {
	handlers := s.errHandlerRegistry.GetHandlers(reason)
	if len(handlers) == 0 {
		return fmt.Errorf("handling session error: %w", reason)
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
		return s.HandleError(err)
	}

	if resp == nil {
		return errors.New("renewal API returned nil response")
	}

	expTime, err := time.Parse(internal.ServerDateFormat, resp.ExpiresAt)
	if err != nil {
		return fmt.Errorf("parsing expiry time: %w", err)
	}

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

// GetToken returns the current access token or empty string if not available
func (s *AccessTokenSessionStore) GetToken() string {
	cfg, err := s.getConfig()
	if err != nil {
		return ""
	}
	return cfg.Token
}
