package auth

import (
	"errors"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestIsTokenExpired(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		input    string
		expected bool
	}{
		{
			input:    "",
			expected: true,
		},
		{
			input:    "1990-01-01 09:18:53",
			expected: true,
		},
		{
			input:    "2990-01-01 09:18:53",
			expected: false,
		},
		{
			input:    "Wed Sep 18 09:27:12 UTC 2019",
			expected: true,
		},
	}

	for _, tt := range tests {
		got := isTokenExpired(tt.input)
		assert.Equal(t, tt.expected, got)
	}
}

type authConfigManager struct {
	config.Manager
	serviceExpiry string
	loadErr       error
	saveErr       error
}

func (cm *authConfigManager) Load(c *config.Config) error {
	*c = config.Config{
		AutoConnectData: config.AutoConnectData{ID: 1},
		TokensData: map[int64]config.TokenData{
			1: {ServiceExpiry: cm.serviceExpiry},
		},
	}
	return cm.loadErr
}

func (cm *authConfigManager) SaveWith(config.SaveFunc) error {
	return cm.saveErr
}

type authAPI struct {
	core.CredentialsAPI
	resp core.ServicesResponse
	err  error
}

func (api *authAPI) Services(string) (core.ServicesResponse, error) {
	return api.resp, api.err
}

func TestIsVPNExpired(t *testing.T) {
	category.Set(t, category.Unit)

	testErr := errors.New("test error")
	tests := []struct {
		name      string
		cm        config.Manager
		api       core.CredentialsAPI
		isExpired bool
		isError   bool
	}{
		{
			name: "no updates needed",
			cm:   &authConfigManager{serviceExpiry: "2990-01-01 09:18:53"},
			api:  &authAPI{},
		},
		{
			name: "update successful",
			cm:   &authConfigManager{serviceExpiry: "1990-01-01 09:18:53"},
			api:  &authAPI{resp: []config.ServiceData{{Service: config.Service{ID: 1}, ExpiresAt: "2990-01-01 09:18:53"}}},
		},
		{
			name:      "expired",
			cm:        &authConfigManager{serviceExpiry: "1990-01-01 09:18:53"},
			api:       &authAPI{resp: []config.ServiceData{{Service: config.Service{ID: 1}, ExpiresAt: "1990-01-01 09:18:53"}}},
			isExpired: true,
		},
		{
			name:    "config load error",
			cm:      &authConfigManager{loadErr: testErr},
			api:     &authAPI{},
			isError: true,
		},
		{
			name:    "config save error",
			cm:      &authConfigManager{saveErr: testErr},
			api:     &authAPI{},
			isError: true,
		},
		{
			name:    "api error",
			cm:      &authConfigManager{serviceExpiry: "1990-01-01 09:18:53"},
			api:     &authAPI{err: testErr},
			isError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rc := NewRenewingChecker(test.cm, test.api)
			expired, err := rc.IsVPNExpired()
			if test.isError {
				assert.ErrorIs(t, err, testErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.isExpired, expired)
			}
		})
	}
}
