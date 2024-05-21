package cli

import "errors"

var (
	ErrInternetConnection = errors.New(CheckYourInternetConnMessage)
	ErrAccountExpired     = errors.New(ExpiredAccountMessage)
	ErrNoDedicatedIP      = errors.New(NoDedicatedIPMessage)
	ErrUpdateAvailable    = errors.New(UpdateAvailableMessage)
)
