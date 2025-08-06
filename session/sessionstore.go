package session

// SessionStore is an interface for managing session data.
type SessionStore interface {
	// Renew extends the lifetime of the current session.
	// This is typically used to keep a session active beyond its default expiration.
	// Returns an error if the session cannot be renewed.
	Renew() error

	// HandleError processes errors that occur during session operations.
	// It returns nil if the error was not handled, or the error itself if it was.
	HandleError(reason error) error
}

// TokenSessionStore extends SessionStore with token access capabilities
type TokenSessionStore interface {
	SessionStore
	GetToken() string
}
