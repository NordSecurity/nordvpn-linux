package core

import "github.com/NordSecurity/nordvpn-linux/core"

type CredentialsAPIMock struct {
	NotificationCredentialsResponse core.NotificationCredentialsResponse
	NotificationCredentialsError    error
}

func (c *CredentialsAPIMock) NotificationCredentials(token, appUserID string) (core.NotificationCredentialsResponse, error) {
	return c.NotificationCredentialsResponse, c.NotificationCredentialsError
}

func (*CredentialsAPIMock) ServiceCredentials(string) (*core.CredentialsResponse, error) {
	return nil, nil
}

func (*CredentialsAPIMock) TokenRenew(string) (*core.TokenRenewResponse, error) {
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
