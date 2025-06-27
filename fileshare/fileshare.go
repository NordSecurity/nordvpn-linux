// Package fileshare provides gRPC interface for the fileshare functionality.
package fileshare

import (
	"errors"
	"net/netip"
	"time"

	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"
)

const (
	DirDepthLimit     = 5
	TransferFileLimit = 1000
)

var ErrAddressAlreadyInUse = errors.New("address already in use")

// Fileshare defines a set of operations that any type that wants to act as a fileshare service
// must implement.
type Fileshare interface {
	// Enable starts service listening at provided address
	Enable(listenAddress netip.Addr) error
	// Disable tears down fileshare service
	Disable() error
	// Send sends the provided file or dir to provided peer and returns transfer ID
	Send(peer netip.Addr, paths []string) (string, error)
	// Accept accepts provided files from provided request and starts download process
	Accept(transferID, dstPath string, fileID string) error
	// Finalize file transfer by ID.
	Finalize(transferID string) error
	// CancelFile id in a transfer
	CancelFile(transferID string, fileID string) error
}

// Storage is used for filesharing history persistence
type Storage interface {
	Load() (map[string]*pb.Transfer, error)
	PurgeTransfersUntil(until time.Time) error
}
