package tray

import (
	"sync"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

// mockCheckableMenuItem is a mock implementation of the CheckableMenuItem interface for testing.
type mockCheckableMenuItem struct {
	mu        sync.Mutex
	clickedCh chan struct{}
	checked   bool
}

func newMockCheckableMenuItem() *mockCheckableMenuItem {
	return &mockCheckableMenuItem{
		clickedCh: make(chan struct{}, 1),
	}
}

func (m *mockCheckableMenuItem) ClickedCh() <-chan struct{} {
	return m.clickedCh
}

func (m *mockCheckableMenuItem) Checked() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.checked
}

func (m *mockCheckableMenuItem) Check() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.checked = true
}

func (m *mockCheckableMenuItem) Uncheck() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.checked = false
}

func (m *mockCheckableMenuItem) Click() {
	m.clickedCh <- struct{}{}
}

func (m *mockCheckableMenuItem) Close() {
	close(m.clickedCh)
}

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

func TestHandleCheckboxOptionSuccessfulCheck(t *testing.T) {
	category.Set(t, category.Unit)
	cs := NewCheckboxSynchronizer()
	item := newMockCheckableMenuItem()
	setterCalled := make(chan bool, 1)
	setter := func(checked bool) bool {
		assert.True(t, checked)
		setterCalled <- true
		return true
	}

	cs.HandleCheckboxOption(item, setter)
	item.Click()

	select {
	case <-setterCalled:
	// continue
	case <-time.After(50 * time.Millisecond):
		t.Fatal("setter was not called")
	}

	assert.True(t, item.Checked(), "item should be checked")
	cs.WaitForOperations()
	assert.False(t, cs.operationInProgress, "operation should be marked as complete")
}

func TestHandleCheckboxOptionSuccessfulUncheck(t *testing.T) {
	category.Set(t, category.Unit)
	cs := NewCheckboxSynchronizer()
	item := newMockCheckableMenuItem()
	item.Check()
	setterCalled := make(chan bool, 1)
	setter := func(checked bool) bool {
		assert.False(t, checked)
		setterCalled <- true
		return true
	}

	cs.HandleCheckboxOption(item, setter)
	item.Click()

	select {
	case <-setterCalled:
	// continue
	case <-time.After(50 * time.Millisecond):
		t.Fatal("setter was not called")
	}

	assert.False(t, item.Checked(), "item should be unchecked")
	cs.WaitForOperations()
	assert.False(t, cs.operationInProgress, "operation should be marked as complete")
}

func TestHandleCheckboxOptionSetterFails(t *testing.T) {
	category.Set(t, category.Unit)
	cs := NewCheckboxSynchronizer()
	item := newMockCheckableMenuItem()
	setterCalled := make(chan bool, 1)
	setter := func(checked bool) bool {
		setterCalled <- true
		return false // Simulate failure
	}

	cs.HandleCheckboxOption(item, setter)
	item.Click()

	select {
	case <-setterCalled:
	// continue
	case <-time.After(50 * time.Millisecond):
		t.Fatal("setter was not called")
	}

	assert.False(t, item.Checked(), "item should not be checked if setter fails")
	cs.WaitForOperations()
	assert.False(t, cs.operationInProgress, "operation should be marked as complete even if setter fails")
}

func TestHandleCheckboxOptionChannelClosed(t *testing.T) {
	category.Set(t, category.Unit)
	cs := NewCheckboxSynchronizer()
	item := newMockCheckableMenuItem()
	setterCalled := make(chan bool, 1)
	setter := func(checked bool) bool {
		setterCalled <- true
		return true
	}

	cs.HandleCheckboxOption(item, setter)
	item.Close()

	select {
	case <-setterCalled:
		t.Fatal("setter was called unexpectedly")
	case <-time.After(50 * time.Millisecond):
	// Test passed
	}
}

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
	item := newMockCheckableMenuItem()
	assert.NotPanics(t, func() {
		cs.HandleCheckboxOption(item, nil)
	})
}
