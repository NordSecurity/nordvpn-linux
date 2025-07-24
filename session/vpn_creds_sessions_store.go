package session

import (
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
	errHandlerRegistry *internal.ErrorHandlingRegistry[int64]
	validator          SessionStoreValidator
	renewAPICall       VPNCredentialsRenewalAPICall
	session            *vpnCredentialsSession
}

// NewVPNCredentialsSessionStore
func NewVPNCredentialsSessionStore(
	cfgManager config.Manager,
	errorHandlingRegistry *internal.ErrorHandlingRegistry[int64],
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

// Renew
func (s *VPNCredentialsSessionStore) Renew() error {
	if err := s.validator.Validate(s.session); err == nil { // everything's valid and up-to-date
		return nil
	}

	resp, err := s.renewAPICall()
	if err != nil {
		return err
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

// Invalidate triggers error handlers for all stored user tokens using the provided error.
// It does not modify or remove any tokens from storage and leaves this responsibility to the
// client.
func (s *VPNCredentialsSessionStore) Invalidate(reason error) error {
	var cfg config.Config
	if err := s.cfgManager.Load(&cfg); err != nil {
		return err
	}

	for uid := range cfg.TokensData {
		s.invokeClientErrorHandlers(uid, reason)
	}

	return nil
}

// invokeClientErrorHandlers executes all registered error handlers associated with the provided
// error for the given user ID.
func (s *VPNCredentialsSessionStore) invokeClientErrorHandlers(uid int64, err error) {
	for _, handler := range s.errHandlerRegistry.GetHandlers(err) {
		handler(uid)
	}
}
