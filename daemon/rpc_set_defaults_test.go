package daemon

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/NordSecurity/nordvpn-linux/config"
	daemonevents "github.com/NordSecurity/nordvpn-linux/daemon/events"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/recents"
	"github.com/NordSecurity/nordvpn-linux/daemon/state"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	testcore "github.com/NordSecurity/nordvpn-linux/test/mock/core"
	"github.com/NordSecurity/nordvpn-linux/test/mock/fs"

	"github.com/NordSecurity/nordvpn-linux/test/mock/networker"
	testnorduser "github.com/NordSecurity/nordvpn-linux/test/mock/norduser/service"
)

func TestResetToDefaults_PauseVariants(t *testing.T) {
	category.Set(t, category.Integration)
	cfgManagerMock := newMockConfigManager()
	fs := fs.NewSystemFileHandleMock(t)
	pauseSchedulerMock := &mock.PauseSchedulerMock{}

	tests := []struct {
		name                     string
		isDataDisconnectExpected bool
	}{
		{
			name:                     "reset to defaults while pause active, empty disconnect event shall be emitted",
			isDataDisconnectExpected: true,
		},
		{
			name:                     "reset to defaults while no pause active, empty disconnect event shall not be emitted",
			isDataDisconnectExpected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockedDisconnectEvents := &daemonevents.MockPublisherSubscriber[events.DataDisconnect]{}
			connectionInfo := state.NewConnectionInfo()

			rpc := RPC{
				ac:             &workingLoginChecker{},
				cm:             cfgManagerMock,
				norduser:       &testnorduser.MockNorduserCombinedService{},
				netw:           &networker.Mock{},
				ncClient:       &mock.NotificationClientMock{},
				publisher:      &subs.Subject[string]{},
				credentialsAPI: &testcore.CredentialsAPIMock{},
				factory:        func(t config.Technology) (vpn.VPN, error) { return nil, nil },
				events: &daemonevents.Events{
					User:    &daemonevents.LoginEvents{Logout: &daemonevents.MockPublisherSubscriber[events.DataAuthorization]{}},
					Service: &daemonevents.ServiceEvents{Disconnect: mockedDisconnectEvents},
					Settings: &daemonevents.SettingsEvents{
						Killswitch:           &daemonevents.MockPublisherSubscriber[bool]{},
						Firewall:             &daemonevents.MockPublisherSubscriber[bool]{},
						Routing:              &daemonevents.MockPublisherSubscriber[bool]{},
						Autoconnect:          &daemonevents.MockPublisherSubscriber[bool]{},
						DNS:                  &daemonevents.MockPublisherSubscriber[events.DataDNS]{},
						ThreatProtectionLite: &daemonevents.MockPublisherSubscriber[bool]{},
						Protocol:             &daemonevents.MockPublisherSubscriber[config.Protocol]{},
						Allowlist:            &daemonevents.MockPublisherSubscriber[events.DataAllowlist]{},
						Meshnet:              &daemonevents.MockPublisherSubscriber[bool]{},
						Technology:           &daemonevents.MockPublisherSubscriber[config.Technology]{},
						Obfuscate:            &daemonevents.MockPublisherSubscriber[bool]{},
						Notify:               &daemonevents.MockPublisherSubscriber[bool]{},
						LANDiscovery:         &daemonevents.MockPublisherSubscriber[bool]{},
						VirtualLocation:      &daemonevents.MockPublisherSubscriber[bool]{},
						PostquantumVPN:       &daemonevents.MockPublisherSubscriber[bool]{},
						Defaults:             &daemonevents.MockPublisherSubscriber[any]{},
					},
				},
				pauseManager:       pauseSchedulerMock,
				connectionInfo:     connectionInfo,
				recentVPNConnStore: recents.NewRecentConnectionsStore("/test/path", &fs, nil),
			}

			if test.isDataDisconnectExpected {
				//simulate pause is activated
				connectionInfo.Pause(time.Now(), time.Second*60*5)
			}
			// actual response code is not relevant for this test
			_, err := rpc.SetDefaults(context.Background(), &pb.SetDefaultsRequest{NoLogout: true, OffKillswitch: false})
			assert.NoError(t, err)
			assert.Equal(t, test.isDataDisconnectExpected, mockedDisconnectEvents.EventPublished)
		})
	}
}
