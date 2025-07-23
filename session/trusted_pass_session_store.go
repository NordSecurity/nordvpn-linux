package session

import (
	"errors"
	"fmt"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

const (
	// it's predefined value, but not retrievable from any API
	trustedPassExpiryPeriod = time.Hour * 24
)

type TrustedPassAccessTokenResponse struct {
	Token   string
	OwnerID string
}

type TrustedPassRenewalAPICall func(token string) (*TrustedPassAccessTokenResponse, error)

type TrustedPassSessionStore struct {
	cfgManager         config.Manager
	errHandlerRegistry *internal.ErrorHandlingRegistry[int64]
	validator          SessionStoreValidator
	renewAPICall       TrustedPassRenewalAPICall
	session            *trustedPassSession
}

// NewTrustedPassSessionStore
func NewTrustedPassSessionStore(
	cfgManager config.Manager,
	errHandlerRegistry *internal.ErrorHandlingRegistry[int64],
	validator SessionStoreValidator,
	renewAPICall TrustedPassRenewalAPICall,
) SessionStore {
	return &TrustedPassSessionStore{
		cfgManager:         cfgManager,
		renewAPICall:       renewAPICall,
		errHandlerRegistry: errHandlerRegistry,
		validator:          validator,
		session:            newTrustedPassSession(cfgManager),
	}
}
func (m *TrustedPassSessionStore) Renew() error {
	var cfg config.Config
	if err := m.cfgManager.Load(&cfg); err != nil {
		return err
	}

	data, ok := cfg.TokensData[cfg.AutoConnectData.ID]
	if !ok {
		return fmt.Errorf("there is no data")
	}

	// check if everything is valid or data renewal is required
	if err := m.validator.Validate(m.session); err != nil {
		if err = m.renewIfOAuth(&data, cfg.AutoConnectData.ID); err != nil {
			return err
		}
	}

	// TODO: is this still necessary?
	// TrustedPass was introduced later on, so it's possible that valid data is not stored even though renew token
	// is still valid. In such cases we need to hit the api to get the initial value.
	isNotValid := (data.TrustedPassToken == "" || data.TrustedPassOwnerID == "")
	if isNotValid {
		if err := m.renewIfOAuth(&data, cfg.AutoConnectData.ID); err != nil {
			return err
		}
	}

	return nil
}

func (m *TrustedPassSessionStore) Invalidate(reason error) error {
	var cfg config.Config
	if err := m.cfgManager.Load(&cfg); err != nil {
		return err
	}

	for uid := range cfg.TokensData {
		m.invokeClientErrorHandlers(uid, reason)
	}

	return nil
}

// invokeClientErrorHandlers executes all registered error handlers associated with the provided
// error for the given user ID.
func (s *TrustedPassSessionStore) invokeClientErrorHandlers(uid int64, err error) {
	for _, handler := range s.errHandlerRegistry.GetHandlers(err) {
		handler(uid)
	}
}

func (m *TrustedPassSessionStore) renewToken(data *config.TokenData) error {
	resp, err := m.renewAPICall(data.Token)
	if err != nil {
		return fmt.Errorf("getting trusted pass token data: %w", err)
	}

	if err := m.session.SetToken(resp.Token); err != nil {
		return err
	}

	if err := m.session.SetOwnerID(resp.OwnerID); err != nil {
		return err
	}

	if err = m.session.SetExpiry(time.Now().Add(trustedPassExpiryPeriod)); err != nil {
		return err
	}

	return nil
}

func (m *TrustedPassSessionStore) isLogoutNeeded(err error) bool {
	return errors.Is(err, ErrUnauthorized) ||
		errors.Is(err, ErrNotFound) ||
		errors.Is(err, ErrBadRequest)
}

func (m *TrustedPassSessionStore) doLogout(reason error, uid int64) {
	// err  handler will take care of event publishing (which will need to be catched for logout)
	for _, handler := range m.errHandlerRegistry.GetHandlers(reason) {
		handler(uid)
	}
}

func (m *TrustedPassSessionStore) renewIfOAuth(data *config.TokenData, uid int64) error {
	if !data.IsOAuth {
		return nil
	}

	if err := m.renewToken(data); err != nil {
		if m.isLogoutNeeded(err) {
			m.doLogout(err, uid)
		}
		return err
	}

	return nil
}
