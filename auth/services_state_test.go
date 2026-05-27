package auth

import (
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/core"
	devicekey "github.com/NordSecurity/nordvpn-linux/device_key"
	"github.com/NordSecurity/nordvpn-linux/internal/caching"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	coremock "github.com/NordSecurity/nordvpn-linux/test/mock/core"
	testdevicekey "github.com/NordSecurity/nordvpn-linux/test/mock/devicekey"
	"github.com/stretchr/testify/assert"
)

func TestNotifyUserServicesChanged(t *testing.T) {
	category.Set(t, category.Unit)

	oldServiceResponse := core.ServicesResponse{
		core.ServiceData{ID: 1, ExpiresAt: "2026-07-14"},
	}

	newServiceResponse := core.ServicesResponse{
		core.ServiceData{ID: 2, ExpiresAt: "2027-04-02"},
	}

	mockApi := coremock.CredentialsAPIMock{}
	mockApi.ServicesResponse = newServiceResponse

	cache := caching.NewCacheWithTTL(time.Hour*5, mockApi.Services)
	cache.Set(oldServiceResponse)

	servicesState := ServicesState{
		cache:                     cache,
		expChecker:                newMockExpirationChecker(),
		dedicatedServerKeyManager: &testdevicekey.MockDeviceKeyManager{},
	}

	err := servicesState.NotifyUserServicesChanged(struct{}{})
	assert.NoError(t, err, "Unexpected error returned by NotifyUserServicesChanged.")

	services, _ := servicesState.fetchServices()
	assert.Equal(t, newServiceResponse, services, "Services should be updated after services change notification.")
}

func TestNotifyUserServicesChanged_DedicatedServersHandling(t *testing.T) {
	category.Set(t, category.Unit)

	dedicatedServersServiceResponse := core.ServicesResponse{
		core.ServiceData{ID: 2, ExpiresAt: "2027-04-02", Service: core.Service{ID: DedicatedServersServiceID}},
	}

	mockApi := coremock.CredentialsAPIMock{}
	mockApi.ServicesResponse = dedicatedServersServiceResponse

	dedicatedServersKeyManagerMock := testdevicekey.MockDeviceKeyManager{
		DedicatedServerRegistrationData: &devicekey.DedicatedServersConnectionData{},
	}

	cache := caching.NewCacheWithTTL(time.Hour*5, mockApi.Services)

	expiredDate := "2027-05-02"
	servicesState := ServicesState{
		cache:                     cache,
		expChecker:                newMockExpirationChecker(expiredDate),
		dedicatedServerKeyManager: &dedicatedServersKeyManagerMock,
	}

	err := servicesState.NotifyUserServicesChanged(struct{}{})
	assert.NoError(t, err, "Unexpected error returned by NotifyUserServicesChanged.")
	assert.True(t, dedicatedServersKeyManagerMock.WasKeyRegistered,
		"Key should be registered if dedicated servers service is present in the service response.")

	dedicatedServersKeyManagerMock.WasKeyRegistered = false

	noDedicatedServersServiceResponse := core.ServicesResponse{core.ServiceData{ID: 3, ExpiresAt: "2027-04-02"}}
	mockApi.ServicesResponse = noDedicatedServersServiceResponse

	err = servicesState.NotifyUserServicesChanged(struct{}{})
	assert.NoError(t, err, "Unexpected error returned by NotifyUserServicesChanged.")
	assert.False(t, dedicatedServersKeyManagerMock.WasKeyRegistered,
		"Key should not be registered if dedicated servers service is not present in the service response.")

	expiredDedicatedServersServiceResponse := core.ServicesResponse{
		core.ServiceData{ID: 2, ExpiresAt: expiredDate, Service: core.Service{ID: DedicatedServersServiceID}},
	}
	mockApi.ServicesResponse = expiredDedicatedServersServiceResponse

	err = servicesState.NotifyUserServicesChanged(struct{}{})
	assert.NoError(t, err, "Unexpected error returned by NotifyUserServicesChanged.")
	assert.False(t, dedicatedServersKeyManagerMock.WasKeyRegistered,
		"Key should not be registered if dedicated servers service in the service response is expired.")
}
