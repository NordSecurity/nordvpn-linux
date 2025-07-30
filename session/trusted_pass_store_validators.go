package session

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidOwnerId = errors.New("invalid owner id")
)

const (
	TrustedPassOwnerID = "nordvpn"
)

type OwnerIDValidator struct{}

// Validate checks if the session has a valid OwnerID.
func (v *OwnerIDValidator) Validate(session interface{}) error {
	carrier, ok := session.(SessionOwnerProvider)
	if !ok {
		return fmt.Errorf("unsupported store for owner id validation: got type %T", session)
	}

	if carrier.GetOwnerID() != TrustedPassOwnerID {
		return ErrInvalidOwnerId
	}

	return nil
}

// NewOwnerIDValidator returns a new OwnerIDValidator.
func NewOwnerIDValidator() SessionStoreValidator {
	return &OwnerIDValidator{}
}
