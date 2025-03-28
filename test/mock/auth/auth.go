package auth

import "github.com/NordSecurity/nordvpn-linux/auth"

// type Checker interface {
// 	// IsLoggedIn returns true when the user is logged in.
// 	IsLoggedIn() bool
// 	// IsMFAEnabled returns true if Multifactor Authentication is enabled.
// 	IsMFAEnabled() (bool, error)
// 	// IsVPNExpired is used to check whether the user is allowed to use VPN
// 	IsVPNExpired() (bool, error)
// 	// GetDedicatedIPServices returns all available server IDs, if server is not selected by the user it will set
// 	// ServerID for that service to NoServerSelected
// 	GetDedicatedIPServices() ([]DedicatedIPService, error)
// }

type AuthCheckerMock struct {
	LoggedIn    bool
	MFAEnabled  bool
	VPNExpired  bool
	DIPServices []auth.DedicatedIPService

	IsMFAEnabledErr           error
	IsVPNExpiredErr           error
	GetDedicatedIPServicesErr error
}

func (a *AuthCheckerMock) IsLoggedIn() bool { return a.LoggedIn }

func (a *AuthCheckerMock) IsMFAEnabled() (bool, error) {
	return a.MFAEnabled, a.IsMFAEnabledErr
}

func (a *AuthCheckerMock) IsVPNExpired() (bool, error) {
	return a.VPNExpired, a.IsVPNExpiredErr
}

func (a *AuthCheckerMock) GetDedicatedIPServices() ([]auth.DedicatedIPService, error) {
	return a.DIPServices, a.GetDedicatedIPServicesErr
}
