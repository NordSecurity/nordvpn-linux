package session

import "errors"

type TokenValidator struct {
	ValidateFunc func(token string) error
}

// Validate
func (v *TokenValidator) Validate(session any) error {
	carrier, ok := session.(SessionTokenProvider)
	if !ok {
		return errors.New("unsupported store for token validation")
	}

	return v.ValidateFunc(carrier.GetToken())
}

// NewTokenValidator
func NewTokenValidator(fn func(token string) error) SessionStoreValidator {
	return &TokenValidator{ValidateFunc: fn}
}

type ExpiryValidator struct{}

// Validate
func (v *ExpiryValidator) Validate(session any) error {
	expirable, ok := session.(ExpirableSession)
	if !ok {
		return errors.New("unsupported store for expiration validation")
	}

	if expirable.IsExpired() {
		return ErrSessionExpired
	}

	return nil
}

// NewExpiryValidator
func NewExpiryValidator() SessionStoreValidator {
	return &ExpiryValidator{}
}

type CredentialValidator struct {
	ValidateFunc func(username, password string) error
}

func (v *CredentialValidator) Validate(session any) error {
	carrier, ok := session.(CredentialsBasedSession)
	if !ok {
		return errors.New("unsupported store for credentials validation")
	}

	return v.ValidateFunc(carrier.GetUsername(), carrier.GetPassword())
}

// NewCredentialValidator
func NewCredentialValidator(fn func(username, password string) error) SessionStoreValidator {
	return &CredentialValidator{ValidateFunc: fn}
}

type CompositeValidator struct {
	Validators []SessionStoreValidator
}

// Validate
func (c *CompositeValidator) Validate(session any) error {
	for _, v := range c.Validators {
		if err := v.Validate(session); err != nil {
			return err
		}
	}

	return nil
}

// NewCompositeValidator
func NewCompositeValidator(validators ...SessionStoreValidator) SessionStoreValidator {
	return &CompositeValidator{Validators: validators}
}
