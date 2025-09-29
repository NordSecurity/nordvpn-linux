package tray

// CheckableMenuItem defines the interface for a menu item that can be checked and unchecked.
// This allows for mocking the systray.MenuItem in tests.
type CheckableMenuItem interface {
	// ClickedCh returns the channel that receives click events.
	ClickedCh() <-chan struct{}
	// Checked returns whether the item is checked.
	Checked() bool
	// Check marks the item as checked.
	Check()
	// Uncheck marks the item as unchecked.
	Uncheck()
}
