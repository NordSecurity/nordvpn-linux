package core

import (
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/google/uuid"
)

type CredentialsAPIMock struct {
	NotificationCredentialsResponse       core.NotificationCredentialsResponse
	NotificationCredentialsRevokeResponse core.NotificationCredentialsRevokeResponse
	NotificationCredentialsError          error

	CurrentUserResponse core.CurrentUserResponse
	CurrentUserErr      error
}

func (c *CredentialsAPIMock) NotificationCredentials(appUserID string) (core.NotificationCredentialsResponse, error) {
	return c.NotificationCredentialsResponse, c.NotificationCredentialsError
}

func (c *CredentialsAPIMock) NotificationCredentialsRevoke(appUserID string, purgeSession bool) (core.NotificationCredentialsRevokeResponse, error) {
	return c.NotificationCredentialsRevokeResponse, nil
}

func (*CredentialsAPIMock) ServiceCredentials(token string) (*core.CredentialsResponse, error) {
	return nil, nil
}

func (*CredentialsAPIMock) TokenRenew(renewalToken string, idempotencyKey uuid.UUID) (*core.TokenRenewResponse, error) {
	return nil, nil
}

func (*CredentialsAPIMock) MultifactorAuthStatus() (*core.MultifactorAuthStatusResponse, error) {
	return nil, nil
}

func (*CredentialsAPIMock) Services() (core.ServicesResponse, error) {
	return core.ServicesResponse{}, nil
}

func (c *CredentialsAPIMock) CurrentUser() (*core.CurrentUserResponse, error) {
	return &c.CurrentUserResponse, c.CurrentUserErr
}

func (*CredentialsAPIMock) DeleteToken() error {
	return nil
}

func (*CredentialsAPIMock) TrustedPassToken() (*core.TrustedPassTokenResponse, error) {
	return nil, nil
}

func (*CredentialsAPIMock) Logout() error {
	return nil
}

type MockTokenManager struct {
	core.TokenManager
}
