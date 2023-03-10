// Package netstate is responsible for network state monitoring and network state based vpn configuration updates.
package netstate

// State defines the current network state of a system
type State uint8

const (
	// Unknown defines unknown system network state
	Unknown = iota
	// Down defines up system network state
	Down
	// Up defines up system network state
	Up
	// TODO: No net. Up but network is not reachable
)
