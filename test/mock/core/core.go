package core

import (
	"github.com/NordSecurity/nordvpn-linux/core"
)

type CredentialsAPIMock struct {
	NotificationCredentialsResponse       core.NotificationCredentialsResponse
	NotificationCredentialsRevokeResponse core.NotificationCredentialsRevokeResponse
	NotificationCredentialsError          error
}

func (c *CredentialsAPIMock) NotificationCredentials(token, appUserID string) (core.NotificationCredentialsResponse, error) {
	return c.NotificationCredentialsResponse, c.NotificationCredentialsError
}

func (c *CredentialsAPIMock) NotificationCredentialsRevoke(token, appUserID string, purgeSession bool) (core.NotificationCredentialsRevokeResponse, error) {
	return c.NotificationCredentialsRevokeResponse, nil
}

func (*CredentialsAPIMock) ServiceCredentials(string) (*core.CredentialsResponse, error) {
	return nil, nil
}

func (*CredentialsAPIMock) TokenRenew(string) (*core.TokenRenewResponse, error) {
	return nil, nil
}

func (*CredentialsAPIMock) MultifactorAuthStatus(string) (*core.MultifactorAuthStatusResponse, error) {
	return nil, nil
}

func (*CredentialsAPIMock) Services(string) (core.ServicesResponse, error) {
	return core.ServicesResponse{}, nil
}

func (*CredentialsAPIMock) CurrentUser(string) (*core.CurrentUserResponse, error) {
	return nil, nil
}

func (*CredentialsAPIMock) DeleteToken(string) error {
	return nil
}

func (*CredentialsAPIMock) TrustedPassToken(string) (*core.TrustedPassTokenResponse, error) {
	return nil, nil
}

func (*CredentialsAPIMock) Logout(string) error {
	return nil
}
