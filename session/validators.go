package session

import (
	"fmt"
	"log"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

type InvalidSessionValidator struct{}

func (v *InvalidSessionValidator) Validate(session interface{}) error {
	return fmt.Errorf("invalid session validator constructed")
}

type TokenValidator struct {
	ValidateFunc func(token string) error
}

// Validate checks if the provided session has a valid token.
func (v *TokenValidator) Validate(session interface{}) error {
	carrier, ok := session.(SessionTokenProvider)
	if !ok {
		return fmt.Errorf("unsupported store for token validation: got type %T", session)
	}

	return v.ValidateFunc(carrier.GetToken())
}

// NewTokenValidator creates a validator that checks session tokens.
func NewTokenValidator(fn func(token string) error) SessionStoreValidator {
	if fn == nil {
		log.Println(
			internal.WarningPrefix,
			"TokenValidator creation failed due to an invalid function argument.")
		return &InvalidSessionValidator{}
	}
	return &TokenValidator{ValidateFunc: fn}
}

type ExpiryValidator struct{}

// Validate checks if the provided session has expired.
func (v *ExpiryValidator) Validate(session interface{}) error {
	expirable, ok := session.(ExpirableSession)
	if !ok {
		return fmt.Errorf("unsupported store for expiration validation: got type %T", session)
	}

	if expirable.IsExpired() {
		return ErrSessionExpired
	}

	return nil
}

// NewExpiryValidator creates a validator that checks session expiry.
func NewExpiryValidator() SessionStoreValidator {
	return &ExpiryValidator{}
}

type CredentialValidator struct {
	ValidateFunc func(username, password string) error
}

// Validate checks if the provided session has valid credentials.
func (v *CredentialValidator) Validate(session interface{}) error {
	carrier, ok := session.(CredentialsBasedSession)
	if !ok {
		return fmt.Errorf("unsupported store for credentials validation: got type %T", session)
	}

	return v.ValidateFunc(carrier.GetUsername(), carrier.GetPassword())
}

// NewCredentialValidator creates a validator that checks session credentials.
func NewCredentialValidator(fn func(username, password string) error) SessionStoreValidator {
	if fn == nil {
		log.Println(
			internal.WarningPrefix,
			"CredentialValidator creation failed due to an invalid function argument.")
		return &InvalidSessionValidator{}
	}
	return &CredentialValidator{ValidateFunc: fn}
}

type CompositeValidator struct {
	validators []SessionStoreValidator
}

// Validate runs all validators in sequence, stopping at first error.
func (c *CompositeValidator) Validate(session interface{}) error {
	for _, v := range c.validators {
		if err := v.Validate(session); err != nil {
			return err
		}
	}

	return nil
}

// NewCompositeValidator creates a validator that runs multiple validators sequentially.
func NewCompositeValidator(validators ...SessionStoreValidator) SessionStoreValidator {
	if len(validators) == 0 {
		log.Println(
			internal.WarningPrefix,
			"CompositeValidator creation failed due to empty arguments.")
		return &InvalidSessionValidator{}
	}
	return &CompositeValidator{validators: validators}
}
