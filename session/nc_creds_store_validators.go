package session

import "errors"

var (
	ErrInvalidEndpoint      = errors.New("invalid endpoint")
	ErrInvalidNCCredentials = errors.New("invalid nc credentials")
)

type SessionEndpointProvider interface {
	GetEndpoint() string
}

type EndpointValidator struct{}

// Validate
func (v *EndpointValidator) Validate(session any) error {
	carrier, ok := session.(SessionEndpointProvider)
	if !ok {
		return errors.New("unsupported store for endpoint validation")
	}

	if carrier.GetEndpoint() == "" {
		return ErrInvalidEndpoint
	}

	return nil
}

// NewEndpointValidator
func NewEndpointValidator() SessionStoreValidator {
	return &EndpointValidator{}
}

// NewOpenVPNCredentialsValidator
func NewNCCredentialsValidator() SessionStoreValidator {
	return NewCredentialValidator(func(username, password string) error {
		if username == "" || password == "" {
			return ErrInvalidNCCredentials
		}
		return nil
	})
}
