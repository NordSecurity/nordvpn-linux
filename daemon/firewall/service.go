package firewall

import (
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
)

// Service adapts system firewall configuration to firewall rules
//
// Used by callers.
type Service interface {
	Configure(vpnInfo *VpnInfo, meshMap *mesh.MachineMap) error
	Remove() error
	Flush() error
	Disable() error
	Enable() error
}

// Agent carries out required firewall changes.
//
// Used by implementers.
// type Agent interface {
// 	// Add a firewall rule
// 	Add(Rule) error
// 	// Delete a firewall rule
// 	Delete(Rule) error
// 	// Flush removes all nordvpn rules
// 	Flush() error
// 	// GetActiveRules gets currently active rules by name
// 	GetActiveRules() ([]string, error)
// }

type FwImpl interface {
	Configure(vpnInfo *VpnInfo, meshMap *mesh.MachineMap) error
	Flush() error
}
