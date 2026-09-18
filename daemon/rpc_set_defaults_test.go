package daemon

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	daemonevents "github.com/NordSecurity/nordvpn-linux/daemon/events"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/recents"
	"github.com/NordSecurity/nordvpn-linux/daemon/state"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/session"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	testcore "github.com/NordSecurity/nordvpn-linux/test/mock/core"
	"github.com/NordSecurity/nordvpn-linux/test/mock/fs"

	testdevicekey "github.com/NordSecurity/nordvpn-linux/test/mock/devicekey"
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

func TestSetDefaults_ResetsNetworkerLanDiscoveryAndAllowlist(t *testing.T) {
	category.Set(t, category.Unit)

	netw := &networker.Mock{
		LanDiscovery: true,
		Allowlist: config.Allowlist{
			Ports: config.Ports{
				TCP: config.PortSet{80: true},
				UDP: config.PortSet{53: true},
			},
			Subnets: []string{"192.168.1.0/24"},
		},
	}

	rpc := testRPC()
	rpc.netw = netw

	_, err := rpc.SetDefaults(context.Background(), &pb.SetDefaultsRequest{NoLogout: true})
	assert.NoError(t, err)
	assert.False(t, netw.LanDiscovery)
	assert.Equal(t, config.Allowlist{}, netw.Allowlist)
}

// setDefaultsCredentialsAPIMock lets a single test case make token deletion fail. The shared mock
// in test/mock/core always succeeds.
type setDefaultsCredentialsAPIMock struct {
	testcore.CredentialsAPIMock
	deleteTokenErr error
}

func (c *setDefaultsCredentialsAPIMock) DeleteToken() error { return c.deleteTokenErr }

// TestSetDefaults_Logout covers `nordvpn set defaults --logout`, which reaches the daemon as the
// inverted NoLogout flag and now performs a full logout instead of only stopping the notification
// center client. Which status access.Logout returns for each kind of session is already covered by
// TestLogout_Token, so this test only asserts what SetDefaults adds on top: how that status is
// mapped to a payload code, and whether the settings reset still happens afterwards.
func TestSetDefaults_Logout(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name                string
		loggedIn            bool
		loggedInWithToken   bool
		deleteTokenErr      error
		expectedCode        int64
		expectSettingsReset bool
	}{
		{
			name:                "logs out a session started via Nord Account",
			loggedIn:            true,
			loggedInWithToken:   false,
			deleteTokenErr:      nil,
			expectedCode:        internal.CodeSuccess,
			expectSettingsReset: true,
		},
		{
			name:                "keeps a manually provided token and still reports success",
			loggedIn:            true,
			loggedInWithToken:   true,
			deleteTokenErr:      nil,
			expectedCode:        internal.CodeSuccess,
			expectSettingsReset: true,
		},
		{
			name:                "resets the settings when nobody is logged in",
			loggedIn:            false,
			loggedInWithToken:   false,
			deleteTokenErr:      nil,
			expectedCode:        internal.CodeSuccess,
			expectSettingsReset: true,
		},
		{
			name:                "keeps the settings when the logout fails",
			loggedIn:            true,
			loggedInWithToken:   false,
			deleteTokenErr:      core.ErrServerInternal,
			expectedCode:        internal.CodeInternalError,
			expectSettingsReset: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfgManagerMock := newMockConfigManager()
			err := cfgManagerMock.SaveWith(func(c config.Config) config.Config {
				tokenData := c.TokensData[c.AutoConnectData.ID]
				tokenData.Token = "1234"
				if test.loggedInWithToken {
					tokenData.RenewToken = ""
					tokenData.TokenExpiry = session.ManualAccessTokenExpiryDateString
				} else {
					tokenData.RenewToken = "1234"
				}
				c.TokensData[c.AutoConnectData.ID] = tokenData
				return c
			})
			assert.NoError(t, err)

			var loginChecker auth.Checker = &workingLoginChecker{}
			if !test.loggedIn {
				loginChecker = failingLoginChecker{}
			}

			logoutEvent := &daemonevents.MockPublisherSubscriber[events.DataAuthorization]{}
			daemonEvents := daemonevents.NewEventsEmpty()
			daemonEvents.User.Logout = logoutEvent

			fs := fs.NewSystemFileHandleMock(t)

			netw := &networker.Mock{
				LanDiscovery: true,
				Allowlist:    config.Allowlist{Subnets: []string{"192.168.1.0/24"}},
			}

			rpc := RPC{
				ac:                        loginChecker,
				cm:                        cfgManagerMock,
				norduser:                  &testnorduser.MockNorduserCombinedService{},
				netw:                      netw,
				ncClient:                  &mock.NotificationClientMock{},
				publisher:                 &subs.Subject[string]{},
				credentialsAPI:            &setDefaultsCredentialsAPIMock{deleteTokenErr: test.deleteTokenErr},
				factory:                   func(config.Technology) (vpn.VPN, error) { return nil, nil },
				events:                    daemonEvents,
				recentVPNConnStore:        recents.NewRecentConnectionsStore("/test/path", &fs, nil),
				pauseManager:              &mock.PauseSchedulerMock{},
				connectionInfo:            state.NewConnectionInfo(),
				dedicatedServerKeyManager: &testdevicekey.MockDeviceKeyManager{},
			}

			resp, err := rpc.SetDefaults(context.Background(), &pb.SetDefaultsRequest{NoLogout: false})
			assert.NoError(t, err)
			assert.Equal(t, test.expectedCode, resp.Type)
			// The settings reset wipes everything else the logout touched, so the event is the
			// only trace left of it having run.
			assert.Equal(t, test.loggedIn, logoutEvent.EventPublished, "unexpected logout attempt")

			if test.expectSettingsReset {
				assert.False(t, netw.LanDiscovery)
				assert.Equal(t, config.Allowlist{}, netw.Allowlist)
			} else {
				assert.True(t, netw.LanDiscovery, "settings were reset despite a failed logout")
			}
		})
	}
}
