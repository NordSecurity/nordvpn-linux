package daemon

import (
	"context"
	"errors"
	"testing"
	"time"

	daemonEvents "github.com/NordSecurity/nordvpn-linux/daemon/events"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/recents"
	"github.com/NordSecurity/nordvpn-linux/daemon/state"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	testnetworker "github.com/NordSecurity/nordvpn-linux/test/mock/networker"
	"gotest.tools/v3/assert"
)

func TestPauseConnection(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name                  string
		isVPNActive           bool
		pauseDuration         int
		disconnectErr         error
		isMeshPeer            bool
		expectedResponseType  int64
		expectedPauseDuration time.Duration
		expectedState         pb.ConnectionState
		expectedVPNState      bool
	}{
		{
			name:                 "VPN is not active, pause returns nothing to do",
			isVPNActive:          false,
			expectedResponseType: internal.CodeNothingToDo,
			expectedVPNState:     false,
		},
		{
			name:                  "VPN is active, VPN is disconnected after pause 10sec",
			isVPNActive:           true,
			pauseDuration:         10,
			expectedResponseType:  internal.CodeSuccess,
			expectedPauseDuration: 10 * time.Second,
			expectedVPNState:      false,
		},
		{
			name:                  "VPN is active, VPN is disconnected after pause 20sec",
			isVPNActive:           true,
			pauseDuration:         20,
			expectedResponseType:  internal.CodeSuccess,
			expectedPauseDuration: 20 * time.Second,
			expectedVPNState:      false,
		},
		{
			name:                 "disconnect failure",
			isVPNActive:          true,
			disconnectErr:        errors.New("failed to disconnect"),
			expectedResponseType: internal.CodeFailure,
			expectedVPNState:     true,
		},
		{
			name:                 "attempt to pause mesh connection",
			isVPNActive:          true,
			isMeshPeer:           true,
			disconnectErr:        errors.New("failed to disconnect"),
			expectedResponseType: internal.CodePauseAttemptWhenConnectedToMeshPeer,
			expectedVPNState:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			networkerMock := testnetworker.Mock{}
			networkerMock.VpnActive = test.isVPNActive
			networkerMock.StopErr = test.disconnectErr

			pauseSchedulerMock := &mock.PauseSchedulerMock{}

			connectionInfo := state.NewConnectionInfo()
			connectionInfo.ConnectionStatusNotifyConnect(events.DataConnect{IsMeshnetPeer: test.isMeshPeer})

			r := RPC{
				netw:               &networkerMock,
				cm:                 newMockConfigManager(),
				events:             daemonEvents.NewEventsEmpty(),
				recentVPNConnStore: recents.NewRecentConnectionsStore(TestdataPath+TestRecentConnFile, &internal.StdFilesystemHandle{}, nil),
				pauseManager:       pauseSchedulerMock,
				connectionInfo:     connectionInfo,
			}

			response, err := r.PauseConnection(context.Background(), &pb.PauseRequest{Seconds: int64(test.pauseDuration)})
			assert.NilError(t, err, "Unexpected error returned by PauseConnection RPC.")
			assert.Equal(t, test.expectedResponseType, response.Type, "Unexpected response type returned by pause RPC.")
			assert.Equal(t, test.expectedPauseDuration, pauseSchedulerMock.PauseDuration, "Unexpected pause duration.")
			assert.Equal(t, test.expectedVPNState, networkerMock.VpnActive, "Unexpected VPN activity status.")
		})
	}
}
