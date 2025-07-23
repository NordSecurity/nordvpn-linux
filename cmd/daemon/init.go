package main

import (
	"errors"
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/session"
	"github.com/google/uuid"
)

func buildAccessTokenSessionStore(
	confman config.Manager,
	errRegistry *internal.ErrorHandlingRegistry[int64],
	clientAPI core.RawClientAPI,
) session.SessionStore {
	return session.NewAccessTokenSessionStore(
		confman,
		session.NewCompositeValidator(
			session.NewExpiryValidator(),
			session.NewManualAccessTokenValidator(func(token string) error {
				// this is the least expensive api call that needs authentication
				_, err := clientAPI.CurrentUser(token)
				// map to internal errors
				if errors.Is(err, core.ErrUnauthorized) {
					return session.ErrUnauthorized
				}
				return nil
			}),
		),
		errRegistry,
		func(token string, idempotencyKey uuid.UUID) (*session.AccessTokenResponse, error) {
			resp, err := clientAPI.TokenRenew(token, idempotencyKey)
			if err == nil {
				return &session.AccessTokenResponse{
					Token:      resp.Token,
					RenewToken: resp.RenewToken,
					ExpiresAt:  resp.ExpiresAt,
				}, nil
			}

			// map to internal errors
			switch {
			case errors.Is(err, core.ErrBadRequest):
				err = session.ErrBadRequest
			case errors.Is(err, core.ErrUnauthorized):
				err = session.ErrUnauthorized
			case errors.Is(err, core.ErrNotFound):
				err = session.ErrNotFound
			}

			return nil, err
		},
	)
}

func buildTrustedPassSessionStore(
	confman config.Manager,
	errRegistry *internal.ErrorHandlingRegistry[int64],
	clientAPI core.ClientAPI,
) session.SessionStore {
	return session.NewTrustedPassSessionStore(
		confman,
		errRegistry,
		session.NewCompositeValidator(session.NewExpiryValidator(), session.NewOwnerIDValidator()),
		func(token string) (*session.TrustedPassAccessTokenResponse, error) {
			resp, err := clientAPI.TrustedPassToken()
			if err != nil {
				// map to internal errors
				switch {
				case errors.Is(err, core.ErrBadRequest):
					err = session.ErrBadRequest
				case errors.Is(err, core.ErrUnauthorized):
					err = session.ErrUnauthorized
				case errors.Is(err, core.ErrNotFound):
					err = session.ErrNotFound
				}

				return nil, fmt.Errorf("getting trusted pass token data: %w", err)
			}

			return &session.TrustedPassAccessTokenResponse{
				Token:   resp.Token,
				OwnerID: resp.OwnerID,
			}, nil
		},
	)
}
