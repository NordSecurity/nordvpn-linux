package tray

import (
	"sync"
	"time"

	"github.com/NordSecurity/systray"
)

// CheckboxSynchronizer manages synchronization between checkbox operations and menu rebuilding
// to prevent race conditions that cause menu item duplication in the systray library.
type CheckboxSynchronizer struct {
	operationInProgress bool
	mu                  sync.Mutex
}

// NewCheckboxSynchronizer creates a new CheckboxSynchronizer instance
func NewCheckboxSynchronizer() *CheckboxSynchronizer {
	return &CheckboxSynchronizer{}
}

// WaitForOperations blocks until all active checkbox operations complete.
// This should be called before menu rebuilding operations like systray.ResetMenu().
func (cs *CheckboxSynchronizer) WaitForOperations() {
	for {
		cs.mu.Lock()
		inProgress := cs.operationInProgress
		cs.mu.Unlock()

		if !inProgress {
			break
		}

		// Small delay to avoid busy waiting
		time.Sleep(10 * time.Millisecond)
	}
}

// SetOperationInProgress marks a checkbox operation as active or complete.
// This should be called at the start (true) and end (false) of checkbox operations.
func (cs *CheckboxSynchronizer) SetOperationInProgress(inProgress bool) {
	cs.mu.Lock()
	cs.operationInProgress = inProgress
	cs.mu.Unlock()
}

// HandleCheckboxOption provides a standardized way to handle checkbox menu items
// with proper synchronization to prevent menu duplication issues.
func (cs *CheckboxSynchronizer) HandleCheckboxOption(item *systray.MenuItem, setter func(bool) bool) {
	if item == nil || setter == nil {
		return
	}
	go func() {
		success := false
		for !success {
			_, open := <-item.ClickedCh
			if !open {
				return
			}

			// Mark checkbox operation as in progress
			cs.SetOperationInProgress(true)

			unchecked := !item.Checked()
			success = setter(unchecked)
			if success {
				if unchecked {
					item.Check()
				} else {
					item.Uncheck()
				}
			}

			// Mark checkbox operation as complete
			cs.SetOperationInProgress(false)
		}
	}()
}
