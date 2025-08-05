package session

import (
	"time"
)

// Common validation functions that can be reused across different session types

// ValidateExpiry checks if the provided expiry time has not passed
func ValidateExpiry(expiry time.Time) error {
	if time.Now().After(expiry) {
		return ErrSessionExpired
	}
	return nil
}

// ValidateToken checks if the token is not empty
func ValidateToken(token string) error {
	if token == "" {
		return ErrInvalidToken
	}
	return nil
}

// ValidateOwnerID checks if the owner ID is not empty
func ValidateOwnerID(ownerID string) error {
	if ownerID == "" {
		return ErrInvalidOwnerID
	}
	return nil
}

// ValidateTrustedPassOwnerID checks if the owner ID matches the expected TrustedPass owner ID
func ValidateTrustedPassOwnerID(ownerID string) error {
	if ownerID != TrustedPassOwnerID {
		return ErrInvalidOwnerId
	}
	return nil
}

// ValidateRenewToken checks if the renew token is not empty
func ValidateRenewToken(renewToken string) error {
	if renewToken == "" {
		return ErrMissingRenewToken
	}
	return nil
}

// Type-safe external validators for each session type
type AccessTokenExternalValidator func(token string, renewToken string) error
type TrustedPassExternalValidator func(token string, ownerID string) error
type CredentialsExternalValidator func(username, password string) error
