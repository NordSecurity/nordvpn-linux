// Package norule implements noop policy router.
package norule

import "net"

type Facade struct{}

func (*Facade) SetupRoutingRules(net.Interface, bool, bool) error { return nil }
func (*Facade) CleanupRouting() error                             { return nil }
func (*Facade) TableID() uint                                     { return 0 }
