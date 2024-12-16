package meshnet

import (
	"sync"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestOnProcessStarted_ManagesPID(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name      string
		savedPID  PID
		finalPID  PID
		procEvent ProcEvent
		pc        ProcessChecker
	}{
		{
			name:      "PID is set if process is fileshare",
			savedPID:  PID(0),
			procEvent: ProcEvent{1337},
			finalPID:  PID(1337),
			pc:        procCheckerStub{isFileshare: true, daemonPID: 1336}, // currentPID lower than the PID from event
		},
		{
			name:      "PID is not updated if the event's PID is older than current process PID",
			savedPID:  PID(0),
			procEvent: ProcEvent{1337},
			finalPID:  PID(0),
			pc:        procCheckerStub{isFileshare: true, daemonPID: 1338}, // currentPID higher than the PID from event
		},
		{
			name:      "PID is not updated if process is NOT fileshare",
			savedPID:  PID(0),
			procEvent: ProcEvent{1337},
			finalPID:  PID(0),
			pc:        procCheckerStub{isFileshare: false},
		},
		{
			name:     "no processing when PID is already set",
			savedPID: PID(1337),
			finalPID: PID(1337),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pac := FilesharePortAccessController{
				filesharePID:   tt.savedPID,
				processChecker: tt.pc,
				netw:           fileshareNetworkerDummy{},
			}
			assert.Equal(t, tt.savedPID, pac.filesharePID)

			pac.OnProcessStarted(tt.procEvent)

			assert.Equal(t, tt.finalPID, pac.filesharePID)
		})
	}
}

func TestOnProcessStarted_PermitsFileshareWhenProcessStarted(t *testing.T) {
	category.Set(t, category.Unit)
	fileshareNetworker := newNetworkerSpy()
	pac := FilesharePortAccessController{
		filesharePID: PID(0),
		processChecker: procCheckerStub{
			isFileshare: true, // detects every new process as fileshare process
			daemonPID:   1336,
		},
		netw: fileshareNetworker,
	}
	// new fileshare process event appeared and it's younger than daemonPID
	newEvent := ProcEvent{1337}
	assert.Equal(t, PID(0), pac.filesharePID)
	assert.False(t, fileshareNetworker.permitCalled)

	pac.OnProcessStarted(newEvent)
	fileshareNetworker.waitForPermitCall(t)

	assert.True(t, fileshareNetworker.permitCalled)
}

func TestOnProcessStarted_DoesNotPermitFileshare(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name      string
		savedPID  PID
		procEvent ProcEvent
		pc        ProcessChecker
	}{
		{
			name:      "when the event's PID is older than current process PID",
			savedPID:  PID(0),
			procEvent: ProcEvent{1337},
			pc:        procCheckerStub{isFileshare: true, daemonPID: 1338}, // daemon PID higher than the PID from event
		},
		{
			name:     "when process is NOT fileshare",
			savedPID: PID(0),
			pc:       procCheckerStub{isFileshare: false},
		},
		{
			name:     "when fileshare was already permitted",
			savedPID: PID(1337),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileshareNetworker := newNetworkerSpy()
			pac := FilesharePortAccessController{
				filesharePID:   tt.savedPID,
				processChecker: tt.pc,
				netw:           fileshareNetworker,
			}
			assert.Equal(t, tt.savedPID, pac.filesharePID)
			assert.False(t, fileshareNetworker.permitCalled)

			pac.OnProcessStarted(tt.procEvent)
			fileshareNetworker.ensurePermitNotCalled(t, 100*time.Microsecond)

			assert.False(t, fileshareNetworker.permitCalled)
		})
	}
}

func TestOnProcessStopped_ManagesPID(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name      string
		savedPID  PID
		finalPID  PID
		procEvent ProcEvent
	}{
		{
			name:      "PID is NOT zeroed when event's PID does not match with the saved one",
			savedPID:  PID(1337),
			procEvent: ProcEvent{666},
			finalPID:  PID(1337),
		},
		{
			name:      "PID is zeroed when event PID does match",
			savedPID:  PID(1337),
			procEvent: ProcEvent{1337},
			finalPID:  PID(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notImportant := false
			pac := FilesharePortAccessController{
				filesharePID:   tt.savedPID,
				processChecker: procCheckerStub{isFileshare: notImportant},
				netw:           fileshareNetworkerDummy{},
			}
			assert.Equal(t, tt.savedPID, pac.filesharePID)

			pac.OnProcessStopped(tt.procEvent)

			assert.Equal(t, tt.finalPID, pac.filesharePID)
		})
	}
}

func TestOnProcessStopped_ForbidsFileshareWhenProcessStopped(t *testing.T) {
	category.Set(t, category.Unit)
	fileshareNetworker := newNetworkerSpy()
	pac := FilesharePortAccessController{
		filesharePID: PID(1337),
		// detects every new process as fileshare process
		processChecker: procCheckerStub{isFileshare: true},
		netw:           fileshareNetworker,
	}
	// new fileshare process event appeared with PID the same as recorded fileshare PID
	newEvent := ProcEvent{1337}
	assert.Equal(t, PID(1337), pac.filesharePID)
	assert.False(t, fileshareNetworker.forbidCalled)

	pac.OnProcessStopped(newEvent)
	fileshareNetworker.waitForForbidCall(t)

	assert.True(t, fileshareNetworker.forbidCalled)
}

func TestOnProcessStarted_DoesNotForbidFileshare(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name      string
		savedPID  PID
		procEvent ProcEvent
	}{
		{
			name:      "when the event's PID does not match saved fileshare PID",
			savedPID:  PID(1337),
			procEvent: ProcEvent{666},
		},
		{
			name:     "when fileshare was already forbidden",
			savedPID: PID(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileshareNetworker := newNetworkerSpy()
			pac := FilesharePortAccessController{
				filesharePID:   tt.savedPID,
				processChecker: procCheckerStub{isFileshare: true},
				netw:           fileshareNetworker,
			}
			assert.Equal(t, tt.savedPID, pac.filesharePID)
			assert.False(t, fileshareNetworker.forbidCalled)

			pac.OnProcessStopped(tt.procEvent)
			fileshareNetworker.ensureForbidNotCalled(t, 100*time.Microsecond)

			assert.False(t, fileshareNetworker.permitCalled)
		})
	}
}

// procChecker
type procCheckerStub struct {
	isFileshare bool
	daemonPID   PID
}

func (pc procCheckerStub) IsFileshareProcess(PID) bool {
	return pc.isFileshare
}

func (pu procCheckerStub) GiveProcessPID(string) *PID {
	return nil
}

func (pc procCheckerStub) CurrentPID() PID {
	return pc.daemonPID
}

// fileshareNetworker spy
type fileshareNetworkerSpy struct {
	permitCh     chan struct{}
	permitCalled bool
	forbidCh     chan struct{}
	forbidCalled bool
}

func newNetworkerSpy() *fileshareNetworkerSpy {
	var wg sync.WaitGroup
	wg.Add(1)
	return &fileshareNetworkerSpy{
		permitCh:     make(chan struct{}),
		permitCalled: false,
		forbidCh:     make(chan struct{}),
		forbidCalled: false,
	}
}

func (fn *fileshareNetworkerSpy) PermitFileshare() error {
	fn.permitCh <- struct{}{}
	return nil
}

func (fn *fileshareNetworkerSpy) waitForPermitCall(t *testing.T) {
	t.Helper()
	select {
	case <-fn.permitCh:
		fn.permitCalled = true
		return
	case <-time.After(time.Second):
		t.Fatal("fileshare should be permitted but was not")
	}
}

func (fn *fileshareNetworkerSpy) ensurePermitNotCalled(t *testing.T, d time.Duration) {
	t.Helper()
	select {
	case <-fn.permitCh:
		t.Fatal("fileshare should NOT be permitted but was")
		return
	case <-time.After(d):
		// OK
	}
}

func (fn *fileshareNetworkerSpy) ForbidFileshare() error {
	fn.forbidCh <- struct{}{}
	return nil
}

func (fn *fileshareNetworkerSpy) waitForForbidCall(t *testing.T) {
	t.Helper()
	select {
	case <-fn.forbidCh:
		fn.forbidCalled = true
		return
	case <-time.After(time.Second):
		t.Fatal("fileshare should be forbidden but was not")
	}
}

func (fn *fileshareNetworkerSpy) ensureForbidNotCalled(t *testing.T, d time.Duration) {
	t.Helper()
	select {
	case <-fn.forbidCh:
		t.Fatal("fileshare should NOT be forbidden but was")
		return
	case <-time.After(d):
		// OK
	}
}

// fileshareNetworker dummy
type fileshareNetworkerDummy struct{}

func (fn fileshareNetworkerDummy) PermitFileshare() error {
	return nil
}

func (fn fileshareNetworkerDummy) ForbidFileshare() error {
	return nil
}
