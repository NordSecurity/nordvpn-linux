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
	errRegistry *internal.ErrorHandlingRegistry[error],
	clientAPI core.RawClientAPI,
) *session.AccessTokenSessionStore {
	return session.NewAccessTokenSessionStore(
		confman,
		errRegistry,
		buildAccessTokenSessionStoreAPIRenewalCall(clientAPI),
		nil,
	)
}

func convertCoreToSessionError(err error) error {
	switch {
	case errors.Is(err, core.ErrBadRequest):
		err = session.ErrBadRequest
	case errors.Is(err, core.ErrUnauthorized):
		err = session.ErrUnauthorized
	case errors.Is(err, core.ErrNotFound):
		err = session.ErrNotFound
	}

	return err
}

func buildAccessTokenSessionStoreAPIRenewalCall(clientAPI core.RawClientAPI) session.AccessTokenRenewalAPICall {
	return func(token string, idempotencyKey uuid.UUID) (*session.AccessTokenResponse, error) {
		resp, err := clientAPI.TokenRenew(token, idempotencyKey)
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

func buildTrustedPassSessionStore(
	confman config.Manager,
	errRegistry *internal.ErrorHandlingRegistry[error],
	clientAPI core.ClientAPI,
) session.SessionStore {
	return session.NewTrustedPassSessionStore(
		confman,
		errRegistry,
		buildTrustedPassSessionStoreAPIRenewalCall(clientAPI),
		nil,
	)
}

func buildTrustedPassSessionStoreAPIRenewalCall(clientAPI core.ClientAPI) session.TrustedPassRenewalAPICall {
	return func(token string) (*session.TrustedPassAccessTokenResponse, error) {
		resp, err := clientAPI.TrustedPassToken()
		if err != nil {
			return nil, fmt.Errorf("getting trusted pass token data: %w", err)
		}

		return &session.TrustedPassAccessTokenResponse{
			Token:   resp.Token,
			OwnerID: resp.OwnerID,
		}, nil
	}
}
