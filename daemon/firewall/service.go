package firewall

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
