package main

import (
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

func buildVPNCredentialsSessionStore(
	confman config.Manager,
	errRegistry *internal.ErrorHandlingRegistry[error],
	clientAPI core.ClientAPI,
) session.SessionStore {
	return session.NewVPNCredentialsSessionStore(
		confman,
		errRegistry,
		session.NewCompositeValidator(
			session.NewExpiryValidator(),
			session.NewOpenVPNCredentialsValidator(),
			session.NewPrivateKeyValidator(),
		),
		buildVPNCredentialsSessionStoreAPIRenewalCall(confman, clientAPI),
	)
}

func buildVPNCredentialsSessionStoreAPIRenewalCall(
	confman config.Manager,
	clientAPI core.ClientAPI,
) session.VPNCredentialsRenewalAPICall {
	return func() (*session.VPNCredentialsResponse, error) {
		var cfg config.Config
		if err := confman.Load(&cfg); err != nil {
			return nil, err
		}

		data, ok := cfg.TokensData[cfg.AutoConnectData.ID]
		if !ok {
			return nil, errors.New("there is not data")
		}

		// actual API call passed into Session Store object
		return func(token string) (*session.VPNCredentialsResponse, error) {
			resp, err := clientAPI.ServiceCredentials(token)
			if err != nil {
				return nil, fmt.Errorf("getting vpn credentials data: %w", err)
			}

			return &session.VPNCredentialsResponse{
				Username:           resp.Username,
				Password:           resp.Password,
				NordLynxPrivateKey: resp.NordlynxPrivateKey,
			}, nil
		}(data.Token)
	}
}
