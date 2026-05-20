package auth

import (
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/internal/caching"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	coremock "github.com/NordSecurity/nordvpn-linux/test/mock/core"
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
		cache: cache,
	}

	err := servicesState.NotifyUserServicesChanged(struct{}{})
	assert.NoError(t, err, "Unexpected error returned by NotifyUserServicesChanged.")

	services, _ := servicesState.fetchServices()
	assert.Equal(t, newServiceResponse, services, "Services should be updated after services change notification.")
}
