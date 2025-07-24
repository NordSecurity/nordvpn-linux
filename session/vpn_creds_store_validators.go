package session

import "errors"

var (
	ErrInvalidPrivateKey         = errors.New("invalid private key")
	ErrInvalidOpenVPNCredentials = errors.New("invalid openvpn credentials")
)

type SessionNordlynxPrivateKeyProvider interface {
	GetNordlynxPrivateKey() string
}

type PrivateKeyValidator struct{}

// Validate
func (v *PrivateKeyValidator) Validate(session any) error {
	carrier, ok := session.(SessionNordlynxPrivateKeyProvider)
	if !ok {
		return errors.New("unsupported store for nordlynx private key validation")
	}

	if carrier.GetNordlynxPrivateKey() == "" {
		return ErrInvalidPrivateKey
	}

	return nil
}

// NewOwnerIDValidator
func NewPrivateKeyValidator() SessionStoreValidator {
	return &PrivateKeyValidator{}
}

// NewOpenVPNCredentialsValidator
func NewOpenVPNCredentialsValidator() SessionStoreValidator {
	return NewCredentialValidator(func(username, password string) error {
		if username == "" || password == "" {
			return ErrInvalidOpenVPNCredentials
		}
		return nil
	})
}
