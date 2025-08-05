package session

import (
	"errors"
)

var (
	// ErrSessionExpired indicates that the user's session has expired.
	ErrSessionExpired = errors.New("session expired")

	// ErrAccessTokenExpired indicates that the access token has expired.
	ErrAccessTokenExpired = errors.New("access token expired")

	// ErrAccessTokenRevoked indicates that the access token has been revoked.
	ErrAccessTokenRevoked = errors.New("access token revoked")

	// ErrInvalidToken indicates that the token is invalid or empty.
	ErrInvalidToken = errors.New("invalid token")

	// ErrInvalidOwnerID indicates that the owner ID is invalid or empty.
	ErrInvalidOwnerID = errors.New("invalid owner ID")

	// ErrMissingRenewToken indicates that the renew token is missing.
	ErrMissingRenewToken = errors.New("missing renew token")

	// ErrInvalidOwnerId indicates that the owner id does not match expected value.
	ErrInvalidOwnerId = errors.New("invalid owner id")
)
