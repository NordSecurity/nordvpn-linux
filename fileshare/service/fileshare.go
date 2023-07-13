// Package service provides structures to start fileshare service (nordfileshared)
package service

// Fileshare service management
type Fileshare interface {
	// uid and gid of the user which is making the call
	Enable(uid, gid uint32) error
	// uid and gid of the user which made the last successful Enable call
	Disable(uid, gid uint32) error
	// Stop is used when filesharing needs to be turned off but meshnet was not disabled by
	// the user, so on app shutdown for example
	Stop(uid, gid uint32) error
}
