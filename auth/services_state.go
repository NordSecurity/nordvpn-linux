package auth

import (
	"time"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal/caching"
)

// ServicesState is responsible for fetching and storing user's available services.
type ServicesState struct {
	cache *caching.Cache[core.ServicesResponse]
}

func NewServicesState(credentialsAPI core.CredentialsAPI) *ServicesState {
	cache := caching.NewCacheWithTTL(time.Hour*5, credentialsAPI.Services)
	return &ServicesState{
		cache: cache,
	}
}

// fetchServices fetches user's services from the API and saves them so that they will be returned on subsequent calls.
func (s *ServicesState) fetchServices() (core.ServicesResponse, error) {
	return s.cache.Get()
}

func (s *ServicesState) NotifyUserServicesChanged(any) error {
	s.cache.Invalidate()
	return nil
}

func (s *ServicesState) NotifyLogout(events.DataAuthorization) error {
	s.cache.Invalidate()
	return nil
}
