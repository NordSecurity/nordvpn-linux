package service

// Service defines the operations for managing a service lifecycle
type Service interface {
	// Enable sets up the service for a specific user
	// uid: User ID
	// gid: Group ID
	// home: Home directory path
	// Returns an error if the setup fails
	Enable(uid uint32, gid uint32, home string) error

	// Stop terminates the service for the given user
	// uid: User ID
	// wait: Whether to wait for the service to stop before returning
	// Returns an error if the stop operation fails
	Stop(uid uint32, wait bool) error

	// StopAll stops all instances of the service across users
	StopAll()

	// Restart reinitializes the service for a given user
	// uid: User ID
	// Returns an error if the restart operation fails
	Restart(uid uint32) error
}
