package session

import (
	"fmt"
	"time"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

type AnyIdempotentAPICallWithToken func(token string) error

type manualAccessTokenValidator struct {
	ValidateFunc func(token string, expiryDate time.Time) error
}

// Validate checks if the session contains a valid manual access token.
func (v *manualAccessTokenValidator) Validate(session interface{}) error {
	tokenProvider, ok := session.(SessionTokenProvider)
	if !ok {
		return fmt.Errorf("unsupported store for token validation: got type %T", session)
	}

	expiryProvider, ok := session.(SessionExpiryProvider)
	if !ok {
		return fmt.Errorf("unsupported store for expiry validation: got type %T", session)
	}

	return v.ValidateFunc(tokenProvider.GetToken(), expiryProvider.GetExpiry())
}

// NewManualAccessTokenValidator creates a validator that verifies manually issued access tokens.
// api should only return an error if API result indicates unauthorized access, otherwise - nil
func NewManualAccessTokenValidator(api AnyIdempotentAPICallWithToken) SessionStoreValidator {
	return &manualAccessTokenValidator{
		ValidateFunc: func(token string, expiryDate time.Time) error {
			if expiryDate.Equal(ManualAccessTokenExpiryDate) {
				// here we have a manually issued access token
				var isFormatValid = internal.AccessTokenFormatValidator
				if !isFormatValid(token) {
					return fmt.Errorf("invalid access token format: %w", ErrAccessTokenExpired)
				}

				return api(token)
			}

			return nil
		},
	}
}
