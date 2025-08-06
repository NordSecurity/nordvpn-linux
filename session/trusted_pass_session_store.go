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
	// TrustedPassOwnerID is the expected owner ID for TrustedPass sessions
	TrustedPassOwnerID = "nordvpn"
)

// TrustedPassAccessTokenResponse represents the response from the TrustedPass token renewal API
type TrustedPassAccessTokenResponse struct {
	Token   string
	OwnerID string
}

// TrustedPassRenewalAPICall renews TrustedPass tokens
type TrustedPassRenewalAPICall func(token string) (*TrustedPassAccessTokenResponse, error)

// TrustedPassSessionStore manages TrustedPass-based sessions
type TrustedPassSessionStore struct {
	cfgManager         config.Manager
	errHandlerRegistry *internal.ErrorHandlingRegistry[error]
	renewAPICall       TrustedPassRenewalAPICall

	// optional external validator with
	externalValidator TrustedPassExternalValidator
}

// NewTrustedPassSessionStore creates a new TrustedPassSessionStore instance
func NewTrustedPassSessionStore(
	cfgManager config.Manager,
	errHandlerRegistry *internal.ErrorHandlingRegistry[error],
	renewAPICall TrustedPassRenewalAPICall,
	externalValidator TrustedPassExternalValidator,
) SessionStore {
	return &TrustedPassSessionStore{
		cfgManager:         cfgManager,
		renewAPICall:       renewAPICall,
		errHandlerRegistry: errHandlerRegistry,
		externalValidator:  externalValidator,
	}
}

// Renew renews the session if invalid or expired
func (s *TrustedPassSessionStore) Renew() error {
	var cfg config.Config
	if err := s.cfgManager.Load(&cfg); err != nil {
		return err
	}

	uid := cfg.AutoConnectData.ID
	data, ok := cfg.TokensData[uid]
	if !ok {
		return fmt.Errorf("there is no data")
	}

	// check if everything is valid or data renewal is required
	if err := s.validate(); err != nil {
		if err = s.renewIfOAuth(uid, &data); err != nil {
			return err
		}
	}

	// TODO: is this still necessary?
	// TrustedPass was introduced later on, so it's possible that valid data is not stored even though renew token
	// is still valid. In such cases we need to hit the api to get the initial value.
	isNotValid := (data.TrustedPassToken == "" || data.TrustedPassOwnerID == "")
	if isNotValid {
		if err := s.renewIfOAuth(uid, &data); err != nil {
			return err
		}
	}

	return nil
}

// HandleError processes errors that occur during session operations by dispatching
// them to registered error handlers. If no handlers are registered for the given
// error type, it returns the error wrapped with additional context.
func (s *TrustedPassSessionStore) HandleError(reason error) error {
	handlers := s.errHandlerRegistry.GetHandlers(reason)
	if len(handlers) == 0 {
		return fmt.Errorf("handling session error: %w", reason)
	}

	for _, handler := range handlers {
		handler(reason)
	}
	return nil
}

// validate performs validation on the TrustedPass session
func (s *TrustedPassSessionStore) validate() error {
	cfg, err := s.getConfig()
	if err != nil {
		return err
	}

	if err := ValidateTrustedPassTokenFormat(cfg.Token); err != nil {
		return err
	}

	if err := ValidateExpiry(cfg.ExpiresAt); err != nil {
		return err
	}

	if err := ValidateTrustedPassOwnerID(cfg.OwnerID); err != nil {
		return err
	}

	if s.externalValidator != nil {
		return s.externalValidator(cfg.Token, cfg.OwnerID)
	}

	return nil
}

func (s *TrustedPassSessionStore) renewToken(uid int64, data *config.TokenData) error {
	if s.renewAPICall == nil {
		return errors.New("renewal API call not configured")
	}

	resp, err := s.renewAPICall(data.Token)
	if err != nil {
		return fmt.Errorf("getting trusted pass token data: %w", err)
	}

	if resp == nil {
		return errors.New("renewal API returned nil response")
	}

	if resp.Token == "" {
		return errors.New("renewal API returned empty token")
	}

	expiryTime := time.Now().Add(trustedPassExpiryPeriod)
	err = s.cfgManager.SaveWith(func(c config.Config) config.Config {
		data := c.TokensData[uid]
		data.TrustedPassToken = resp.Token
		data.TrustedPassOwnerID = resp.OwnerID
		data.TrustedPassTokenExpiry = expiryTime.Format(internal.ServerDateFormat)
		c.TokensData[uid] = data
		return c
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *TrustedPassSessionStore) renewIfOAuth(uid int64, data *config.TokenData) error {
	if !data.IsOAuth {
		return nil
	}

	if err := s.renewToken(uid, data); err != nil {
		return s.HandleError(err)
	}

	return nil
}

// trustedPassConfig holds the TrustedPass session configuration
type trustedPassConfig struct {
	Token     string
	OwnerID   string
	ExpiresAt time.Time
}

// getConfig retrieves the current TrustedPass session configuration
func (s *TrustedPassSessionStore) getConfig() (trustedPassConfig, error) {
	var cfg config.Config
	if err := s.cfgManager.Load(&cfg); err != nil {
		return trustedPassConfig{}, err
	}

	data, ok := cfg.TokensData[cfg.AutoConnectData.ID]
	if !ok {
		return trustedPassConfig{}, errors.New("non existing data")
	}

	expiryTime, err := time.Parse(internal.ServerDateFormat, data.TrustedPassTokenExpiry)
	if err != nil {
		expiryTime = time.Now().Add(-1 * time.Second)
	}

	return trustedPassConfig{
		Token:     data.TrustedPassToken,
		OwnerID:   data.TrustedPassOwnerID,
		ExpiresAt: expiryTime,
	}, nil
}
