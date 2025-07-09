package core

import (
	"errors"
	"regexp"
	"time"
)

var (
	ErrLoginTokenExpired = errors.New("login token expired")
)

type LoginTokenValidator struct {
	expiryChecker ExpirationChecker
	credsAPI      RawCredentialsAPI
}

// Validate checks if a login token is still valid based on its expiry date.
// Returns nil if the token is valid, or ErrLoginTokenExpired if it has expired.
func (l *LoginTokenValidator) Validate(token string, expiryDate string) error {
	switch expiryDate {
	case "":
		return ErrLoginTokenExpired

	case ManualLoginTokenExpiryDate:
		return l.validateCredibility(token)

	default:
		return l.validateExpiryDate(expiryDate)
	}
}

func (l LoginTokenValidator) validateCredibility(token string) error {
	var isFormatValid = regexp.MustCompile(`^[a-f0-9]*$`).MatchString
	if !isFormatValid(token) {
		return ErrLoginTokenExpired
	}

	// this is the least expensive api call that needs authentication
	_, err := l.credsAPI.CurrentUser(token)
	// the only interest is in whether we receive "unauthorized access" error
	if errors.Is(err, ErrUnauthorized) {
		return ErrLoginTokenExpired
	}

	return nil
}

func (l LoginTokenValidator) validateExpiryDate(date string) error {
	if l.expiryChecker.IsExpired(date) {
		return ErrLoginTokenExpired
	}

	return nil
}

// NewLoginTokenValidator creates a new login-token validator
func NewLoginTokenValidator(api RawCredentialsAPI, expiryChecker ExpirationChecker) TokenValidator {
	return &LoginTokenValidator{credsAPI: api, expiryChecker: expiryChecker}
}

// remoteCheckDataValidityPeriod defines how long cache entries remain valid
const (
	remoteCheckDataValidityPeriod = time.Minute
)

// isCacheValid determines if a cached item is still within its validity period
func isCacheValid(addedAt time.Time) bool {
	// calculate when the cache entry expires
	expirationTime := addedAt.Add(remoteCheckDataValidityPeriod)
	return time.Now().Before(expirationTime)
}
