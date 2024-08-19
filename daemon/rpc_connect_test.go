package daemon

import (
	"errors"
	"net/http"
	"net/netip"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	daemonevents "github.com/NordSecurity/nordvpn-linux/daemon/events"
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/response"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/networker"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	testnetworker "github.com/NordSecurity/nordvpn-linux/test/mock/networker"
	testnorduser "github.com/NordSecurity/nordvpn-linux/test/mock/norduser/service"

	"github.com/stretchr/testify/assert"
)

type mockRPCServer struct {
	pb.Daemon_ConnectServer
	msg *pb.Payload
}

func (m *mockRPCServer) Send(p *pb.Payload) error { m.msg = p; return nil }

type mockAuthenticationAPI struct{}

func (mockAuthenticationAPI) Login() (string, error) {
	return "", nil
}

func (mockAuthenticationAPI) Token(string) (*core.LoginResponse, error) {
	return nil, nil
}

type validCredentialsAPI struct {
	renewToken string
}

func (validCredentialsAPI) NotificationCredentials(token, appUserID string) (core.NotificationCredentialsResponse, error) {
	return core.NotificationCredentialsResponse{}, nil
}

func (v validCredentialsAPI) NotificationCredentialsRevoke(token, appUserID string, purgeSession bool) (core.NotificationCredentialsRevokeResponse, error) {
	return core.NotificationCredentialsRevokeResponse{}, nil
}

func (validCredentialsAPI) ServiceCredentials(string) (*core.CredentialsResponse, error) {
	return &core.CredentialsResponse{
		NordlynxPrivateKey: "nordpriv",
		Username:           "elite",
		Password:           "hacker",
	}, nil
}

func (v validCredentialsAPI) TokenRenew(renewToken string) (*core.TokenRenewResponse, error) {
	return &core.TokenRenewResponse{
		RenewToken: v.renewToken,
	}, nil
}

func (validCredentialsAPI) DeleteToken(token string) error {
	return nil
}

func (validCredentialsAPI) TrustedPassToken(token string) (*core.TrustedPassTokenResponse, error) {
	return nil, nil
}

func (validCredentialsAPI) MultiFactorAuthStatus(token string) (*core.MultiFactorAuthStatusResponse, error) {
	return nil, nil
}

func (validCredentialsAPI) Services(string) (core.ServicesResponse, error) {
	return core.ServicesResponse{
		{
			ExpiresAt: "2029-12-27 00:00:00",
			Service:   core.Service{ID: 1},
		},
	}, nil
}

func (validCredentialsAPI) CurrentUser(string) (*core.CurrentUserResponse, error) {
	return nil, nil
}

type workingLoginChecker struct {
	isVPNExpired         bool
	vpnErr               error
	isDedicatedIPExpired bool
	dedicatedIPErr       error
	dedicatedIPService   []auth.DedicatedIPService
}

func (*workingLoginChecker) IsLoggedIn() bool              { return true }
func (*workingLoginChecker) IsMFAEnabled() (bool, error)   { return false, nil }
func (c *workingLoginChecker) IsVPNExpired() (bool, error) { return c.isVPNExpired, c.vpnErr }
func (c *workingLoginChecker) GetDedicatedIPServices() ([]auth.DedicatedIPService, error) {
	if c.isDedicatedIPExpired {
		return nil, nil
	}

	if c.dedicatedIPErr != nil {
		return nil, c.dedicatedIPErr
	}

	return c.dedicatedIPService, nil
}

type mockAnalytics struct{}

func (*mockAnalytics) Enable() error  { return nil }
func (*mockAnalytics) Disable() error { return nil }

func TestRpcConnect(t *testing.T) {
	category.Set(t, category.Unit)

	defer testsCleanup()
	tests := []struct {
		name          string
		serverGroup   string
		serverTag     string
		factory       FactoryFunc
		netw          networker.Networker
		fw            firewall.Service
		checker       auth.Checker
		resp          int64
		expectedError error
	}{
		{
			name: "Quick connect works",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},
			netw:    &testnetworker.Mock{},
			fw:      &workingFirewall{},
			checker: &workingLoginChecker{},
			resp:    internal.CodeConnected,
		},
		{
			name: "Fail for broken Networker and VPN",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.FailingVPN{}, nil
			},
			netw:    testnetworker.Failing{},
			fw:      &workingFirewall{},
			checker: &workingLoginChecker{},
			resp:    internal.CodeFailure,
		},
		{
			name: "fFail when VPN subscription is expired",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},
			netw:    &testnetworker.Mock{},
			fw:      &workingFirewall{},
			checker: &workingLoginChecker{isVPNExpired: true},
			resp:    internal.CodeAccountExpired,
		},
		{
			name: "Fail when VPN subscription API calls fails",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},
			netw:    &testnetworker.Mock{},
			fw:      &workingFirewall{},
			checker: &workingLoginChecker{vpnErr: errors.New("test error")},
			resp:    internal.CodeTokenRenewError,
		},
		{
			name:      "Connects using country name",
			serverTag: "germany",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},
			netw:    &testnetworker.Mock{},
			fw:      &workingFirewall{},
			checker: &workingLoginChecker{},
			resp:    internal.CodeConnected,
		},
		{
			name:      "Connects using country name + city name",
			serverTag: "germany berlin",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},
			netw:    &testnetworker.Mock{},
			fw:      &workingFirewall{},
			checker: &workingLoginChecker{},
			resp:    internal.CodeConnected,
		},
		{
			name:      "Connects for city name",
			serverTag: "berlin",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},
			netw:    &testnetworker.Mock{},
			fw:      &workingFirewall{},
			checker: &workingLoginChecker{},
			resp:    internal.CodeConnected,
		},
		{
			name:      "Connects using country code + city name",
			serverTag: "de berlin",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},
			netw:    &testnetworker.Mock{},
			fw:      &workingFirewall{},
			checker: &workingLoginChecker{},
			resp:    internal.CodeConnected,
		},
		{
			name:      "Connects using country code",
			serverTag: "de",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},
			netw:    &testnetworker.Mock{},
			fw:      &workingFirewall{},
			checker: &workingLoginChecker{},
			resp:    internal.CodeConnected,
		},
		{
			name:        "Dedicated IP group connect works",
			serverGroup: "Dedicated_IP",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},
			netw: &testnetworker.Mock{},
			fw:   &workingFirewall{},
			checker: &workingLoginChecker{
				isDedicatedIPExpired: false,
				dedicatedIPService:   []auth.DedicatedIPService{{ExpiresAt: "", ServerID: 7}},
			},
			resp: internal.CodeConnected,
		},
		{
			name:      "Dedicated IP with server name works",
			serverTag: "lt7",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},
			netw: &testnetworker.Mock{},
			fw:   &workingFirewall{},
			checker: &workingLoginChecker{
				isDedicatedIPExpired: false,
				dedicatedIPService:   []auth.DedicatedIPService{{ExpiresAt: "", ServerID: 7}},
			},
			resp: internal.CodeConnected,
		},
		{
			name:      "fails when Dedicated IP subscription is expired",
			serverTag: "lt7",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},
			netw:    &testnetworker.Mock{},
			fw:      &workingFirewall{},
			checker: &workingLoginChecker{isDedicatedIPExpired: true},
			resp:    internal.CodeDedicatedIPRenewError,
		},
		{
			name:      "fails for Dedicated IP when API fails",
			serverTag: "lt7",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},
			netw:    &testnetworker.Mock{},
			fw:      &workingFirewall{},
			checker: &workingLoginChecker{isDedicatedIPExpired: true},
			resp:    internal.CodeDedicatedIPRenewError,
		},
		{
			name:      "fails when server not into Dedicated IP servers list",
			serverTag: "lt8",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},
			netw: &testnetworker.Mock{},
			fw:   &workingFirewall{},
			checker: &workingLoginChecker{
				isDedicatedIPExpired: false,
				dedicatedIPService:   []auth.DedicatedIPService{{ExpiresAt: "", ServerID: 7}},
			},
			resp: internal.CodeDedicatedIPNoServer,
		},
		{
			name:      "fails because Dedicated IP servers list is empty",
			serverTag: "lt7",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},
			netw: &testnetworker.Mock{},
			fw:   &workingFirewall{},
			checker: &workingLoginChecker{
				isDedicatedIPExpired: false,
				dedicatedIPService:   []auth.DedicatedIPService{{ExpiresAt: "", ServerID: auth.NoServerSelected}},
			},
			resp: internal.CodeDedicatedIPServiceButNoServers,
		},
	}

	for _, test := range tests {
		// run each test using working API for servers list and using local cached servers list
		servers := map[string]core.ServersAPI{
			"Remote": mockServersAPI{},
			"Local":  mockFailingServersAPI{},
		}
		for key, serversAPI := range servers {
			t.Run(test.name+" "+key, func(t *testing.T) {
				cm := newMockConfigManager()
				tokenData := cm.c.TokensData[cm.c.AutoConnectData.ID]
				tokenData.TokenExpiry = time.Now().Add(time.Hour * 1).Format(internal.ServerDateFormat)
				tokenData.ServiceExpiry = time.Now().Add(time.Hour * 1).Format(internal.ServerDateFormat)
				cm.c.TokensData[cm.c.AutoConnectData.ID] = tokenData
				dm := testNewDataManager()
				dm.SetServersData(time.Now(), serversList(), "")
				api := core.NewDefaultAPI(
					"",
					"",
					http.DefaultClient,
					response.NoopValidator{},
				)
				rpc := NewRPC(
					internal.Development,
					test.checker,
					cm,
					dm,
					api,
					serversAPI,
					&validCredentialsAPI{},
					testNewCDNAPI(),
					testNewRepoAPI(),
					&mockAuthenticationAPI{},
					"1.0.0",
					test.fw,
					daemonevents.NewEventsEmpty(),
					test.factory,
					newEndpointResolverMock(netip.MustParseAddr("127.0.0.1")),
					test.netw,
					&subs.Subject[string]{},
					&mock.DNSGetter{Names: []string{"1.1.1.1"}},
					nil,
					&mockAnalytics{},
					&testnorduser.MockNorduserCombinedService{},
					&RegistryMock{},
				)
				server := &mockRPCServer{}
				err := rpc.Connect(&pb.ConnectRequest{ServerGroup: test.serverGroup, ServerTag: test.serverTag}, server)
				assert.Equal(t, test.expectedError, err)
				if err == nil {
					assert.Equal(t, test.resp, server.msg.Type)
				}
			})
		}
	}
}

func TestRpcReconnect(t *testing.T) {
	category.Set(t, category.Route)

	var fail bool
	factory := func(config.Technology) (vpn.VPN, error) {
		if fail {
			fail = false
			return &mock.FailingVPN{}, nil
		}
		fail = true
		return &mock.WorkingVPN{}, nil
	}

	cm := newMockConfigManager()
	tokenData := cm.c.TokensData[cm.c.AutoConnectData.ID]
	tokenData.TokenExpiry = time.Now().Add(time.Hour * 1).Format(internal.ServerDateFormat)
	tokenData.ServiceExpiry = time.Now().Add(time.Hour * 1).Format(internal.ServerDateFormat)
	cm.c.TokensData[cm.c.AutoConnectData.ID] = tokenData
	dm := testNewDataManager()
	api := core.NewDefaultAPI(
		"",
		"",
		http.DefaultClient,
		response.NoopValidator{},
	)
	rpc := NewRPC(
		internal.Development,
		&workingLoginChecker{},
		cm,
		dm,
		api,
		&mockServersAPI{},
		&validCredentialsAPI{},
		testNewCDNAPI(),
		testNewRepoAPI(),
		&mockAuthenticationAPI{},
		"1.0.0",
		&workingFirewall{},
		daemonevents.NewEventsEmpty(),
		factory,
		newEndpointResolverMock(netip.MustParseAddr("127.0.0.1")),
		&testnetworker.Mock{},
		&subs.Subject[string]{},
		&mock.DNSGetter{Names: []string{"1.1.1.1"}},
		nil,
		&mockAnalytics{},
		&testnorduser.MockNorduserCombinedService{},
		&RegistryMock{},
	)
	err := rpc.Connect(&pb.ConnectRequest{}, &mockRPCServer{})
	assert.NoError(t, err)

	rpc.netw = testnetworker.Failing{} // second connect has to fail
	err = rpc.Connect(&pb.ConnectRequest{}, &mockRPCServer{})
	assert.NoError(t, err)

	rpc.netw = &testnetworker.Mock{}
	err = rpc.Connect(&pb.ConnectRequest{}, &mockRPCServer{})
	assert.NoError(t, err)
}
