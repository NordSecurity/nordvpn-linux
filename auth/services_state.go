package auth

import (
	"fmt"
	"time"

	"github.com/NordSecurity/nordvpn-linux/core"
	devicekey "github.com/NordSecurity/nordvpn-linux/device_key"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/internal/caching"
	"github.com/NordSecurity/nordvpn-linux/log"
)

// ServicesState is responsible for fetching and storing user's available services.
type ServicesState struct {
	cache                     *caching.Cache[core.ServicesResponse]
	dedicatedServerKeyManager devicekey.DedicatedServersKeyManager
	expChecker                core.ExpirationChecker
}

func NewServicesState(credentialsAPI core.CredentialsAPI,
	dedicatdServerKeyManager devicekey.DedicatedServersKeyManager) *ServicesState {
	cache := caching.NewCacheWithTTL(time.Hour*5, credentialsAPI.Services)
	return &ServicesState{
		cache:                     cache,
		expChecker:                systemTimeExpirationChecker{},
		dedicatedServerKeyManager: dedicatdServerKeyManager,
	}
}

// fetchServices fetches user's services from the API and saves them so that they will be returned on subsequent calls.
func (s *ServicesState) fetchServices() (core.ServicesResponse, error) {
	return s.cache.Get()
}

func (s *ServicesState) NotifyUserServicesChanged(any) error {
	log.Println(internal.DebugPrefix, "received user service change notification, invalidating service cache")

	newServices, err := s.cache.Fetch()
	if err != nil {
		return fmt.Errorf("fetching new services: %w", err)
	}

	if hasDedicatedServerService(newServices, s.expChecker) {
		s.dedicatedServerKeyManager.CheckAndRegisterDedicatedServers()
	}

	return nil
}

func (s *ServicesState) NotifyLogout(events.DataAuthorization) error {
	s.cache.Invalidate()
	return nil
}
