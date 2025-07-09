package core

import (
	"errors"
	"fmt"
	"log"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/google/uuid"
)

type TokenRenewAPICall func(token string, idempotencyKey uuid.UUID) (*TokenRenewResponse, error)

type LoginTokenManager struct {
	cfgManager         config.Manager
	tokenRenewAPICall  TokenRenewAPICall
	errHandlerRegistry *ErrorHandlingRegistry[func(uid int64)]
	validator          TokenValidator
}

// NewLoginTokenManager creates new login-token manager object
func NewLoginTokenManager(
	cfgManager config.Manager,
	tokenRenewAPICall TokenRenewAPICall,
	errorHandlingRegistry *ErrorHandlingRegistry[func(uid int64)],
	tokenValidator TokenValidator,
) TokenManager {
	return &LoginTokenManager{
		cfgManager:         cfgManager,
		tokenRenewAPICall:  tokenRenewAPICall,
		errHandlerRegistry: errorHandlingRegistry,
		validator:          tokenValidator,
	}
}

// Token retrieves the login token assigned to the current user.
func (l *LoginTokenManager) Token() (string, error) {
	var cfg config.Config
	if err := l.cfgManager.Load(&cfg); err != nil {
		return "", err
	}

	data, ok := cfg.TokensData[cfg.AutoConnectData.ID]
	if !ok {
		return "", fmt.Errorf("there is no token")
	}

	return data.Token, nil
}

// Renew iterates over stored user tokens and attempts to renew each one that has expired.
// Returns a combined error if one or more renewals fail.
func (l *LoginTokenManager) Renew() error {
	var cfg config.Config
	if err := l.cfgManager.Load(&cfg); err != nil {
		return err
	}

	var errs error
	for uid, data := range cfg.TokensData {
		err := l.validator.Validate(data.Token, data.TokenExpiry)
		if err == nil {
			continue
		}
		errs = errors.Join(errs, err)

		if err = l.renewToken(uid, data); err != nil {
			log.Printf("[auth] %s Renewing token for uid(%v): %s\n", internal.ErrorPrefix, uid, err)
			errs = errors.Join(errs, err)
		}
	}

	return errs
}

// Store saves the given token for the current user.
// It returns an error if the user ID is not found in the configuration.
func (l *LoginTokenManager) Store(token string) error {
	var cfg config.Config
	if err := l.cfgManager.Load(&cfg); err != nil {
		return err
	}
	uid := cfg.AutoConnectData.ID
	if _, exists := cfg.TokensData[uid]; !exists {
		return fmt.Errorf("autoconnectdata id (%d) not found in TokensData", uid)
	}

	return l.cfgManager.SaveWith(func(c config.Config) config.Config {
		cfg := c.TokensData[c.AutoConnectData.ID]
		cfg.Token = token
		c.TokensData[c.AutoConnectData.ID] = cfg
		return c
	})
}

// Invalidate triggers error handlers for all stored user tokens using the provided error.
// It does not modify or remove any tokens from storage and leaves this responsibility to the
// client.
func (l *LoginTokenManager) Invalidate(reason error) error {
	var cfg config.Config
	if err := l.cfgManager.Load(&cfg); err != nil {
		return err
	}

	for uid := range cfg.TokensData {
		l.invokeClientErrorHandlers(uid, reason)
	}

	return nil
}

// renewToken attempts to renew the authentication token for the specified user.
// If token renewal fails due to specific client-side errors, registered error handlers are invoked.
func (l *LoginTokenManager) renewToken(uid int64, data config.TokenData) error {
	if err := l.tryUpdateIdempotencyKey(uid, &data); err != nil {
		return err
	}

	resp, err := l.fetchTokenData(&data)
	if err == nil {
		data.Token = resp.Token
		data.RenewToken = resp.RenewToken
		data.TokenExpiry = resp.ExpiresAt
		return l.cfgManager.SaveWith(l.loginTokenDataSaver(uid, data))
	}

	if errors.Is(err, ErrUnauthorized) || errors.Is(err, ErrNotFound) || errors.Is(err, ErrBadRequest) {
		defer l.invokeClientErrorHandlers(uid, err)
		if err := l.cfgManager.SaveWith(l.tokenDataRemover(uid)); err != nil {
			return fmt.Errorf("removing token data: %w", err)
		}
	}

	return nil
}

func (l *LoginTokenManager) tryUpdateIdempotencyKey(uid int64, data *config.TokenData) error {
	if data.IdempotencyKey != nil {
		return nil
	}

	key := uuid.New()
	data.IdempotencyKey = &key
	err := l.cfgManager.SaveWith(func(c config.Config) config.Config {
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
// This is typically used for client-side HTTP errors such as 400/401/404.
func (l *LoginTokenManager) invokeClientErrorHandlers(uid int64, err error) {
	for _, handler := range l.errHandlerRegistry.GetHandlers(err) {
		handler(uid)
	}
}

// fetchTokenData calls the remote API to refresh the token using the current 'renew token'
// and idempotency key.
func (l *LoginTokenManager) fetchTokenData(data *config.TokenData) (*TokenRenewResponse, error) {
	resp, err := l.tokenRenewAPICall(data.RenewToken, *data.IdempotencyKey)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// tokenDataRemover returns a function that removes the token data associated with the user ID
// from the config object.
func (l *LoginTokenManager) tokenDataRemover(uid int64) config.SaveFunc {
	return func(c config.Config) config.Config {
		delete(c.TokensData, uid)
		return c
	}
}

// loginTokenDataSaver returns a function that stores token-related fields (token, renew token,
// and expiry timestamp) for the specified user into the configuration.
func (l *LoginTokenManager) loginTokenDataSaver(uid int64, data config.TokenData) config.SaveFunc {
	return func(c config.Config) config.Config {
		user := c.TokensData[uid]
		user.Token = data.Token
		user.RenewToken = data.RenewToken
		user.TokenExpiry = data.TokenExpiry
		c.TokensData[uid] = user
		return c
	}
}
