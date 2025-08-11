package session

import (
	"time"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

// Common validation functions that can be reused across different session types

// ValidateExpiry checks if the provided expiry time has not passed
func ValidateExpiry(expiry time.Time) error {
	if time.Now().After(expiry) {
		return ErrSessionExpired
	}
	return nil
}

// ValidateAccessTokenFormat checks if the access token has valid format (hexadecimal characters only)
func ValidateAccessTokenFormat(token string) error {
	if !internal.AccessTokenFormatValidatorFunc(token) {
		return ErrInvalidToken
	}
	return nil
}

// ValidateTrustedPassTokenFormat checks if the TrustedPass token is not empty and has valid format
func ValidateTrustedPassTokenFormat(token string) error {
	if !internal.TrustedPassTokenFormatValidatorFunc(token) {
		return ErrInvalidToken
	}
	return nil
}

// ValidateTrustedPassOwnerID checks if the owner ID matches the expected TrustedPass owner ID
func ValidateTrustedPassOwnerID(ownerID string) error {
	if ownerID != TrustedPassOwnerID {
		return ErrInvalidOwnerID
	}
	return nil
}

// ValidateRenewToken checks if the renew token is not empty and has valid format (hexadecimal characters only)
func ValidateRenewToken(renewToken string) error {
	if !internal.RenewalTokenFormatValidatorFunc(renewToken) {
		return ErrInvalidRenewToken
	}
	return nil
}

// Type-safe external validators for each session type
type TrustedPassExternalValidator func(token string, ownerID string) error
type CredentialsExternalValidator func(username, password string) error
