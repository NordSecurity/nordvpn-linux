package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/session"
	"github.com/google/uuid"
)

// SessionStoresBuilder provides a fluent interface for building session stores
type SessionStoresBuilder struct {
	confman    config.Manager
	stores     sessionStores
	registries errorRegistries
}

// sessionStores holds all session store instances
type sessionStores struct {
	accessToken *session.AccessTokenSessionStore
	vpnCreds    session.SessionStore
	trustedPass session.SessionStore
	ncCreds     session.SessionStore
}

// errorRegistries holds all error handling registries
type errorRegistries struct {
	accessToken *internal.ErrorHandlingRegistry[error]
	vpnCreds    *internal.ErrorHandlingRegistry[error]
	trustedPass *internal.ErrorHandlingRegistry[error]
	ncCreds     *internal.ErrorHandlingRegistry[error]
}

// NewSessionStoresBuilder creates a new builder for session stores
func NewSessionStoresBuilder(confman config.Manager) *SessionStoresBuilder {
	return &SessionStoresBuilder{
		confman: confman,
		registries: errorRegistries{
			accessToken: internal.NewErrorHandlingRegistry[error](),
			vpnCreds:    internal.NewErrorHandlingRegistry[error](),
			trustedPass: internal.NewErrorHandlingRegistry[error](),
			ncCreds:     internal.NewErrorHandlingRegistry[error](),
		},
	}
}

// BuildAccessTokenStore builds and returns the access token session store
func (b *SessionStoresBuilder) BuildAccessTokenStore(rawClientAPI core.RawClientAPI) *session.AccessTokenSessionStore {
	b.stores.accessToken = session.NewAccessTokenSessionStore(
		b.confman,
		b.registries.accessToken,
		renewAccessToken(rawClientAPI),
		nil,
	)
	return b.stores.accessToken
}

// BuildVPNCredsStore builds and returns the VPN credentials session store
func (b *SessionStoresBuilder) BuildVPNCredsStore(clientAPI core.ClientAPI) session.SessionStore {
	b.stores.vpnCreds = session.NewVPNCredentialsSessionStore(
		b.confman,
		b.registries.vpnCreds,
		renewVPNCredentials(b.confman, clientAPI),
		nil,
	)
	return b.stores.vpnCreds
}

// BuildTrustedPassStore builds and returns the trusted pass session store
func (b *SessionStoresBuilder) BuildTrustedPassStore(clientAPI core.ClientAPI) session.SessionStore {
	b.stores.trustedPass = session.NewTrustedPassSessionStore(
		b.confman,
		b.registries.trustedPass,
		renewTrustedPass(clientAPI),
		nil,
	)
	return b.stores.trustedPass
}

// BuildNCCredsStore builds and returns the NC credentials session store
func (b *SessionStoresBuilder) BuildNCCredsStore(clientAPI core.ClientAPI) session.SessionStore {
	b.stores.ncCreds = session.NewNCCredentialsSessionStore(
		b.confman,
		b.registries.ncCreds,
		renewNCCredentials(b.confman, clientAPI),
		nil,
	)
	return b.stores.ncCreds
}

// ConfigureErrorHandlers sets up error handlers for all session stores
func (b *SessionStoresBuilder) ConfigureErrorHandlers(logoutHandler *daemon.LogoutHandler) {
	b.registerAccessTokenHandlers(logoutHandler)
	b.registerVPNCredsHandlers(logoutHandler)
	b.registerTrustedPassHandlers(logoutHandler)
	b.registerNCCredsHandlers(logoutHandler)
}

// GetStores returns all built stores as a slice
func (b *SessionStoresBuilder) GetStores() []session.SessionStore {
	return []session.SessionStore{
		b.stores.accessToken,
		b.stores.ncCreds,
		b.stores.trustedPass,
		b.stores.vpnCreds,
	}
}

// registerAccessTokenHandlers configures access token error handlers
func (b *SessionStoresBuilder) registerAccessTokenHandlers(logoutHandler *daemon.LogoutHandler) {
	handlers := map[error]events.ReasonCode{
		core.ErrNotFound:                      events.ReasonTokenMissing,
		core.ErrBadRequest:                    events.ReasonAuthTokenBad,
		session.ErrInvalidRenewToken:          events.ReasonTokenCorrupted,
		session.ErrSessionInvalidated:         events.ReasonAuthTokenInvalidated,
		session.ErrMissingAccessTokenResponse: events.ReasonTokenMissing,
		core.ErrUnauthorized:                  events.ReasonNone, // no dedicated exception code
	}

	errs := []error{
		core.ErrBadRequest,
		core.ErrNotFound,
		session.ErrInvalidRenewToken,
		session.ErrSessionInvalidated,
		core.ErrUnauthorized,
		session.ErrMissingAccessTokenResponse,
	}

	logoutHandler.Register(b.registries.accessToken, errs, func(reason error) events.ReasonCode {
		for err, code := range handlers {
			if errors.Is(reason, err) {
				return code
			}
		}
		return events.ReasonNone
	})
}

// registerVPNCredsHandlers configures VPN credentials error handlers
func (b *SessionStoresBuilder) registerVPNCredsHandlers(logoutHandler *daemon.LogoutHandler) {
	errs := []error{
		core.ErrUnauthorized,
		core.ErrBadRequest,
		session.ErrMissingVPNCredsResponse,
		session.ErrMissingVPNCredentials,
		session.ErrMissingNordLynxPrivateKey,
	}

	logoutHandler.Register(b.registries.vpnCreds, errs, func(reason error) events.ReasonCode {
		// For VPN credential errors that indicate corruption, try access token renewal first
		if errors.Is(reason, session.ErrMissingVPNCredsResponse) ||
			errors.Is(reason, session.ErrMissingVPNCredentials) ||
			errors.Is(reason, session.ErrMissingNordLynxPrivateKey) {
			// Attempt to renew access token silently to avoid duplicate logout calls
			// We use SilentRenewal() to prevent the access token error handler from triggering
			if err := b.stores.accessToken.Renew(session.SilentRenewal(), session.ForceRenewal()); err != nil {
				switch {
				case errors.Is(err, core.ErrBadRequest):
					return events.ReasonCorruptedVPNCredsAuthBad
				case errors.Is(err, core.ErrNotFound):
					return events.ReasonCorruptedVPNCredsAuthMissing
				default:
					return events.ReasonNone
				}
			}

			// Access token renewal succeeded
			if errors.Is(reason, session.ErrMissingVPNCredsResponse) {
				return events.ReasonCorruptedVPNCreds
			}

			// For missing credentials or missing nordlynx key, return corrupted VPN creds
			return events.ReasonCorruptedVPNCreds
		}

		return events.ReasonNone
	})
}

// registerTrustedPassHandlers configures trusted pass error handlers
func (b *SessionStoresBuilder) registerTrustedPassHandlers(logoutHandler *daemon.LogoutHandler) {
	errs := []error{
		core.ErrBadRequest,
		core.ErrUnauthorized,
		core.ErrNotFound,
		session.ErrInvalidToken,
		session.ErrInvalidOwnerID,
		session.ErrMissingTrustedPassResponse,
	}

	logoutHandler.Register(b.registries.trustedPass, errs, alwaysReasonNone)
}

// registerNCCredsHandlers configures NC credentials error handlers
func (b *SessionStoresBuilder) registerNCCredsHandlers(logoutHandler *daemon.LogoutHandler) {
	errs := []error{
		core.ErrBadRequest,
		core.ErrUnauthorized,
		session.ErrMissingNCCredentials,
		session.ErrMissingEndpoint,
		session.ErrMissingNCCredentialsResponse,
	}

	logoutHandler.Register(b.registries.ncCreds, errs, alwaysReasonNone)
}

// alwaysReasonNone is a helper function that always returns ReasonNone
func alwaysReasonNone(error) events.ReasonCode {
	return events.ReasonNone
}

// Renewal functions

// getTokenData retrieves token data from the configuration
func getTokenData(confman config.Manager) (*config.TokenData, error) {
	var cfg config.Config
	if err := confman.Load(&cfg); err != nil {
		return nil, err
	}

	data, ok := cfg.TokensData[cfg.AutoConnectData.ID]
	if !ok {
		return nil, errors.New("token data not found")
	}

	return &data, nil
}

// renewAccessToken creates an access token renewal function
func renewAccessToken(api core.RawClientAPI) session.AccessTokenRenewalAPICall {
	return func(token string, idempotencyKey uuid.UUID) (*session.AccessTokenResponse, error) {
		resp, err := api.TokenRenew(token, idempotencyKey)
		if err != nil {
			return nil, fmt.Errorf("renewing access token: %w", err)
		}

		return &session.AccessTokenResponse{
			Token:      resp.Token,
			RenewToken: resp.RenewToken,
			ExpiresAt:  resp.ExpiresAt,
		}, nil
	}
}

// renewTrustedPass creates a trusted pass renewal function
func renewTrustedPass(api core.ClientAPI) session.TrustedPassRenewalAPICall {
	return func(token string) (*session.TrustedPassAccessTokenResponse, error) {
		resp, err := api.TrustedPassToken()
		if err != nil {
			return nil, fmt.Errorf("getting trusted pass token data: %w", err)
		}

		return &session.TrustedPassAccessTokenResponse{
			Token:   resp.Token,
			OwnerID: resp.OwnerID,
		}, nil
	}
}

// renewVPNCredentials creates a VPN credentials renewal function
func renewVPNCredentials(confman config.Manager, api core.ClientAPI) session.VPNCredentialsRenewalAPICall {
	return func() (*session.VPNCredentialsResponse, error) {
		data, err := getTokenData(confman)
		if err != nil {
			return nil, err
		}

		resp, err := api.ServiceCredentials(data.Token)
		if err != nil {
			return nil, fmt.Errorf("getting vpn credentials data: %w", err)
		}

		return &session.VPNCredentialsResponse{
			Username:           resp.Username,
			Password:           resp.Password,
			NordLynxPrivateKey: resp.NordlynxPrivateKey,
		}, nil
	}
}

// renewNCCredentials creates a NC credentials renewal function
func renewNCCredentials(confman config.Manager, api core.ClientAPI) session.NCCredentialsRenewalAPICall {
	return func() (*session.NCCredentialsResponse, error) {
		data, err := getTokenData(confman)
		if err != nil {
			return nil, err
		}

		resp, err := api.NotificationCredentials(data.NCData.UserID.String())
		if err != nil {
			return nil, fmt.Errorf("getting nc credentials data: %w", err)
		}

		return &session.NCCredentialsResponse{
			Username:  resp.Username,
			Password:  resp.Password,
			Endpoint:  resp.Endpoint,
			ExpiresIn: time.Duration(resp.ExpiresIn),
		}, nil
	}
}
