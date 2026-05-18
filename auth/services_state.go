package auth

import (
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
)

// ServicesState is responsible for fetching and storing user's available services.
type ServicesState struct {
	credentialsAPI core.CredentialsAPI

	servicesFetched bool
	services        core.ServicesResponse
}

func NewServicesState(credentialsAPI core.CredentialsAPI) ServicesState {
	return ServicesState{
		credentialsAPI: credentialsAPI,
	}
}

// fetchServices fetches user's services from the API and saves them so that they will be returned on subsequent calls.
func (s *ServicesState) fetchServices() (core.ServicesResponse, error) {
	if s.servicesFetched {
		return s.services, nil
	}

	services, err := s.credentialsAPI.Services()
	if err != nil {
		return core.ServicesResponse{}, fmt.Errorf("fetching services: %w", err)
	}

	s.services = services
	s.servicesFetched = true

	return services, nil
}

func (s *ServicesState) NotifyUserServicesChanged(any) error {
	log.Println(internal.InfoPrefix, "received user services update, fetching new services")
	services, err := s.credentialsAPI.Services()
	if err != nil {
		return fmt.Errorf("fetching services: %w", err)
	}

	s.services = services
	s.servicesFetched = true
	return nil
}
