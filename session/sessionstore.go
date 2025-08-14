package session

// RenewalOption configures renewal behavior
type RenewalOption func(*renewalOptions)

// renewalOptions holds configuration for renewal operations
type renewalOptions struct {
	skipErrorHandlers bool
	forceRenewal      bool
}

// SilentRenewal returns a RenewalOption that performs renewal without triggering side effects
// such as error callbacks, notifications, or other external actions
func SilentRenewal() RenewalOption {
	return func(o *renewalOptions) {
		o.skipErrorHandlers = true
	}
}

// ForceRenewal returns a RenewalOption that forces token renewal even if the current token is valid.
func ForceRenewal() RenewalOption {
	return func(o *renewalOptions) {
		o.forceRenewal = true
	}
}

// SessionStore is an interface for managing session data.
type SessionStore interface {
	// Renew extends the lifetime of the current session.
	// This is typically used to keep a session active beyond its default expiration.
	// Options can be provided to customize the renewal behavior.
	// Returns an error if the session cannot be renewed.
	Renew(opts ...RenewalOption) error

	// HandleError processes errors that occur during session operations.
	// It returns nil if the error was not handled, or the error itself if it was.
	HandleError(reason error) error
}

// TokenSessionStore extends SessionStore with token access capabilities
type TokenSessionStore interface {
	SessionStore
	GetToken() string
}
