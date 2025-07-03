package core

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/google/uuid"
)

const (
	authTokenLogTag = "[auth-token]"
)

// TokenManager manages access tokens including renewal and storage
type TokenManager interface {
	// Token returns a valid token, renewing it if necessary.
	Token() (string, error)

	// Renew tries to renew the token (regardless of validity).
	Renew() error

	// StoreToken allows injecting a freshly obtained token manually (e.g., during login).
	StoreToken(token string) error

	// Invalidate clears the current token and forces a refresh on next call.
	Invalidate(reason error) error
}

type TokenRenewAPICall func(token string, idempotencyKey uuid.UUID) (*TokenRenewResponse, error)

type LoginTokenManager struct {
	cm                 config.Manager
	tokenRenewAPICall  TokenRenewAPICall
	errHandlerRegistry *ErrorHandlingRegistry[func(uid int64)]
	expChecker         ExpirationChecker
	mutex              sync.Mutex
}

// NewLoginTokenManager
func NewLoginTokenManager(
	cm config.Manager,
	tokenRenewAPICall TokenRenewAPICall,
	errorHandlingRegistry *ErrorHandlingRegistry[func(uid int64)],
	loginTokenExpirationChecker ExpirationChecker,
) TokenManager {
	return &LoginTokenManager{
		cm:                 cm,
		tokenRenewAPICall:  tokenRenewAPICall,
		errHandlerRegistry: errorHandlingRegistry,
		expChecker:         loginTokenExpirationChecker,
	}
}

// Token
func (l *LoginTokenManager) Token() (string, error) {
	var cfg config.Config
	if err := l.cm.Load(&cfg); err != nil {
		return "", err
	}

	data, ok := cfg.TokensData[cfg.AutoConnectData.ID]
	if !ok {
		return "", fmt.Errorf("there is no token")
	}

	return data.Token, nil
}

// Renew
func (l *LoginTokenManager) Renew() error {
	var cfg config.Config
	if err := l.cm.Load(&cfg); err != nil {
		return err
	}

	var errs error
	for uid, data := range cfg.TokensData {
		if !l.expChecker.IsExpired(data.TokenExpiry) {
			continue
		}

		if err := l.renewToken(uid, data); err != nil {
			log.Printf("%s %s Renewing token for uid(%v): %s\n",
				authTokenLogTag, internal.ErrorPrefix, uid, err)
			errs = errors.Join(errs, err)
		}

	}

	return errs
}

// StoreToken
func (l *LoginTokenManager) StoreToken(token string) error {
	var cfg config.Config
	if err := l.cm.Load(&cfg); err != nil {
		return err
	}
	uid := cfg.AutoConnectData.ID
	if _, exists := cfg.TokensData[uid]; !exists {
		return fmt.Errorf("autoconnectdata id (%d) not found in TokensData", uid)
	}

	return l.cm.SaveWith(func(c config.Config) config.Config {
		cfg := c.TokensData[c.AutoConnectData.ID]
		cfg.Token = token
		c.TokensData[c.AutoConnectData.ID] = cfg
		return c
	})
}

func (l *LoginTokenManager) Invalidate(reason error) error {
	var cfg config.Config
	if err := l.cm.Load(&cfg); err != nil {
		return err
	}

	for uid := range cfg.TokensData {
		l.invokeClientErrorHandlers(uid, reason)
	}

	return nil
}

func (l *LoginTokenManager) renewToken(uid int64, data config.TokenData) error {
	if data.IdempotencyKey == nil {
		key := uuid.New()
		data.IdempotencyKey = &key
		if err := l.cm.SaveWith(l.idempotencyKeySaver(uid, data)); err != nil {
			return fmt.Errorf("saving idempotency key: %w", err)
		}
	}

	err := l.fetchTokenData(&data)
	if err == nil {
		return l.cm.SaveWith(l.loginTokenDataSaver(uid, data))
	}

	if errors.Is(err, ErrUnauthorized) || errors.Is(err, ErrNotFound) || errors.Is(err, ErrBadRequest) {
		defer l.invokeClientErrorHandlers(uid, err)
		if err := l.cm.SaveWith(l.tokenDataRemover(uid)); err != nil {
			return fmt.Errorf("removing token data: %w", err)
		}
	}

	return err
}

func (l *LoginTokenManager) invokeClientErrorHandlers(uid int64, err error) {
	for _, handler := range l.errHandlerRegistry.GetHandlers(err) {
		handler(uid)
	}
}

func (l *LoginTokenManager) fetchTokenData(data *config.TokenData) error {
	resp, err := l.tokenRenewAPICall(data.RenewToken, *data.IdempotencyKey)
	if err != nil {
		return err
	}

	data.Token = resp.Token
	data.RenewToken = resp.RenewToken
	data.TokenExpiry = resp.ExpiresAt
	return nil
}

func (l *LoginTokenManager) tokenDataRemover(uid int64) config.SaveFunc {
	return func(c config.Config) config.Config {
		delete(c.TokensData, uid)
		return c
	}
}

func (l *LoginTokenManager) idempotencyKeySaver(uid int64, data config.TokenData) config.SaveFunc {
	return func(c config.Config) config.Config {
		user := c.TokensData[uid]
		user.IdempotencyKey = data.IdempotencyKey
		c.TokensData[uid] = user
		return c
	}
}

// loginTokenDataSaver persists only token related data
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
