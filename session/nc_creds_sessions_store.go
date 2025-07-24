package session

import (
	"fmt"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

type NCCredentialsResponse struct {
	Username  string
	Password  string
	Endpoint  string
	ExpiresIn int
}

type NCCredentialsRenewalAPICall func() (*NCCredentialsResponse, error)

type NCCredentialsSessionStore struct {
	cfgManager         config.Manager
	errHandlerRegistry *internal.ErrorHandlingRegistry[error]
	validator          SessionStoreValidator
	renewAPICall       NCCredentialsRenewalAPICall
	session            *ncCredentialsSession
}

// NewNCCredentialsSessionStore
func NewNCCredentialsSessionStore(
	cfgManager config.Manager,
	errorHandlingRegistry *internal.ErrorHandlingRegistry[error],
	validator SessionStoreValidator,
	renewAPICall NCCredentialsRenewalAPICall,
) SessionStore {
	return &NCCredentialsSessionStore{
		cfgManager:         cfgManager,
		errHandlerRegistry: errorHandlingRegistry,
		validator:          validator,
		renewAPICall:       renewAPICall,
		session:            newNCCredentialsSession(cfgManager),
	}
}

// Renew
func (s *NCCredentialsSessionStore) Renew() error {
	if err := s.validator.Validate(s.session); err == nil { // everything's valid and up-to-date
		return nil
	}

	var cfg config.Config
	if err := s.cfgManager.Load(&cfg); err != nil {
		return err
	}

	_, ok := cfg.TokensData[cfg.AutoConnectData.ID]
	if !ok {
		return fmt.Errorf("there is no data")
	}

	resp, err := s.renewAPICall()
	if err != nil {
		return s.Invalidate(err)
	}

	if err := s.session.SetUsername(resp.Username); err != nil {
		s.session.reset()
		return err
	}

	if err := s.session.SetPassword(resp.Password); err != nil {
		s.session.reset()
		return err
	}

	if err := s.session.SetEndpoint(resp.Endpoint); err != nil {
		s.session.reset()
		return err
	}

	expiresdAt := time.Now().Add(time.Duration(resp.ExpiresIn))
	if err := s.session.SetExpiry(expiresdAt); err != nil {
		s.session.reset()
		return err
	}

	return nil
}

// Invalidate triggers error handlers for all stored user tokens using the provided error.
// It does not modify or remove any tokens from storage and leaves this responsibility to the
// client.
func (s *NCCredentialsSessionStore) Invalidate(reason error) error {
	handlers := s.errHandlerRegistry.GetHandlers(reason)
	if len(handlers) == 0 {
		return fmt.Errorf("invalidating session: %w", reason)
	}

	for _, handler := range handlers {
		handler(reason)
	}
	return nil
}
