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

// Renew renews the NC credentials session
func (s *NCCredentialsSessionStore) Renew() error {
	if err := s.validate(); err == nil { // Credentials are still valid
		return nil
	}

	if s.renewAPICall == nil {
		return fmt.Errorf("renewal API not configured")
	}

	resp, err := s.renewAPICall()
	if err != nil {
		return s.HandleError(err)
	}

	if resp == nil {
		return fmt.Errorf("renewal API returned nil response")
	}

	if err := ValidateNCCredentialsPresence(resp.Username, resp.Password); err != nil {
		return err
	}

	if err := ValidateEndpointPresence(resp.Endpoint); err != nil {
		return err
	}

	expiryTime := time.Now().UTC().Add(resp.ExpiresIn)

	err = s.cfgManager.SaveWith(func(c config.Config) config.Config {
		if c.TokensData == nil {
			c.TokensData = make(map[int64]config.TokenData)
		}

		data, ok := c.TokensData[c.AutoConnectData.ID]
		if !ok {
			data = config.TokenData{}
		}
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
// It returns nil if the error was not handled, or the error itself if it was.
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

// validate performs validation on the NC credentials session
func (s *NCCredentialsSessionStore) validate() error {
	cfg, err := s.getConfig()
	if err != nil {
		return err
	}

	if err := ValidateExpiry(cfg.ExpiresAt); err != nil {
		return err
	}

	if err := ValidateNCCredentialsPresence(cfg.Username, cfg.Password); err != nil {
		return err
	}

	if err := ValidateEndpointPresence(cfg.Endpoint); err != nil {
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
