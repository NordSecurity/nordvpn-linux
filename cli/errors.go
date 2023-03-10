package cli

import "errors"

var (
	ErrInternetConnection = errors.New(CheckYourInternetConnMessage)
	ErrAccountExpired     = errors.New(ExpiredAccountMessage)
	ErrUpdateAvailable    = errors.New(UpdateAvailableMessage)
)
