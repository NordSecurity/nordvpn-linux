package meshnet

import (
	"errors"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestJobMonitorFileshare(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name                        string
		isFileshareInitiallyAllowed bool
		meshChecker                 meshChecker
		processChecker              processChecker
		wasForbidCalled             bool
		wasPermitCalled             bool
		shouldRulesChangeFail       bool
		isFileshareFinallyAllowed   bool
	}{
		{
			name:                        "nothing happens with disabled meshnet",
			meshChecker:                 meshCheckerStub{isMeshnetOn: false},
			wasForbidCalled:             false,
			wasPermitCalled:             false,
			isFileshareInitiallyAllowed: false,
			isFileshareFinallyAllowed:   false,
		},
		{
			name:                        "fileshare is allowed when meshnet is on and process is running",
			meshChecker:                 meshCheckerStub{isMeshnetOn: true},
			processChecker:              processCheckerStub{isFileshareUp: true},
			wasForbidCalled:             false,
			wasPermitCalled:             true,
			isFileshareInitiallyAllowed: false,
			isFileshareFinallyAllowed:   true,
		},
		{
			name:                        "fileshare is forbidden when meshnet is on process was stopped",
			meshChecker:                 meshCheckerStub{isMeshnetOn: true},
			processChecker:              processCheckerStub{isFileshareUp: false},
			wasForbidCalled:             true,
			wasPermitCalled:             false,
			isFileshareInitiallyAllowed: true,
			isFileshareFinallyAllowed:   false,
		},
		{
			name:                        "when mesh is off, but fileshare was allowed in previous run, now it gets blocked",
			meshChecker:                 meshCheckerStub{isMeshnetOn: false},
			wasForbidCalled:             true,
			wasPermitCalled:             false,
			isFileshareInitiallyAllowed: true,
			isFileshareFinallyAllowed:   false,
		},
		{
			name:                        "when mesh is off, fileshare was allowed in previous run, the state does not change until block succeeds",
			meshChecker:                 meshCheckerStub{isMeshnetOn: false},
			wasForbidCalled:             true,
			wasPermitCalled:             false,
			shouldRulesChangeFail:       true,
			isFileshareInitiallyAllowed: true,
			isFileshareFinallyAllowed:   true,
		},
	}

	for _, tt := range tests {
		rulesController := rulesControllerSpy{shouldFail: tt.shouldRulesChangeFail}
		job := monitorFileshareProcessJob{
			isFileshareAllowed: tt.isFileshareInitiallyAllowed,
			meshChecker:        tt.meshChecker,
			rulesController:    &rulesController,
			processChecker:     tt.processChecker,
		}
		assert.Equal(t, tt.isFileshareInitiallyAllowed, job.isFileshareAllowed)
		assert.False(t, rulesController.wasForbidCalled)
		assert.False(t, rulesController.wasPermitCalled)

		err := job.run()

		assert.Nil(t, err)
		assert.Equal(t, tt.isFileshareFinallyAllowed, job.isFileshareAllowed)
		assert.Equal(t, tt.wasForbidCalled, rulesController.wasForbidCalled)
		assert.Equal(t, tt.wasPermitCalled, rulesController.wasPermitCalled)
	}
}

type meshCheckerStub struct {
	isMeshnetOn bool
}

func (m meshCheckerStub) isMeshOn() bool {
	return m.isMeshnetOn
}

type rulesControllerSpy struct {
	wasForbidCalled bool
	wasPermitCalled bool
	shouldFail      bool
}

func (rc *rulesControllerSpy) ForbidFileshare() error {
	rc.wasForbidCalled = true
	if rc.shouldFail {
		return errors.New("intentional failure for testing")
	}
	return nil
}

func (rc *rulesControllerSpy) PermitFileshare() error {
	rc.wasPermitCalled = true
	if rc.shouldFail {
		return errors.New("intentional failure for testing")
	}
	return nil
}

type processCheckerStub struct {
	isFileshareUp bool
}

func (m processCheckerStub) isFileshareRunning() bool {
	return m.isFileshareUp
}
