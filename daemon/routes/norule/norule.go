// Package norule implements noop policy router.
package norule

type Facade struct{}

func (*Facade) SetupRoutingRules(bool, bool, bool) error { return nil }
func (*Facade) CleanupRouting() error                    { return nil }
func (*Facade) TableID() uint                            { return 0 }
