package auth

import "github.com/NordSecurity/nordvpn-linux/auth"

type AuthCheckerMock struct {
	LoggedIn               bool
	MFAEnabled             bool
	VPNExpired             bool
	DIPServices            []auth.DedicatedIPService
	DedicatedServerService auth.DedicatedServersService

	IsMFAEnabledErr              error
	IsVPNExpiredErr              error
	GetDedicatedIPServicesErr    error
	GetDedicatedServerServiceErr error
}

func (a *AuthCheckerMock) IsLoggedIn() (bool, error) {
	return a.LoggedIn, nil
}

func (a *AuthCheckerMock) IsMFAEnabled() (bool, error) {
	return a.MFAEnabled, a.IsMFAEnabledErr
}

func (a *AuthCheckerMock) IsVPNExpired() (bool, error) {
	return a.VPNExpired, a.IsVPNExpiredErr
}

func (a *AuthCheckerMock) GetDedicatedIPServices() ([]auth.DedicatedIPService, error) {
	return a.DIPServices, a.GetDedicatedIPServicesErr
}

func (a *AuthCheckerMock) GetDedicatedServersService() (auth.DedicatedServersService, error) {
	return a.DedicatedServerService, a.GetDedicatedServerServiceErr
}
