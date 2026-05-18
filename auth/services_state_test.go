package auth

import (
	"errors"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	coremock "github.com/NordSecurity/nordvpn-linux/test/mock/core"
	"github.com/stretchr/testify/assert"
)

func TestFetchServices(t *testing.T) {
	category.Set(t, category.Unit)

	servicesResponseA := core.ServicesResponse{
		core.ServiceData{ID: 1, ExpiresAt: "2026-07-14"},
	}

	servicesResponseB := core.ServicesResponse{
		core.ServiceData{ID: 2, ExpiresAt: "2026-10-27"},
	}

	tests := []struct {
		name                    string
		hasFetchedServiceData   bool
		storedServiceData       core.ServicesResponse
		serviceResponse         core.ServicesResponse
		fetchServiceErr         error
		expectedServiceResponse core.ServicesResponse
		expectedStatus          bool
		shouldReturnError       bool
	}{
		{
			name:                    "fetch new service",
			serviceResponse:         servicesResponseA,
			expectedServiceResponse: servicesResponseA,
			expectedStatus:          true,
		},
		{
			name:                    "services already stored",
			hasFetchedServiceData:   true,
			storedServiceData:       servicesResponseA,
			serviceResponse:         servicesResponseB,
			expectedServiceResponse: servicesResponseA,
			expectedStatus:          true,
		},
		{
			name:                    "service fetch fails",
			serviceResponse:         servicesResponseA,
			expectedServiceResponse: core.ServicesResponse{},
			fetchServiceErr:         errors.New("failed to fetch the services data"),
			expectedStatus:          false,
			shouldReturnError:       true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			servicesState := ServicesState{
				credentialsAPI: &coremock.CredentialsAPIMock{
					ServicesResponse: test.serviceResponse,
					ServicesErr:      test.fetchServiceErr,
				},
				services:        test.storedServiceData,
				servicesFetched: test.hasFetchedServiceData,
			}

			services, err := servicesState.fetchServices()

			if test.shouldReturnError {
				assert.Error(t, err, "Expected error when fetching services.")
			} else {
				assert.NoError(t, err, "Unexpected error when fetching services.")
			}
			assert.Equal(t, test.expectedServiceResponse, services, "Unexpected services returned when fetching services.")
			assert.Equal(t, test.expectedStatus, servicesState.servicesFetched,
				"servicesFetched should be set to true after fetching services for the first time.")
		})
	}
}

func TestNotifyUserServicesChanged(t *testing.T) {
	category.Set(t, category.Unit)

	servicesResponse := core.ServicesResponse{
		core.ServiceData{ID: 1, ExpiresAt: "2026-07-14"},
	}

	servicesState := ServicesState{
		credentialsAPI: &coremock.CredentialsAPIMock{
			ServicesResponse: servicesResponse,
		},
		servicesFetched: true,
	}

	err := servicesState.NotifyUserServicesChanged(struct{}{})
	assert.NoError(t, err, "Unexpected error returned by NotifyUserServicesChanged.")

	services, _ := servicesState.fetchServices()
	assert.Equal(t, servicesResponse, services, "Services should be updated after services change notification.")
}

func TestNotifyUserServicesChangedError(t *testing.T) {
	category.Set(t, category.Unit)

	servicesResponse := core.ServicesResponse{
		core.ServiceData{ID: 1, ExpiresAt: "2026-07-14"},
	}
	currentServiceData := core.ServicesResponse{
		core.ServiceData{ID: 2, ExpiresAt: "2026-10-27"},
	}

	apiMock := coremock.CredentialsAPIMock{
		ServicesResponse: servicesResponse,
		ServicesErr:      errors.New("failed to fetch services"),
	}
	servicesState := ServicesState{
		credentialsAPI:  &apiMock,
		servicesFetched: true,
		services:        currentServiceData,
	}

	err := servicesState.NotifyUserServicesChanged(struct{}{})
	assert.Error(t, err, "Expected error when updating services.")

	apiMock.ServicesErr = nil
	services, _ := servicesState.fetchServices()

	assert.Equal(t, currentServiceData, services,
		"Old services should be kept in case of services change notification failure.")
}
