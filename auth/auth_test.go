package auth

import (
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	daemonevents "github.com/NordSecurity/nordvpn-linux/daemon/events"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
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
	resp    core.ServicesResponse
	mfaResp core.MultifactorAuthStatusResponse
	err     error
}

func (api *authAPI) Services(string) (core.ServicesResponse, error) {
	return api.resp, api.err
}

func (api *authAPI) MultifactorAuthStatus(string) (*core.MultifactorAuthStatusResponse, error) {
	return &api.mfaResp, api.err
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

type mockBoolPublisher struct {
	enabled bool
}

func (p *mockBoolPublisher) Publish(b bool) {
	p.enabled = b
}

type mockAuthPublisher struct {
}

func (p *mockAuthPublisher) Publish(events.DataAuthorization) {
}

type mockErrPublisher struct {
	err error
}

func (p *mockErrPublisher) Publish(e error) {
	p.err = e
}

func TestIsMFAEnabled(t *testing.T) {
	category.Set(t, category.Unit)

	configError := errors.New("config error")
	apiError := errors.New("api error")
	tests := []struct {
		name      string
		cm        config.Manager
		api       core.CredentialsAPI
		mfaPub    events.Publisher[bool]
		loutPub   events.Publisher[events.DataAuthorization]
		errPub    events.Publisher[error]
		isEnabled bool
		err       error
	}{
		{
			name:      "mfa enabled",
			cm:        &authConfigManager{},
			api:       &authAPI{mfaResp: core.MultifactorAuthStatusResponse{Status: internal.MFAEnabledStatusName}},
			mfaPub:    &mockBoolPublisher{},
			loutPub:   &mockAuthPublisher{},
			errPub:    &mockErrPublisher{},
			isEnabled: true,
			err:       nil,
		},
		{
			name:      "mfa disabled",
			cm:        &authConfigManager{},
			api:       &authAPI{mfaResp: core.MultifactorAuthStatusResponse{Status: "not enabled"}},
			mfaPub:    &mockBoolPublisher{},
			loutPub:   &mockAuthPublisher{},
			errPub:    &mockErrPublisher{},
			isEnabled: false,
			err:       nil,
		},
		{
			name:      "config load fails",
			cm:        &authConfigManager{loadErr: configError},
			api:       &authAPI{mfaResp: core.MultifactorAuthStatusResponse{Status: "not enabled"}},
			mfaPub:    &mockBoolPublisher{},
			loutPub:   &mockAuthPublisher{},
			errPub:    &mockErrPublisher{},
			isEnabled: false,
			err:       configError,
		},
		{
			name:      "api call fails",
			cm:        &authConfigManager{},
			api:       &authAPI{mfaResp: core.MultifactorAuthStatusResponse{Status: "not enabled"}, err: apiError},
			mfaPub:    &mockBoolPublisher{},
			loutPub:   &mockAuthPublisher{},
			errPub:    &mockErrPublisher{},
			isEnabled: false,
			err:       apiError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rc := NewRenewingChecker(test.cm, test.api, test.mfaPub, test.loutPub, test.errPub, daemonevents.NewAccountUpdateEvents())
			enabled, err := rc.isMFAEnabled()
			assert.Equal(t, test.isEnabled, enabled)

			bp, _ := test.mfaPub.(*mockBoolPublisher)
			assert.Equal(t, test.isEnabled, bp.enabled)

			assert.True(t, (test.err != nil && errors.Is(err, test.err)) || (test.err == nil && err == nil))

			ep, _ := test.errPub.(*mockErrPublisher)
			assert.True(t, (test.err != nil && errors.Is(err, test.err)) || (test.err == nil && ep.err == nil))
		})
	}
}

func TestIsVPNExpired(t *testing.T) {
	category.Set(t, category.Unit)

	testErr := errors.New("test error")
	tests := []struct {
		name          string
		cm            config.Manager
		api           core.CredentialsAPI
		accPub        *daemonevents.MockPublisherSubscriber[*pb.AccountModification]
		isExpired     bool
		subRefreshed  bool
		newExpiryDate string
		isError       bool
	}{
		{
			name: "no updates needed",
			cm:   &authConfigManager{serviceExpiry: "2990-01-01 09:18:53"},
			api:  &authAPI{},
		},
		{
			name:          "update successful",
			cm:            &authConfigManager{serviceExpiry: "1990-01-01 09:18:53"},
			api:           &authAPI{resp: []core.ServiceData{{Service: core.Service{ID: 1}, ExpiresAt: "2990-01-01 09:18:53"}}},
			subRefreshed:  true,
			newExpiryDate: "2990-01-01 09:18:53",
			accPub:        &daemonevents.MockPublisherSubscriber[*pb.AccountModification]{},
		},
		{
			name:          "expired",
			cm:            &authConfigManager{serviceExpiry: "1990-01-01 09:18:53"},
			api:           &authAPI{resp: []core.ServiceData{{Service: core.Service{ID: 1}, ExpiresAt: "1990-01-01 09:18:53"}}},
			accPub:        &daemonevents.MockPublisherSubscriber[*pb.AccountModification]{},
			subRefreshed:  true,
			newExpiryDate: "1990-01-01 09:18:53",
			isExpired:     true,
		},
		{
			name:    "config load error",
			cm:      &authConfigManager{loadErr: testErr},
			api:     &authAPI{},
			accPub:  &daemonevents.MockPublisherSubscriber[*pb.AccountModification]{},
			isError: true,
		},
		{
			name:    "config save error",
			cm:      &authConfigManager{saveErr: testErr},
			api:     &authAPI{},
			accPub:  &daemonevents.MockPublisherSubscriber[*pb.AccountModification]{},
			isError: true,
		},
		{
			name:    "api error",
			cm:      &authConfigManager{serviceExpiry: "1990-01-01 09:18:53"},
			api:     &authAPI{err: testErr},
			accPub:  &daemonevents.MockPublisherSubscriber[*pb.AccountModification]{},
			isError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rc := NewRenewingChecker(
				test.cm,
				test.api,
				&mockBoolPublisher{},
				&mockAuthPublisher{},
				&mockErrPublisher{},
				&daemonevents.AccountUpdateEvents{SubscriptionUpdate: test.accPub},
			)
			expired, err := rc.IsVPNExpired()
			if test.isError {
				assert.ErrorIs(t, err, testErr)
				assert.False(t, test.accPub.EventPublished)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.isExpired, expired)
				if test.subRefreshed {
					assert.Equal(t, test.accPub.EventPublished, test.subRefreshed)
					assert.NotNil(t, test.accPub.Event)
					assert.Equal(t, *test.accPub.Event.ExpiresAt, test.newExpiryDate)
				}
			}
		})
	}
}

func TestGetDedicatedIPServices(t *testing.T) {
	category.Set(t, category.Unit)

	dipService1ExpDate := "2050-06-04 00:00:00"
	var dipService1ServerID int64 = 11111
	dipService1 := core.ServiceData{
		ExpiresAt: dipService1ExpDate,
		Service: core.Service{
			ID: DedicatedIPServiceID,
		},
		Details: core.ServiceDetails{
			Servers: []core.ServiceServer{
				{ID: dipService1ServerID},
			},
		},
	}

	dipService2ExpDate := "2050-01-25 00:00:00"
	var dipService2ServerID int64 = 222222
	dipService2 := core.ServiceData{
		ExpiresAt: dipService2ExpDate,
		Service: core.Service{
			ID: DedicatedIPServiceID,
		},
		Details: core.ServiceDetails{
			Servers: []core.ServiceServer{
				{ID: dipService2ServerID},
			},
		},
	}

	dipService3ExpDate := "2043-05-10 00:00:00"
	dipService3 := core.ServiceData{
		ExpiresAt: dipService3ExpDate,
		Service: core.Service{
			ID: DedicatedIPServiceID,
		},
		Details: core.ServiceDetails{
			Servers: []core.ServiceServer{
				{ID: dipService1ServerID},
				{ID: dipService2ServerID},
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
				{ID: 33333},
			},
		},
	}

	dipServiceNoServerExpirationDate := "2050-08-22 00:00:00"
	dipServiceNoServer := core.ServiceData{
		ExpiresAt: dipServiceNoServerExpirationDate,
		Service: core.Service{
			ID: DedicatedIPServiceID,
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
				{ExpiresAt: dipService1ExpDate, ServerIDs: []int64{dipService1ServerID}},
			},
		},
		{
			name: "multiple dip services",
			servicesResponse: []core.ServiceData{
				dipService1,
				dipService2,
			},
			expectedDIPSerivces: []DedicatedIPService{
				{ExpiresAt: dipService1ExpDate, ServerIDs: []int64{dipService1ServerID}},
				{ExpiresAt: dipService2ExpDate, ServerIDs: []int64{dipService2ServerID}},
			},
		},
		{
			name: "multiple dip servers in single service",
			servicesResponse: []core.ServiceData{
				dipService3,
			},
			expectedDIPSerivces: []DedicatedIPService{
				{ExpiresAt: dipService3ExpDate, ServerIDs: []int64{dipService1ServerID, dipService2ServerID}},
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
				{ExpiresAt: dipService1ExpDate, ServerIDs: []int64{dipService1ServerID}},
			},
		},
		{
			name: "multiple service types",
			servicesResponse: []core.ServiceData{
				vpnService,
				unknownService,
				expiredDIPService,
				dipService1,
			},
			expectedDIPSerivces: []DedicatedIPService{
				{ExpiresAt: dipService1ExpDate, ServerIDs: []int64{dipService1ServerID}},
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
		{
			name: "no server associated with DIP service",
			servicesResponse: []core.ServiceData{
				dipServiceNoServer,
			},
			expectedDIPSerivces: []DedicatedIPService{
				{ExpiresAt: dipServiceNoServerExpirationDate, ServerIDs: []int64{}},
			},
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
