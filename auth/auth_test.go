package auth

import (
	"errors"
	"fmt"
	"slices"
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
		expirationChecker := systemTimeExpirationChecker{}
		got := expirationChecker.isExpired(tt.input)
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

type mockExpirationChecker struct {
	expiredDates []string
}

func newMockExpirationChecker(expiredDates ...string) mockExpirationChecker {
	return mockExpirationChecker{
		expiredDates: expiredDates,
	}
}

func (m mockExpirationChecker) isExpired(expiryTime string) bool {
	if idx := slices.Index(m.expiredDates, expiryTime); idx != -1 {
		return true
	}
	return false
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
			api:  &authAPI{resp: []core.ServiceData{{Service: core.Service{ID: 1}, ExpiresAt: "2990-01-01 09:18:53"}}},
		},
		{
			name:      "expired",
			cm:        &authConfigManager{serviceExpiry: "1990-01-01 09:18:53"},
			api:       &authAPI{resp: []core.ServiceData{{Service: core.Service{ID: 1}, ExpiresAt: "1990-01-01 09:18:53"}}},
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

func TestGetDedicatedIPServices(t *testing.T) {
	category.Set(t, category.Unit)

	dipService1ExpDate := "2050-06-04 00:00:00"
	var dipSercice1ServerID int64 = 11111
	dipService1 := core.ServiceData{
		ExpiresAt: dipService1ExpDate,
		Service: core.Service{
			ID: DedicatedIPServiceID,
		},
		Details: core.ServiceDetails{
			Servers: []core.ServiceServer{
				{ID: dipSercice1ServerID},
			},
		},
	}

	dipService2ExpDate := "2050-08-22 00:00:00"
	var dipSercice2ServerID int64 = 11111
	dipService2 := core.ServiceData{
		ExpiresAt: dipService2ExpDate,
		Service: core.Service{
			ID: DedicatedIPServiceID,
		},
		Details: core.ServiceDetails{
			Servers: []core.ServiceServer{
				core.ServiceServer{ID: dipSercice2ServerID},
			},
		},
	}

	expiredDate := "2023-08-22 00:00:00"
	expiredDIPService := core.ServiceData{
		ExpiresAt: expiredDate,
		Service: core.Service{
			ID: DedicatedIPServiceID,
		},
		Details: core.ServiceDetails{
			Servers: []core.ServiceServer{
				core.ServiceServer{ID: 33333},
			},
		},
	}

	vpnService := core.ServiceData{
		ExpiresAt: "2050-08-22 00:00:00",
		Service: core.Service{
			ID: VPNServiceID,
		},
	}

	unknownService := core.ServiceData{
		ExpiresAt: "2050-08-22 00:00:00",
		Service: core.Service{
			ID: 1111,
		},
	}

	expirationChecker := newMockExpirationChecker(expiredDate)

	test := []struct {
		name                string
		servicesResponse    []core.ServiceData
		servicesErr         error
		configLoadErr       error
		expectedDIPSerivces []DedicatedIPService
		shouldBeErr         bool
	}{
		{
			name: "single dip service",
			servicesResponse: []core.ServiceData{
				dipService1,
			},
			expectedDIPSerivces: []DedicatedIPService{
				{ExpiresAt: dipService1ExpDate, ServerID: dipSercice1ServerID},
			},
		},
		{
			name: "multiple dip services",
			servicesResponse: []core.ServiceData{
				dipService1,
				dipService2,
			},
			expectedDIPSerivces: []DedicatedIPService{
				{ExpiresAt: dipService1ExpDate, ServerID: dipSercice1ServerID},
				{ExpiresAt: dipService2ExpDate, ServerID: dipSercice2ServerID},
			},
		},
		{
			name: "only expired dip services",
			servicesResponse: []core.ServiceData{
				expiredDIPService,
			},
			expectedDIPSerivces: []DedicatedIPService{},
		},
		{
			name: "expired and unexpired dip services",
			servicesResponse: []core.ServiceData{
				expiredDIPService,
				dipService1,
			},
			expectedDIPSerivces: []DedicatedIPService{
				{ExpiresAt: dipService1ExpDate, ServerID: dipSercice1ServerID},
			},
		},
		{
			name: "mutliple service types",
			servicesResponse: []core.ServiceData{
				vpnService,
				unknownService,
				expiredDIPService,
				dipService1,
			},
			expectedDIPSerivces: []DedicatedIPService{
				{ExpiresAt: dipService1ExpDate, ServerID: dipSercice1ServerID},
			},
		},
		{
			name: "no dip services",
			servicesResponse: []core.ServiceData{
				unknownService,
			},
			expectedDIPSerivces: []DedicatedIPService{},
		},
		{
			name:                "fetch services error",
			servicesErr:         fmt.Errorf("failed to fetch new services"),
			expectedDIPSerivces: []DedicatedIPService{},
			shouldBeErr:         true,
		},
		{
			name:                "config error",
			configLoadErr:       fmt.Errorf("config load error"),
			expectedDIPSerivces: []DedicatedIPService{},
			shouldBeErr:         true,
		},
	}

	for _, test := range test {
		t.Run(test.name, func(t *testing.T) {
			mockAPI := authAPI{
				resp: test.servicesResponse,
				err:  test.servicesErr,
			}

			configMock := authConfigManager{
				loadErr: test.configLoadErr,
			}

			rc := RenewingChecker{
				cm:         &configMock,
				creds:      &mockAPI,
				expChecker: expirationChecker,
			}

			dipServices, err := rc.GetDedicatedIPServices()
			if test.shouldBeErr {
				assert.NotNil(t, err, "GetDedicatedIPServices didn't return an error when errror was expected.")
				return
			}
			assert.Equal(t, test.expectedDIPSerivces, dipServices,
				"Invalid services returned by GetDedicatedIPServices.")
		})
	}
}
