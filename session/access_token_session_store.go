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
}

// NewAccessTokenSessionStore creates a new AccessTokenSessionStore instance
func NewAccessTokenSessionStore(
	cfgManager config.Manager,
	errorHandlingRegistry *internal.ErrorHandlingRegistry[error],
	renewAPICall AccessTokenRenewalAPICall,
) *AccessTokenSessionStore {
	return &AccessTokenSessionStore{
		cfgManager:         cfgManager,
		errHandlerRegistry: errorHandlingRegistry,
		renewAPICall:       renewAPICall,
	}
}

// Renew checks if the access token needs renewal and renews it if necessary.
// By default, errors are processed through the error handling registry.
// Use SilentRenewal() option to perform renewal without triggering side effects.
// Use ForceRenewal() option to force renewal even if the token is valid.
func (s *AccessTokenSessionStore) Renew(opts ...RenewalOption) error {
	options := renewalOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	if !options.forceRenewal {
		if err := s.validate(); err == nil {
			return nil
		}
	}

	var cfg config.Config
	if err := s.cfgManager.Load(&cfg); err != nil {
		return err
	}

	uid := cfg.AutoConnectData.ID
	data, ok := cfg.TokensData[uid]
	if !ok {
		return errors.New("no token data during access token session renewal")
	}

	if err := s.renewToken(uid, data, options.skipErrorHandlers); err != nil {
		log.Printf("[auth] %s Renewing token for uid(%v): %s\n", internal.ErrorPrefix, uid, err)
		return err
	}

	return nil
}

func (s *AccessTokenSessionStore) validate() error {
	cfg, err := s.getConfig()
	if err != nil {
		return err
	}

	if err := ValidateExpiry(cfg.ExpiresAt); err != nil {
		return fmt.Errorf("validating access token: %w", err)
	}

	if err := ValidateAccessTokenFormat(cfg.Token); err != nil {
		return fmt.Errorf("validating access token format: %w", err)
	}

	return nil
}

// HandleError processes errors that occur during session operations.
// It returns nil if no handlers are registered, or a wrapped error if handlers were called.
func (s *AccessTokenSessionStore) HandleError(reason error) error {
	handlers := s.errHandlerRegistry.GetHandlers(reason)
	if len(handlers) == 0 {
		log.Println(internal.InfoPrefix, "No handlers for access token session store is registered")
		return nil
	}

	for _, handler := range handlers {
		handler(reason)
	}

	return fmt.Errorf("handling session error: %w", reason)
}

func (s *AccessTokenSessionStore) renewToken(
	uid int64,
	data config.TokenData,
	skipErrorHandlers bool,
) error {
	if s.renewAPICall == nil {
		return errors.New("renewal api call not configured")
	}

	if err := s.tryUpdateIdempotencyKey(uid, &data); err != nil {
		return err
	}

	resp, err := s.renewAPICall(data.Token, *data.IdempotencyKey)
	if err != nil {
		if skipErrorHandlers {
			return err
		}
		return s.HandleError(err)
	}

	if resp == nil {
		return ErrMissingAccessTokenResponse
	}

	if err := ValidateAccessTokenFormat(resp.Token); err != nil {
		return err
	}

	if err := ValidateRenewToken(resp.RenewToken); err != nil {
		return err
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
		return fmt.Errorf("saving access token data: %w", err)
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
		return accessTokenConfig{}, errors.New("non existing data for access token session store")
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

// GetToken returns the current access token or empty string if not available
func (s *AccessTokenSessionStore) GetToken() string {
	cfg, err := s.getConfig()
	if err != nil {
		return ""
	}
	return cfg.Token
}
