package tray

import (
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/systray"
	"github.com/stretchr/testify/assert"
)

func TestNewCheckboxSynchronizer(t *testing.T) {
	category.Set(t, category.Unit)
	cs := NewCheckboxSynchronizer()
	assert.NotNil(t, cs, "NewCheckboxSynchronizer should not return nil")
	assert.False(t, cs.operationInProgress, "operationInProgress should be false initially")
}

func TestSetOperationInProgress(t *testing.T) {
	category.Set(t, category.Unit)
	cs := NewCheckboxSynchronizer()

	cs.SetOperationInProgress(true)
	assert.True(t, cs.operationInProgress, "operationInProgress should be true after setting to true")

	cs.SetOperationInProgress(false)
	assert.False(t, cs.operationInProgress, "operationInProgress should be false after setting to false")
}

func TestWaitForOperations(t *testing.T) {
	category.Set(t, category.Unit)
	cs := NewCheckboxSynchronizer()

	// Test that it returns immediately if no operation is in progress
	start := time.Now()
	cs.WaitForOperations()
	duration := time.Since(start)
	assert.Less(t, duration, 10*time.Millisecond, "WaitForOperations should return immediately when no operation is in progress")

	// Test that it waits for an operation to complete
	cs.SetOperationInProgress(true)
	go func() {
		time.Sleep(20 * time.Millisecond)
		cs.SetOperationInProgress(false)
	}()

	start = time.Now()
	cs.WaitForOperations()
	duration = time.Since(start)
	assert.GreaterOrEqual(t, duration, 20*time.Millisecond, "WaitForOperations should wait for the operation to complete")
}

// Note: The success paths for HandleCheckboxOption (where the setter function
// returns true) are not tested here. This is because they call item.Check()
// or item.Uncheck() on a systray.MenuItem. These methods attempt to update
// the UI and require the systray event loop to be running. In a unit test
// environment, this loop is not active, and the underlying data structures
// in the systray package are not initialized, leading to a panic.
// Testing this behavior would require a significant refactoring or a dedicated
// test harness that can mock the systray environment.

func TestHandleCheckboxOptionNilItem(t *testing.T) {
	category.Set(t, category.Unit)
	cs := NewCheckboxSynchronizer()
	assert.NotPanics(t, func() {
		cs.HandleCheckboxOption(nil, func(b bool) bool { return true })
	})
}

func TestHandleCheckboxOptionNilSetter(t *testing.T) {
	category.Set(t, category.Unit)
	cs := NewCheckboxSynchronizer()
	item := &systray.MenuItem{ClickedCh: make(chan struct{})}
	assert.NotPanics(t, func() {
		cs.HandleCheckboxOption(item, nil)
	})
}

func TestHandleCheckboxOptionSetterFails(t *testing.T) {
	category.Set(t, category.Unit)
	cs := NewCheckboxSynchronizer()
	item := &systray.MenuItem{ClickedCh: make(chan struct{}, 1)}
	setterCalledCh := make(chan bool, 1)

	setter := func(checked bool) bool {
		setterCalledCh <- true
		return false // Simulate failure
	}

	cs.HandleCheckboxOption(item, setter)

	item.ClickedCh <- struct{}{}

	select {
	case <-setterCalledCh:
		// Test passed
	case <-time.After(50 * time.Millisecond):
		t.Fatal("setter was not called")
	}

	assert.False(t, item.Checked(), "item should not be checked if setter fails")

	// Wait for the operation to be marked as complete
	cs.WaitForOperations()

	assert.False(t, cs.operationInProgress, "operation should be marked as complete even if setter fails")
}

func TestHandleCheckboxOptionChannelClosed(t *testing.T) {
	category.Set(t, category.Unit)
	cs := NewCheckboxSynchronizer()
	item := &systray.MenuItem{ClickedCh: make(chan struct{})}
	setterCalledCh := make(chan bool, 1)

	setter := func(checked bool) bool {
		setterCalledCh <- true
		return true
	}

	cs.HandleCheckboxOption(item, setter)

	close(item.ClickedCh)

	select {
	case <-setterCalledCh:
		t.Fatal("setter was called unexpectedly")
	case <-time.After(50 * time.Millisecond):
		// Test passed
	}
}
