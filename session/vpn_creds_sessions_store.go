package session

import (
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

type VPNCredentialsResponse struct {
	Username           string
	Password           string
	NordLynxPrivateKey string
}

type VPNCredentialsRenewalAPICall func() (*VPNCredentialsResponse, error)

type VPNCredentialsSessionStore struct {
	cfgManager         config.Manager
	errHandlerRegistry *internal.ErrorHandlingRegistry[error]
	validator          SessionStoreValidator
	renewAPICall       VPNCredentialsRenewalAPICall
	session            *vpnCredentialsSession
}

// NewVPNCredentialsSessionStore create new VPN credential session store
func NewVPNCredentialsSessionStore(
	cfgManager config.Manager,
	errorHandlingRegistry *internal.ErrorHandlingRegistry[error],
	validator SessionStoreValidator,
	renewAPICall VPNCredentialsRenewalAPICall,
) SessionStore {
	return &VPNCredentialsSessionStore{
		cfgManager:         cfgManager,
		errHandlerRegistry: errorHandlingRegistry,
		validator:          validator,
		renewAPICall:       renewAPICall,
		session:            newVPNCredentialsSession(cfgManager),
	}
}

// Renew VPN credentials if they have expired.
func (s *VPNCredentialsSessionStore) Renew() error {
	if err := s.validator.Validate(s.session); err == nil { // everything's valid and up-to-date
		return nil
	}

	resp, err := s.renewAPICall()
	if err != nil {
		return s.Invalidate(err)
	}

	if err := s.session.SetNordlynxPrivateKey(resp.NordLynxPrivateKey); err != nil {
		return err
	}

	if err := s.session.SetUsername(resp.Username); err != nil {
		s.session.reset()
		return err
	}

	if err := s.session.SetPassword(resp.Password); err != nil {
		s.session.reset()
		return err
	}

	return nil
}

// Invalidate triggers registered error handlers.
// It does not modify or remove any tokens from storage and leaves this responsibility to the
// client.
func (s *VPNCredentialsSessionStore) Invalidate(reason error) error {
	handlers := s.errHandlerRegistry.GetHandlers(reason)
	if len(handlers) == 0 {
		return fmt.Errorf("invalidating session: %w", reason)
	}

	for _, handler := range handlers {
		handler(reason)
	}
	return nil
}
