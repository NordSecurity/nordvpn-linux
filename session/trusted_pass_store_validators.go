package session

import "errors"

var (
	ErrInvalidOwnerId = errors.New("invalid owner id")
)

const (
	TrustedPassOwnerID = "nordvpn"
)

type OwnerIDValidator struct{}

// Validate
func (v *OwnerIDValidator) Validate(session any) error {
	carrier, ok := session.(SessionOwnerProvider)
	if !ok {
		return errors.New("unsupported store for owner id validation")
	}

	if carrier.GetOwnerID() != TrustedPassOwnerID {
		return ErrInvalidOwnerId
	}

	return nil
}

// NewOwnerIDValidator
func NewOwnerIDValidator() SessionStoreValidator {
	return &OwnerIDValidator{}
}
