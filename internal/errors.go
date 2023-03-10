package internal

import (
	"errors"
)

var (
	ErrDaemonConnectionRefused = errors.New(DaemonConnRefusedErrorMessage)
	ErrSocketAccessDenied      = errors.New("Whoops! Permission denied accessing " + DaemonSocket)
	ErrSocketNotFound          = errors.New("Whoops! " + DaemonSocket + " not found")
	ErrUnhandled               = errors.New(UnhandledMessage)
	ErrGateway                 = errors.New("can't find gateway")
	ErrStdin                   = errors.New("Stdin: missing argument")
	ErrServerIsUnavailable     = errors.New(ServerUnavailableErrorMessage)
	ErrTagDoesNotExist         = errors.New(TagNonexistentErrorMessage)
	ErrGroupDoesNotExist       = errors.New(GroupNonexistentErrorMessage)
	ErrDoubleGroup             = errors.New(DoubleGroupErrorMessage)
	// ErrAlreadyLoggedIn is returned on repeated logins
	ErrAlreadyLoggedIn = errors.New("you are already logged in")
	// ErrNotLoggedIn is returned when the caller is expected to be logged in
	// but is not
	ErrNotLoggedIn = errors.New("you are not logged in")
)
