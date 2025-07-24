package session

import (
	"errors"
	"fmt"
	"time"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

type AnyIdempotentAPICallWithToken func(token string) error

type manualAccessTokenValidator struct {
	ValidateFunc func(token string, expiryDate time.Time) error
}

// Validate checks if the session contains a valid manual access token.
func (v *manualAccessTokenValidator) Validate(session any) error {
	tokenProvider, ok := session.(SessionTokenProvider)
	if !ok {
		return errors.New("unsupported store for token validation")
	}

	expiryProvider, ok := session.(SessionExpiryProvider)
	if !ok {
		return errors.New("unsupported store for expiry validation")
	}

	return v.ValidateFunc(tokenProvider.GetToken(), expiryProvider.GetExpiry())
}

// NewManualAccessTokenValidator creates a validator that verifies manually issued access tokens.
func NewManualAccessTokenValidator(api AnyIdempotentAPICallWithToken) SessionStoreValidator {
	return &manualAccessTokenValidator{
		ValidateFunc: func(token string, expiryDate time.Time) error {
			if expiryDate.Equal(ManualAccessTokenExpiryDate) {
				// here we have a manually issuesed access token
				var isFormatValid = internal.AccessTokenFormatValidator
				if !isFormatValid(token) {
					return fmt.Errorf("invalid access token format: %w", ErrAccessTokenExpired)
				}

				err := api(token)
				// the only interest is in whether we receive "unauthorized access" error
				if errors.Is(err, ErrUnauthorized) {
					return ErrAccessTokenRevoked
				}

				return nil
			}
			return nil
		},
	}
}
