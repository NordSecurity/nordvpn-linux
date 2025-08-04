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

type AccessTokenSessionStore struct {
	cfgManager         config.Manager
	errHandlerRegistry *internal.ErrorHandlingRegistry[error]
	validator          SessionStoreValidator
	renewAPICall       AccessTokenRenewalAPICall
	session            *accessTokenSession
}

// NewAccessTokenSessionStore
func NewAccessTokenSessionStore(
	cfgManager config.Manager,
	validator SessionStoreValidator,
	errorHandlingRegistry *internal.ErrorHandlingRegistry[error],
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

// Renew attempts to renew a token if it has expired.
func (s *AccessTokenSessionStore) Renew() error {
	err := s.validator.Validate(s)
	if err == nil {
		return nil
	}

	// if this error happens, then there is no way to recover
	if errors.Is(err, ErrAccessTokenRevoked) {
		s.Invalidate(ErrAccessTokenRevoked)
		return err
	}

	var cfg config.Config
	if err := s.cfgManager.Load(&cfg); err != nil {
		return err
	}

	uid := cfg.AutoConnectData.ID
	data, ok := cfg.TokensData[uid]
	if !ok {
		return errors.New("no token data")
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
	handlers := s.errHandlerRegistry.GetHandlers(reason)
	if len(handlers) == 0 {
		return fmt.Errorf("invalidating session: %w", reason)
	}

	for _, handler := range handlers {
		handler(reason)
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
	if err != nil {
		return s.Invalidate(err)
	}

	if err = s.SetToken(resp.Token); err != nil {
		return err
	}

	if err = s.SetRenewToken(resp.RenewToken); err != nil {
		s.session.reset()
		return err
	}

	expTime, errParse := time.Parse(internal.ServerDateFormat, resp.ExpiresAt)
	if errParse != nil {
		s.session.reset()
		return err
	}

	if err = s.SetExpiry(expTime); err != nil {
		s.session.reset()
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
