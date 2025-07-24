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

func (v *manualAccessTokenValidator) Validate(session any) error {
	tokenProvider, ok := session.(SessionTokenProvider)
	if !ok {
		return nil
	}

	expiryProvider, ok := session.(SessionExpiryProvider)
	if !ok {
		return nil
	}

	return v.ValidateFunc(tokenProvider.GetToken(), expiryProvider.GetExpiry())
}

// NewManualAccessTokenValidator
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
