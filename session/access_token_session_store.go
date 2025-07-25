package session

import (
	"errors"
	"fmt"
	"log"

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

type AccessTokenSessionStore struct {
	cfgManager         config.Manager
	errHandlerRegistry *internal.ErrorHandlingRegistry[int64]
	validator          SessionStoreValidator
	renewAPICall       AccessTokenRenewalAPICall
	session            *accessTokenSession
}

// NewAccessTokenSessionStore
func NewAccessTokenSessionStore(
	cfgManager config.Manager,
	validator SessionStoreValidator,
	errorHandlingRegistry *internal.ErrorHandlingRegistry[int64],
	renewAPICall AccessTokenRenewalAPICall,
) SessionStore {
	return &AccessTokenSessionStore{
		cfgManager:         cfgManager,
		errHandlerRegistry: errorHandlingRegistry,
		validator:          validator,
		renewAPICall:       renewAPICall,
		session:            newAccessTokenSession(cfgManager),
	}
}

// Renew iterates over stored user tokens and attempts to renew each one that has expired.
// Returns a combined error if one or more renewals fail.
func (s *AccessTokenSessionStore) Renew() error {
	var cfg config.Config
	if err := s.cfgManager.Load(&cfg); err != nil {
		return err
	}

	uid := cfg.AutoConnectData.ID
	data, ok := cfg.TokensData[uid]
	if !ok {
		return errors.New("no token data")
	}

	err := s.validator.Validate(s)
	if err == nil {
		return nil
	}
	// if this error happens, then there is no way to recover
	if errors.Is(err, ErrAccessTokenRevoked) {
		return err
	}

	if err := s.renewToken(uid, data); err != nil {
		log.Printf("[auth] %s Renewing token for uid(%v): %s\n", internal.ErrorPrefix, uid, err)
		return err
	}

	return nil
}

// Invalidate triggers error handlers for all stored user tokens using the provided error.
// It does not modify or remove any tokens from storage and leaves this responsibility to the
// client.
func (s *AccessTokenSessionStore) Invalidate(reason error) error {
	var cfg config.Config
	if err := s.cfgManager.Load(&cfg); err != nil {
		return err
	}

	for uid := range cfg.TokensData {
		s.invokeClientErrorHandlers(uid, reason)
	}

	return nil
}

// renewToken attempts to renew the authentication token for the specified user.
// If token renewal fails due to specific client-side errors, registered error handlers are invoked.
func (s *AccessTokenSessionStore) renewToken(uid int64, data config.TokenData) error {
	if err := s.tryUpdateIdempotencyKey(uid, &data); err != nil {
		return err
	}

	resp, err := s.renewAPICall(data.Token, *data.IdempotencyKey)
	if err == nil {
		data.Token = resp.Token
		data.RenewToken = resp.RenewToken
		data.TokenExpiry = resp.ExpiresAt
		return s.cfgManager.SaveWith(s.loginTokenDataSaver(uid, data))
	}

	if errors.Is(err, ErrUnauthorized) || errors.Is(err, ErrNotFound) || errors.Is(err, ErrBadRequest) {
		defer s.invokeClientErrorHandlers(uid, err)
		if err := s.cfgManager.SaveWith(func(c config.Config) config.Config {
			delete(c.TokensData, uid)
			return c
		}); err != nil {
			return fmt.Errorf("removing token data: %w", err)
		}
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

// invokeClientErrorHandlers executes all registered error handlers associated with the provided
// error for the given user ID.
func (s *AccessTokenSessionStore) invokeClientErrorHandlers(uid int64, err error) {
	for _, handler := range s.errHandlerRegistry.GetHandlers(err) {
		handler(uid)
	}
}

// loginTokenDataSaver returns a function that stores token-related fields (token, renew token,
// and expiry timestamp) for the specified user into the accessTokenConfig.
func (s *AccessTokenSessionStore) loginTokenDataSaver(uid int64, data config.TokenData) config.SaveFunc {
	return func(c config.Config) config.Config {
		user := c.TokensData[uid]
		user.Token = data.Token
		user.RenewToken = data.RenewToken
		user.TokenExpiry = data.TokenExpiry
		c.TokensData[uid] = user
		return c
	}
}
