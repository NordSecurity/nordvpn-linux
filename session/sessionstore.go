package session

// SessionStore is an interface for managing session data.
type SessionStore interface {
	// Renew extends the lifetime of the current session.
	// This is typically used to keep a session active beyond its default expiration.
	// Returns an error if the session cannot be renewed.
	Renew() error

	// Invalidate terminates the current session with the specified reason.
	// The reason is provided as an error that explains why the session was invalidated.
	// Returns an error if the session cannot be invalidated.
	Invalidate(reason error) error
}

// SessionStoreValidator is an interface for validating session data.
type SessionStoreValidator interface {
	// Validate checks whether the provided session object is valid.
	// Returns an error if the session is invalid or fails validation checks.
	Validate(session any) error
}
