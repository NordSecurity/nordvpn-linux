package core

import (
	"errors"
	"regexp"
)

var (
	ErrLoginTokenExpired = errors.New("login token expired")
	ErrLoginTokenRevoked = errors.New("login token revoked")
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
		return ErrLoginTokenRevoked
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
