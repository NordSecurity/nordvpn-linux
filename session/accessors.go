package session

import (
	"errors"
	"time"
)

// GetToken retrieves a token from a SessionTokenProvider
func GetToken(store interface{}) (string, error) {
	if provider, ok := store.(SessionTokenProvider); ok {
		return provider.GetToken(), nil
	}
	return "", errors.New("gettoken: store does not implement session token provider")
}

// GetRenewalToken retrieves a renewal token from a SessionRenewalTokenProvider
func GetRenewalToken(store interface{}) (string, error) {
	if provider, ok := store.(SessionRenewalTokenProvider); ok {
		return provider.GetRenewalToken(), nil
	}
	return "", errors.New("getrenewaltoken: store does not implement session renewal token provider")
}

// GetOwnerID retrieves an owner ID from a SessionOwnerProvider
func GetOwnerID(store interface{}) (string, error) {
	if provider, ok := store.(SessionOwnerProvider); ok {
		return provider.GetOwnerID(), nil
	}
	return "", errors.New("getownerid: store does not implement session owner provider")
}

// GetExpiry retrieves the expiry time from a SessionExpiryProvider
func GetExpiry(store interface{}) (time.Time, error) {
	if provider, ok := store.(SessionExpiryProvider); ok {
		return provider.GetExpiry(), nil
	}
	return time.Time{}, errors.New("getexpiry: store does not implement session expiry provider")
}
