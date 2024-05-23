package cli

import "errors"

var (
	ErrInternetConnection = errors.New(CheckYourInternetConnMessage)
	ErrNoDedicatedIP      = errors.New(NoDedicatedIPMessage)
	ErrUpdateAvailable    = errors.New(UpdateAvailableMessage)
)
