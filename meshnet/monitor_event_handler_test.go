package meshnet

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"github.com/stretchr/testify/assert"
)

func TestEventHandler_OnProcessStarted(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name      string
		savedPID  PID
		finalPID  PID
		procEvent ProcEvent
		pc        processChecker
	}{
		{
			name:      "PID is set if process is fileshare",
			savedPID:  PID(0),
			procEvent: ProcEvent{1337},
			finalPID:  PID(1337),
			pc:        procCheckerStub{isFileshare: true},
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
				cm:             mock.NewMockConfigManager(),
				reg:            &mock.RegistryMock{},
			}
			assert.Equal(t, tt.savedPID, pac.filesharePID)

			pac.OnProcessStarted(tt.procEvent)

			assert.Equal(t, tt.finalPID, pac.filesharePID)
		})
	}
}

func TestEventHandler_OnProcessStopped(t *testing.T) {
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
				cm:             mock.NewMockConfigManager(),
				reg:            &mock.RegistryMock{},
			}
			assert.Equal(t, tt.savedPID, pac.filesharePID)

			pac.OnProcessStopped(tt.procEvent)

			assert.Equal(t, tt.finalPID, pac.filesharePID)
		})
	}
}

type procCheckerStub struct {
	isFileshare bool
}

func (pu procCheckerStub) isFileshareProcess(PID) bool {
	return pu.isFileshare
}
