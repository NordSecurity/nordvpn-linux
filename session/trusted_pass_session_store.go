package session

import (
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
	errHandlerRegistry *internal.ErrorHandlingRegistry[error]
	validator          SessionStoreValidator
	renewAPICall       TrustedPassRenewalAPICall
	session            *trustedPassSession
}

// NewTrustedPassSessionStore
func NewTrustedPassSessionStore(
	cfgManager config.Manager,
	errHandlerRegistry *internal.ErrorHandlingRegistry[error],
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
func (s *TrustedPassSessionStore) Renew() error {
	var cfg config.Config
	if err := s.cfgManager.Load(&cfg); err != nil {
		return err
	}

	data, ok := cfg.TokensData[cfg.AutoConnectData.ID]
	if !ok {
		return fmt.Errorf("there is no data")
	}

	// check if everything is valid or data renewal is required
	if err := s.validator.Validate(s); err != nil {
		if err = s.renewIfOAuth(&data); err != nil {
			return err
		}
	}

	// TODO: is this still necessary?
	// TrustedPass was introduced later on, so it's possible that valid data is not stored even though renew token
	// is still valid. In such cases we need to hit the api to get the initial value.
	isNotValid := (data.TrustedPassToken == "" || data.TrustedPassOwnerID == "")
	if isNotValid {
		if err := s.renewIfOAuth(&data); err != nil {
			return err
		}
	}

	return nil
}

func (s *TrustedPassSessionStore) Invalidate(reason error) error {
	handlers := s.errHandlerRegistry.GetHandlers(reason)
	if len(handlers) == 0 {
		return fmt.Errorf("invalidating session: %w", reason)
	}

	for _, handler := range handlers {
		handler(reason)
	}
	return nil
}

func (s *TrustedPassSessionStore) renewToken(data *config.TokenData) error {
	resp, err := s.renewAPICall(data.Token)
	if err != nil {
		return fmt.Errorf("getting trusted pass token data: %w", err)
	}

	if err := s.SetToken(resp.Token); err != nil {
		return err
	}

	if err := s.SetOwnerID(resp.OwnerID); err != nil {
		s.session.reset()
		return err
	}

	if err = s.SetExpiry(time.Now().Add(trustedPassExpiryPeriod)); err != nil {
		s.session.reset()
		return err
	}

	return nil
}

func (s *TrustedPassSessionStore) renewIfOAuth(data *config.TokenData) error {
	if !data.IsOAuth {
		return nil
	}

	if err := s.renewToken(data); err != nil {
		return s.Invalidate(err)
	}

	return nil
}
