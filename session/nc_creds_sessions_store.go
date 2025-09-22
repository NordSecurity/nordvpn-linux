package session

import (
	"fmt"
	"log"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

type NCCredentialsResponse struct {
	Username  string
	Password  string
	Endpoint  string
	ExpiresIn time.Duration
}

type NCCredentialsRenewalAPICall func() (*NCCredentialsResponse, error)

type NCCredentialsSessionStore struct {
	cfgManager         config.Manager
	errHandlerRegistry *internal.ErrorHandlingRegistry[error]
	renewAPICall       NCCredentialsRenewalAPICall
}

// NewNCCredentialsSessionStore creates a new NC credentials session store
func NewNCCredentialsSessionStore(
	cfgManager config.Manager,
	errorHandlingRegistry *internal.ErrorHandlingRegistry[error],
	renewAPICall NCCredentialsRenewalAPICall,
) SessionStore {
	return &NCCredentialsSessionStore{
		cfgManager:         cfgManager,
		errHandlerRegistry: errorHandlingRegistry,
		renewAPICall:       renewAPICall,
	}
}

// Renew renews the NC credentials session.
// Use SilentRenewal() option to perform renewal without triggering side effects.
// Use ForceRenewal() option to force renewal even if the credentials are valid.
func (s *NCCredentialsSessionStore) Renew(opts ...RenewalOption) error {
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
		return fmt.Errorf("renewal API not configured")
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
			return ErrMissingNCCredentialsResponse
		}
		return ErrMissingNCCredentialsResponse
	}

	if err := ValidateNCCredentialsPresence(resp.Username, resp.Password); err != nil {
		if options.skipErrorHandlers {
			return err
		}
		return err
	}

	if err := ValidateEndpointPresence(resp.Endpoint); err != nil {
		if options.skipErrorHandlers {
			return err
		}
		return err
	}

	err = s.cfgManager.SaveWith(func(c config.Config) config.Config {
		if c.TokensData == nil {
			c.TokensData = make(map[int64]config.TokenData)
		}

		data, ok := c.TokensData[c.AutoConnectData.ID]
		if !ok {
			data = config.TokenData{}
		}
		expiryTime := time.Now().UTC().Add(resp.ExpiresIn)

		data.NCData.Username = resp.Username
		data.NCData.Password = resp.Password
		data.NCData.Endpoint = resp.Endpoint
		data.NCData.ExpirationDate = expiryTime
		c.TokensData[c.AutoConnectData.ID] = data
		return c
	})
	if err != nil {
		return fmt.Errorf("saving renewed nc creds: %w", err)
	}

	return nil
}

// HandleError processes errors that occur during session operations.
// It returns nil if no handlers are registered, or the error itself after handlers are called.
func (s *NCCredentialsSessionStore) HandleError(err error) error {
	handlers := s.errHandlerRegistry.GetHandlers(err)
	if len(handlers) == 0 {
		log.Println(internal.InfoPrefix, "No handlers for nc creds session store is registered")
		return nil
	}

	for _, handler := range handlers {
		handler(err)
	}

	return err
}

func (s *NCCredentialsSessionStore) validate() error {
	cfg, err := s.getConfig()
	if err != nil {
		return err
	}

	if err := ValidateNCCredentialsPresence(cfg.Username, cfg.Password); err != nil {
		return err
	}

	if err := ValidateEndpointPresence(cfg.Endpoint); err != nil {
		return err
	}

	if err := ValidateExpiry(cfg.ExpiresAt); err != nil {
		return err
	}

	return nil
}

// ncCredentialsConfig holds the NC credentials session configuration
type ncCredentialsConfig struct {
	Username  string
	Password  string
	Endpoint  string
	ExpiresAt time.Time
}

// getConfig retrieves the current NC credentials session configuration
func (s *NCCredentialsSessionStore) getConfig() (ncCredentialsConfig, error) {
	var cfg config.Config
	if err := s.cfgManager.Load(&cfg); err != nil {
		return ncCredentialsConfig{}, err
	}

	data, ok := cfg.TokensData[cfg.AutoConnectData.ID]
	if !ok {
		return ncCredentialsConfig{}, fmt.Errorf("no token data found for user ID: %d", cfg.AutoConnectData.ID)
	}

	return ncCredentialsConfig{
		Username:  data.NCData.Username,
		Password:  data.NCData.Password,
		Endpoint:  data.NCData.Endpoint,
		ExpiresAt: data.NCData.ExpirationDate,
	}, nil
}
