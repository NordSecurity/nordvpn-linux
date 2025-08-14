package session

import (
	"errors"
	"fmt"
	"log"

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
	renewAPICall       VPNCredentialsRenewalAPICall
	externalValidator  VPNCredentialsExternalValidator
}

// NewVPNCredentialsSessionStore create new VPN credential session store
func NewVPNCredentialsSessionStore(
	cfgManager config.Manager,
	errorHandlingRegistry *internal.ErrorHandlingRegistry[error],
	renewAPICall VPNCredentialsRenewalAPICall,
	externalValidator VPNCredentialsExternalValidator,
) SessionStore {
	return &VPNCredentialsSessionStore{
		cfgManager:         cfgManager,
		errHandlerRegistry: errorHandlingRegistry,
		renewAPICall:       renewAPICall,
		externalValidator:  externalValidator,
	}
}

// Renew VPN credentials if they have expired.
// Use SilentRenewal() option to perform renewal without triggering side effects.
// Use ForceRenewal() option to force renewal even if the credentials are valid.
func (s *VPNCredentialsSessionStore) Renew(opts ...RenewalOption) error {
	options := &renewalOptions{}
	for _, opt := range opts {
		opt(options)
	}

	if !options.forceRenewal {
		if err := s.validate(); err == nil {
			return nil
		}
	}

	if s.renewAPICall == nil {
		return errors.New("vpn creds renewal api call not configured")
	}

	resp, err := s.renewAPICall()
	if err != nil {
		if options.skipErrorHandlers {
			return err
		}
		return s.HandleError(err)
	}

	if resp == nil {
		if options.skipErrorHandlers {
			return ErrMissingVPNCredsResponse
		}
		return s.HandleError(ErrMissingVPNCredsResponse)
	}

	if err := ValidateOpenVPNCredentialsPresence(resp.Username, resp.Password); err != nil {
		if options.skipErrorHandlers {
			return ErrMissingVPNCredsResponse
		}
		return s.HandleError(ErrMissingVPNCredsResponse)
	}

	if err := ValidateNordLynxPrivateKeyPresence(resp.NordLynxPrivateKey); err != nil {
		if options.skipErrorHandlers {
			return ErrMissingVPNCredsResponse
		}
		return s.HandleError(ErrMissingVPNCredsResponse)
	}

	err = s.cfgManager.SaveWith(func(c config.Config) config.Config {
		data := c.TokensData[c.AutoConnectData.ID]
		data.NordLynxPrivateKey = resp.NordLynxPrivateKey
		data.OpenVPNUsername = resp.Username
		data.OpenVPNPassword = resp.Password
		c.TokensData[c.AutoConnectData.ID] = data
		return c
	})

	if err != nil {
		return fmt.Errorf("failed to save vpn credentials: %w", err)
	}

	return nil
}

// HandleError processes errors that occur during session operations.
// It returns nil if no handlers are registered, or a wrapped error if handlers were called.
func (s *VPNCredentialsSessionStore) HandleError(reason error) error {
	handlers := s.errHandlerRegistry.GetHandlers(reason)
	if len(handlers) == 0 {
		log.Println(internal.InfoPrefix, "No handlers for vpn creds session store is registered")
		return nil
	}

	for _, handler := range handlers {
		handler(reason)
	}

	return fmt.Errorf("handling session error: %w", reason)
}

func (s *VPNCredentialsSessionStore) validate() error {
	cfg, err := s.getConfig()
	if err != nil {
		return err
	}

	if err := ValidateOpenVPNCredentialsPresence(cfg.Username, cfg.Password); err != nil {
		return err
	}

	if err := ValidateNordLynxPrivateKeyPresence(cfg.NordLynxPrivateKey); err != nil {
		return err
	}

	if s.externalValidator != nil {
		if err := s.externalValidator(cfg.Username, cfg.Password, cfg.NordLynxPrivateKey); err != nil {
			return err
		}
	}

	return nil
}

type vpnCredentialsConfig struct {
	Username           string
	Password           string
	NordLynxPrivateKey string
}

func (s *VPNCredentialsSessionStore) getConfig() (vpnCredentialsConfig, error) {
	var cfg config.Config
	if err := s.cfgManager.Load(&cfg); err != nil {
		return vpnCredentialsConfig{}, err
	}

	data, ok := cfg.TokensData[cfg.AutoConnectData.ID]
	if !ok {
		return vpnCredentialsConfig{}, errors.New("non existing data")
	}

	return vpnCredentialsConfig{
		Username:           data.OpenVPNUsername,
		Password:           data.OpenVPNPassword,
		NordLynxPrivateKey: data.NordLynxPrivateKey,
	}, nil
}
