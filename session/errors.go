package session

import (
	"errors"
)

var (
	// ErrSessionExpired indicates that the user's session has expired.
	ErrSessionExpired = errors.New("session expired")

	// ErrInvalidToken indicates that the token is invalid or empty.
	ErrInvalidToken = errors.New("invalid token")

	// ErrInvalidOwnerID indicates that the owner ID is invalid or empty.
	ErrInvalidOwnerID = errors.New("invalid owner id")

	// ErrInvalidRenewToken indicates that the renew token has invalid format.
	ErrInvalidRenewToken = errors.New("invalid renew token")

	// ErrMissingVPNCredentials indicates that VPN credentials (username/password) are missing.
	ErrMissingVPNCredentials = errors.New("missing openvpn credentials")

	// ErrMissingNordLynxPrivateKey indicates that the NordLynx private key is missing.
	ErrMissingNordLynxPrivateKey = errors.New("missing nordlynx private key")
)
