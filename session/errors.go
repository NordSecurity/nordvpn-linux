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
)
