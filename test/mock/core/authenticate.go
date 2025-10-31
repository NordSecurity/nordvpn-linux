package core

import "github.com/NordSecurity/nordvpn-linux/core"

type AuthenticationAPImock struct {
	URL        string
	TokenValue string
	LoginError error
	TokenError error
}

func (a *AuthenticationAPImock) Login(bool) (string, error) {
	return a.URL, a.LoginError
}

func (a *AuthenticationAPImock) Token(string) (*core.LoginResponse, error) {
	return &core.LoginResponse{Token: a.TokenValue}, a.TokenError
}
