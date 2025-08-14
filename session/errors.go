package session

import (
	"errors"
)

var (
	// ErrSessionExpired indicates that the user's session has expired.
	ErrSessionExpired = errors.New("session expired")

	// ErrSessionInvalidated indicates that the session was invalidated
	ErrSessionInvalidated = errors.New("session invalidated")

	// ErrInvalidToken indicates that the token is invalid or empty.
	ErrInvalidToken = errors.New("invalid token")

	// ErrMissingAccessTokenResponse indicates that access token data is missing
	ErrMissingAccessTokenResponse = errors.New("renewal API returned nil response")

	// ErrInvalidOwnerID indicates that the owner ID is invalid or empty.
	ErrInvalidOwnerID = errors.New("invalid owner id")

	// ErrMissingTrustedPassResponse indicates that tp credentials are missing
	ErrMissingTrustedPassResponse = errors.New("tp creds renewal api returned nil or partial response")

	// ErrInvalidRenewToken indicates that the renew token has invalid format.
	ErrInvalidRenewToken = errors.New("invalid renew token")

	// ErrMissingVPNCredentials indicates that VPN credentials (username/password) are missing.
	ErrMissingVPNCredentials = errors.New("missing openvpn credentials")

	// ErrMissingNordLynxPrivateKey indicates that the NordLynx private key is missing.
	ErrMissingNordLynxPrivateKey = errors.New("missing nordlynx private key")

	// ErrMissingNCCredentials indicates that NC credentials (username/password) are missing.
	ErrMissingNCCredentials = errors.New("missing nc credentials")

	// ErrMissingNCCredentials indicates that NC credentials credentials are missing
	ErrMissingNCCredentialsResponse = errors.New("renewal API returned nil response")

	// ErrMissingEndpoint indicates that the endpoint is empty.
	ErrMissingEndpoint = errors.New("invalid endpoint")

	// ErrMissingVPNCredsResponse indicates that vpn credentials are missing
	ErrMissingVPNCredsResponse = errors.New("vpn creds renewal api returned nil or partial response")
)
