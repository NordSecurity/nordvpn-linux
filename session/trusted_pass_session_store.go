package session

import (
	"errors"
	"fmt"
	"log"
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

	// optional external validator
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

// Renew renews the session if invalid or expired.
// Use SilentRenewal() option to perform renewal without triggering side effects.
// Use ForceRenewal() option to force renewal even if the session is valid.
func (s *TrustedPassSessionStore) Renew(opts ...RenewalOption) error {
	options := &renewalOptions{}
	for _, opt := range opts {
		opt(options)
	}

	var cfg config.Config
	if err := s.cfgManager.Load(&cfg); err != nil {
		return err
	}

	uid := cfg.AutoConnectData.ID
	data, ok := cfg.TokensData[uid]
	if !ok {
		return fmt.Errorf("there is no data during during trusted pass session renewal")
	}

	// check if everything is valid or data renewal is required
	if options.forceRenewal || s.validate(false) != nil {
		if err := s.renewIfOAuth(uid, &data, options.skipErrorHandlers); err != nil {
			return err
		}
	}

	// TODO: is this still necessary?
	// TrustedPass was introduced later on, so it's possible that valid data is not stored even though renew token
	// is still valid. In such cases we need to hit the api to get the initial value.
	isNotValid := (data.TrustedPassToken == "" || data.TrustedPassOwnerID == "")
	if isNotValid {
		if err := s.renewIfOAuth(uid, &data, options.skipErrorHandlers); err != nil {
			return err
		}
	}

	return nil
}

// HandleError processes errors that occur during session operations.
// It returns nil if the error was not handled, or the error itself if it was.
func (s *TrustedPassSessionStore) HandleError(reason error) error {
	handlers := s.errHandlerRegistry.GetHandlers(reason)
	if len(handlers) == 0 {
		log.Println(internal.InfoPrefix, "No handlers for trusted pass session store is registered")
		return nil
	}

	for _, handler := range handlers {
		handler(reason)
	}

	return fmt.Errorf("handling session error: %w", reason)
}

func (s *TrustedPassSessionStore) validate(skipExpiry bool) error {
	cfg, err := s.getConfig()
	if err != nil {
		return err
	}

	if !skipExpiry {
		if err := ValidateExpiry(cfg.ExpiresAt); err != nil {
			return err
		}
	}

	if err := ValidateTrustedPassTokenFormat(cfg.Token); err != nil {
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

func (s *TrustedPassSessionStore) renewToken(uid int64, data *config.TokenData, skipErrorHandlers bool) error {
	if s.renewAPICall == nil {
		return errors.New("renewal api call not configured")
	}

	resp, err := s.renewAPICall(data.Token)
	if err != nil {
		if skipErrorHandlers {
			return err
		}
		return s.HandleError(err)
	}

	if resp == nil {
		if skipErrorHandlers {
			return ErrMissingTrustedPassResponse
		}
		return s.HandleError(ErrMissingTrustedPassResponse)
	}

	if err := ValidateTrustedPassTokenFormat(resp.Token); err != nil {
		if skipErrorHandlers {
			return ErrMissingTrustedPassResponse
		}
		return s.HandleError(ErrMissingTrustedPassResponse)
	}

	if err := ValidateTrustedPassOwnerID(resp.OwnerID); err != nil {
		if skipErrorHandlers {
			return ErrMissingTrustedPassResponse
		}
		return s.HandleError(ErrMissingTrustedPassResponse)
	}

	err = s.cfgManager.SaveWith(func(c config.Config) config.Config {
		expiryTime := time.Now().Add(trustedPassExpiryPeriod)
		data := c.TokensData[uid]
		data.TrustedPassToken = resp.Token
		data.TrustedPassOwnerID = resp.OwnerID
		data.TrustedPassTokenExpiry = expiryTime.Format(internal.ServerDateFormat)
		c.TokensData[uid] = data
		return c
	})

	if err != nil {
		return fmt.Errorf("saving trusted pass session config: %w", err)
	}

	return nil
}

func (s *TrustedPassSessionStore) renewIfOAuth(uid int64, data *config.TokenData, skipErrorHandlers bool) error {
	if !data.IsOAuth {
		return nil
	}

	if err := s.renewToken(uid, data, skipErrorHandlers); err != nil {
		if skipErrorHandlers {
			return err
		}
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
