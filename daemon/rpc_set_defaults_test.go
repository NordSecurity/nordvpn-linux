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
	"github.com/NordSecurity/nordvpn-linux/internal"
	netwpkg "github.com/NordSecurity/nordvpn-linux/networker"
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
						Killswitch:         &daemonevents.MockPublisherSubscriber[bool]{},
						Firewall:           &daemonevents.MockPublisherSubscriber[bool]{},
						Routing:            &daemonevents.MockPublisherSubscriber[bool]{},
						Autoconnect:        &daemonevents.MockPublisherSubscriber[bool]{},
						DNS:                &daemonevents.MockPublisherSubscriber[events.DataDNS]{},
						RealTimeProtection: &daemonevents.MockPublisherSubscriber[bool]{},
						Protocol:           &daemonevents.MockPublisherSubscriber[config.Protocol]{},
						Allowlist:          &daemonevents.MockPublisherSubscriber[events.DataAllowlist]{},
						Meshnet:            &daemonevents.MockPublisherSubscriber[bool]{},
						Technology:         &daemonevents.MockPublisherSubscriber[config.Technology]{},
						Obfuscate:          &daemonevents.MockPublisherSubscriber[bool]{},
						Notify:             &daemonevents.MockPublisherSubscriber[bool]{},
						LANDiscovery:       &daemonevents.MockPublisherSubscriber[bool]{},
						PostquantumVPN:     &daemonevents.MockPublisherSubscriber[bool]{},
						Defaults:           &daemonevents.MockPublisherSubscriber[any]{},
					},
				},
				pauseManager:       pauseSchedulerMock,
				connectionInfo:     connectionInfo,
				recentVPNConnStore: recents.NewRecentConnectionsStore("/test/path", &fs, nil),
			}

			if test.isDataDisconnectExpected {
				// simulate pause is activated
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

func TestSetDefaults_SyncsNetworkerFirewallState(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name                        string
		firewallDisabledBeforeReset bool
	}{
		{
			name:                        "firewall disabled before reset is re-enabled and killswitch is applied",
			firewallDisabledBeforeReset: true,
		},
		{
			name:                        "firewall enabled before reset stays enabled and killswitch is applied",
			firewallDisabledBeforeReset: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			netw := &networker.Mock{}
			rpc := testRPC()
			rpc.netw = netw

			if test.firewallDisabledBeforeReset {
				resp, err := rpc.SetFirewall(context.Background(), &pb.SetGenericRequest{Enabled: false})
				assert.NoError(t, err)
				assert.Equal(t, internal.CodeSuccess, resp.Type)
				assert.True(t, netw.FirewallDisabled)
			}

			resp, err := rpc.SetDefaults(context.Background(), &pb.SetDefaultsRequest{NoLogout: true})
			assert.NoError(t, err)
			assert.Equal(t, internal.CodeSuccess, resp.Type)
			assert.False(t, netw.FirewallDisabled)

			resp, err = rpc.SetKillSwitch(context.Background(), &pb.SetKillSwitchRequest{KillSwitch: true})
			assert.NoError(t, err)
			assert.Equal(t, internal.CodeSuccess, resp.Type)
			assert.True(t, netw.KillSwitchApplied)
		})
	}
}

func TestSetDefaults_SyncsNetworkerRoutingState(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name                       string
		routingDisabledBeforeReset bool
	}{
		{
			name:                       "routing disabled before reset is re-enabled",
			routingDisabledBeforeReset: true,
		},
		{
			name:                       "routing enabled before reset stays enabled",
			routingDisabledBeforeReset: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			netw := &networker.Mock{}
			rpc := testRPC()
			rpc.netw = netw

			if test.routingDisabledBeforeReset {
				// routing cannot be disabled while meshnet is enabled
				rpc.cm.(*mockConfigManager).c.Mesh = false

				resp, err := rpc.SetRouting(context.Background(), &pb.SetGenericRequest{Enabled: false})
				assert.NoError(t, err)
				assert.Equal(t, internal.CodeSuccess, resp.Type)
				assert.True(t, netw.RoutingDisabled)
			}

			resp, err := rpc.SetDefaults(context.Background(), &pb.SetDefaultsRequest{NoLogout: true})
			assert.NoError(t, err)
			assert.Equal(t, internal.CodeSuccess, resp.Type)
			assert.False(t, netw.RoutingDisabled)
		})
	}
}

func TestSetDefaults_AppliesSettingsToNetworker(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name         string
		applyErr     error
		expectedCode int64
	}{
		{
			name:         "settings applied",
			applyErr:     nil,
			expectedCode: internal.CodeSuccess,
		},
		{
			name:         "applying settings fails",
			applyErr:     mock.ErrOnPurpose,
			expectedCode: internal.CodeFailure,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			netw := &networker.Mock{ApplySettingsErr: test.applyErr}
			defaultsEvents := &daemonevents.MockPublisherSubscriber[any]{}
			rpc := testRPC()
			rpc.netw = netw
			rpc.events.Settings.Defaults = defaultsEvents

			resp, err := rpc.SetDefaults(context.Background(), &pb.SetDefaultsRequest{NoLogout: true})
			assert.NoError(t, err)
			assert.Equal(t, test.expectedCode, resp.Type)
			assert.Equal(t, &netwpkg.Settings{
				Firewall:     true,
				Routing:      true,
				LanDiscovery: false,
				ARPIgnore:    true,
				Allowlist:    config.Allowlist{},
			}, netw.AppliedSettings)
			// config is reset even if applying settings fails, so defaults event is always published
			assert.True(t, defaultsEvents.EventPublished)
		})
	}
}
