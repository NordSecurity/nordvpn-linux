package daemon

import (
	"errors"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	mockapi "github.com/NordSecurity/nordvpn-linux/test/mock/core"
	mockevents "github.com/NordSecurity/nordvpn-linux/test/mock/events"
	"github.com/stretchr/testify/assert"
)

func Test_NotifyDedicatedServerStateChange(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name                         string
		dedicatedServerStatus        core.DedicatedServerStatus
		serversListFetchErr          error
		shouldPublishEvent           bool
		expectedDedicatedServerState string
		shouldBeErr                  bool
	}{
		{
			name:                         "server status new",
			dedicatedServerStatus:        core.DedicatedServerStatusNew,
			shouldPublishEvent:           true,
			expectedDedicatedServerState: string(core.DedicatedServerStatusNew),
		},
		{
			name:                         "server status running",
			dedicatedServerStatus:        core.DedicatedServerStatusRunning,
			shouldPublishEvent:           true,
			expectedDedicatedServerState: string(core.DedicatedServerStatusRunning),
		},
		{
			name:                "fails to fetch dedicated servers list",
			shouldPublishEvent:  false,
			serversListFetchErr: errors.New("failed to fetch ds list"),
			shouldBeErr:         true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPublisherSubsriber := mockevents.NewMockPublisherSubscriber[events.DataDedicatedServerStatus]()
			mockDedicatedServersAPI := mockapi.DedicatedServersAPIMock{
				DedicatedServersResponse: core.DedicatedServers{
					core.DedicatedServer{Status: test.dedicatedServerStatus},
				},
				DedicatedServerErr: test.serversListFetchErr,
			}

			dedicatedServersStatePublisher := DedicatedServerState{
				dedicatedServerStatusPublisher: mockPublisherSubsriber,
				dedicatedServersAPI:            &mockDedicatedServersAPI,
			}

			err := dedicatedServersStatePublisher.NotifyDedicatedServerStateChange(struct{}{})
			assert.Equal(t, test.shouldPublishEvent, mockPublisherSubsriber.EventPublished,
				"Unexpected event published after dedicated server state change notification.")
			assert.Equal(t, test.expectedDedicatedServerState, mockPublisherSubsriber.Event.Status,
				"Invalid status in received dedicated server state notification.")
			if test.shouldBeErr {
				assert.NotNil(t, err, "Expected error but error was not returned.")
			}
		})
	}
}
