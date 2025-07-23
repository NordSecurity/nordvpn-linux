package session

import (
	"errors"
	"time"
)

// GetToken
func GetToken(store SessionStore) (string, error) {
	switch s := store.(type) {
	case *AccessTokenSessionStore:
		return s.session.GetToken(), nil
	default:
		return "", errors.New("gettoken: incompatible session store")
	}
}

// GetRenewalToken
func GetRenewalToken(store SessionStore) (string, error) {
	switch s := store.(type) {
	case *AccessTokenSessionStore:
		return s.session.GetRenewalToken(), nil
	default:
		return "", errors.New("getrenewaltoken: incompatible session store")
	}
}

// GetOwnerID
func GetOwnerID(store SessionStore) (string, error) {
	switch s := store.(type) {
	case *TrustedPassSessionStore:
		return s.session.GetOwnerID(), nil
	default:
		return "", errors.New("getownerid: incompatible session store")
	}
}

// GetExpiry
func GetExpiry(store SessionStore) (time.Time, error) {
	switch s := store.(type) {
	case *AccessTokenSessionStore:
		return s.session.GetExpiry(), nil
	default:
		return time.Time{}, errors.New("getexpiry: incompatible session store")
	}
}
