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
		name           string
		processChecker processChecker
	}{
		{
			name:           "fileshare is allowed when process is running",
			processChecker: processCheckerStub{isFileshareUp: true},
		},
		{
			name:           "fileshare is forbidden when process was stopped",
			processChecker: processCheckerStub{isFileshareUp: false},
		},
	}

	for _, tt := range tests {
		rulesController := rulesControllerSpy{}
		job := monitorFileshareProcessJob{
			rulesController: &rulesController,
			processChecker:  tt.processChecker,
		}
		assert.False(t, rulesController.wasForbidCalled)
		assert.False(t, rulesController.wasPermitCalled)

		err := job.run()

		assert.Nil(t, err)
		assert.Equal(t, !tt.processChecker.isFileshareRunning(), rulesController.wasForbidCalled)
		assert.Equal(t, tt.processChecker.isFileshareRunning(), rulesController.wasPermitCalled)
	}
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
