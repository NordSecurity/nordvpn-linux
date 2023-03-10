package firewall

const (
	// Established means that packet is associated with a connection
	Established ConnectionState = iota
	// Related means that packet creates a new connection, but it is related with the existing one
	Related
	// New means that packet creates a new connection
	New
)

const (
	// Inbound defines that rule is applicable for incoming packets
	Inbound Direction = iota
	// Outbound defines that rule is applicable for outgoing packets
	Outbound
	// TwoWay defines that rule is applicable for both incoming and outgoing packets
	TwoWay
)

// ConnectionState defines a state of a connection
type ConnectionState int

// Direction defines a direction of packages to which rule is applicable
type Direction int

// Service adapts system firewall configuration to firewall rules
//
// Used by callers.
type Service interface {
	// Add and apply firewall rules
	Add([]Rule) error
	// Delete a list of firewall rules by defined names
	Delete(names []string) error
	// Enable firewall
	Enable() error
	// Disable firewall
	Disable() error
	// IsEnabled reports whether firewall is enabled or not
	IsEnabled() bool
}

// Agent carries out required firewall changes.
//
// Used by implementers.
type Agent interface {
	// Add a firewall rule
	Add(Rule) error
	// Delete a firewall rule
	Delete(Rule) error
}
