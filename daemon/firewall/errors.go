package firewall

import "fmt"

var (
	// ErrRuleNotFound defines that rule was not found in the firewall
	ErrRuleNotFound = fmt.Errorf("rule with specified name does not exist")
	// ErrRuleAlreadyExists defines that rule with specified name or parameters already exists
	ErrRuleAlreadyExists = fmt.Errorf("rule with specified name already exists")
	// ErrRuleWithoutName is returned when provided firewall rule does not have a name
	ErrRuleWithoutName = fmt.Errorf("rule must have a name")
	// ErrFirewallAlreadyEnabled defines that enable was called twice in a row
	ErrFirewallAlreadyEnabled = fmt.Errorf("firewall is already enabled")
	// ErrFirewallAlreadyDisabled defines that disable was called twice in a row
	ErrFirewallAlreadyDisabled = fmt.Errorf("firewall is already disabled")
)

// Error marks that it originated in firewall package
type Error struct {
	original error
}

func NewError(err error) error {
	return &Error{original: err}
}

func (e *Error) Error() string {
	return e.original.Error()
}

func (e *Error) Unwrap() error {
	return e.original
}
