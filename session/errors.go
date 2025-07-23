package session

import (
	"errors"
)

// generic errors
var (
	// ErrBadRequest represents a "Bad Request" error.
	ErrBadRequest = errors.New("bad request")

	// ErrUnauthorized represents an "Unauthorized" error.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden represents a "Forbidden" error.
	ErrForbidden = errors.New("forbidden")

	// ErrNotFound represents a "Not Found" error.
	ErrNotFound = errors.New("not found")

	// ErrSessionExpired indicates that the user's session has expired.
	ErrSessionExpired = errors.New("session expired")
)

// specific errors
var (
	// ErrAccessTokenExpired indicates that the access token has expired.
	ErrAccessTokenExpired = errors.New("access token expired")

	// ErrAccessTokenRevoked indicates that the access token has been revoked.
	ErrAccessTokenRevoked = errors.New("access token revoked")
)
