package fileshare

import (
	"net/netip"

	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"
)

// NoopFileshare is a noop implementation of fileshare. It is used when libdrop
// is not available and should be used only for development purposes.
type NoopFileshare struct{}

// Enable is a stub
func (NoopFileshare) Enable(netip.Addr) error { return nil }

// Disable is a stub
func (NoopFileshare) Disable() error { return nil }

// Send is a stub
func (NoopFileshare) Send(netip.Addr, []string) (string, error) { return "", nil }

// Accept is a stub
func (NoopFileshare) Accept(string, string, string) error { return nil }

// Cancel is a stub
func (NoopFileshare) Cancel(string) error { return nil }

// CancelFile is a stub
func (NoopFileshare) CancelFile(transferID string, fileID string) error { return nil }

// NoopStorage is a noop implementation of fileshare storage. It is used when no persistence is desired.
type NoopStorage struct{}

// Load is a stub
func (NoopStorage) Load() (map[string]*pb.Transfer, error) {
	return map[string]*pb.Transfer{}, nil
}

// Save is a stub
func (NoopStorage) Save(map[string]*pb.Transfer) error { return nil }
