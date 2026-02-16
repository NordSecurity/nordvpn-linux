package internal

import (
	"errors"
	"fmt"
)

// CodedError wraps internal code value for the error
type CodedError struct {
	Code    int
	Message string
	Err     error
}

// Error formats human readable representation of an underlying coded error
func (e CodedError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[Code %d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[Code %d] %s", e.Code, e.Message)
}

// Unwrap underlying error is one exists
func (e *CodedError) Unwrap() error {
	return e.Err
}

// NewCodedError constructs a new coded error
func NewCodedError(code int, message string, err error) error {
	return &CodedError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

var (
	ErrDaemonConnectionRefused = errors.New(DaemonConnRefusedErrorMessage)
	ErrSocketAccessDenied      = errors.New("Couldn't access " + DaemonSocket + ". To fix the issue, you may need to reinstall the app.")
	ErrSocketNotFound          = errors.New("Couldn't find " + DaemonSocket + ". To fix the issue, you may need to reinstall the app.")
	ErrUnhandled               = errors.New(UnhandledMessage)
	ErrGateway                 = errors.New("can't find gateway")
	ErrStdin                   = errors.New("Stdin: missing argument")
	ErrServerIsUnavailable     = errors.New(ServerUnavailableErrorMessage)
	ErrServerDataIsNotReady    = errors.New("Server data is not ready")
	ErrTagDoesNotExist         = errors.New(TagNonexistentErrorMessage)
	ErrGroupDoesNotExist       = errors.New(GroupNonexistentErrorMessage)
	ErrDoubleGroup             = errors.New(DoubleGroupErrorMessage)
	// ErrAnalyticsConsentMissing is returned when user tries to login via tray
	// but settings analytics consent failed for some reason. This should not happen.
	ErrAnalyticsConsentMissing = errors.New("analytics consent is required before continuing")
	ErrVirtualServerSelected   = errors.New(SpecifiedServerIsVirtualLocation)
	ErrNoNetWhenLoggingIn      = errors.New("You’re offline.\nWe can’t run this action without an internet connection. Please check it and try again.")

	// WARN: The error messages below are also used to detect error states in GUI.
	// When updating, make sure to double check the GUI errors here:
	// https://github.com/NordSecurity/nordvpn-linux/blob/main/gui/lib/data/repository/daemon_status_codes.dart#L49-L49

	// ErrAlreadyLoggedIn is returned on repeated logins
	ErrAlreadyLoggedIn = errors.New("You're already logged in")
	// ErrNotLoggedIn is returned when the caller is expected to be logged in
	// but is not
	ErrNotLoggedIn = errors.New("You're not logged in")
	// ErrMissingExchangeToken is returned when login was successful but
	// there is not enough data to request the token
	ErrMissingExchangeToken = errors.New("The exchange token is missing. Please try logging in again. If the issue persists, contact our customer support.")
)
