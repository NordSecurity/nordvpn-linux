package daemon

import (
	"errors"
	"fmt"
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
}

func (*workingLoginChecker) IsLoggedIn() bool              { return true }
func (c *workingLoginChecker) IsVPNExpired() (bool, error) { return c.isVPNExpired, c.vpnErr }
func (*workingLoginChecker) GetDedicatedIPServices() ([]auth.DedicatedIPService, error) {
	return nil, fmt.Errorf("Not implemented")
}

type mockAnalytics struct{}

func (*mockAnalytics) Enable() error  { return nil }
func (*mockAnalytics) Disable() error { return nil }

func TestRpcConnect(t *testing.T) {
	category.Set(t, category.Route)

	defer testsCleanup()
	tests := []struct {
		name        string
		serverGroup string
		factory     FactoryFunc
		netw        networker.Networker
		fw          firewall.Service
		checker     auth.Checker
		resp        int64
	}{
		{
			name:        "successful connect",
			serverGroup: "",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},
			netw:    &testnetworker.Mock{},
			fw:      &workingFirewall{},
			checker: &workingLoginChecker{},
			resp:    internal.CodeConnected,
		},
		{
			name:        "failed connect",
			serverGroup: "",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.FailingVPN{}, nil
			},
			netw:    testnetworker.Failing{},
			fw:      &workingFirewall{},
			checker: &workingLoginChecker{},
			resp:    internal.CodeFailure,
		},
		{
			name:        "VPN expired",
			serverGroup: "",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},
			netw:    &testnetworker.Mock{},
			fw:      &workingFirewall{},
			checker: &workingLoginChecker{isVPNExpired: true},
			resp:    internal.CodeAccountExpired,
		},
		{
			name:        "VPN expiration check fails",
			serverGroup: "",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},
			netw:    &testnetworker.Mock{},
			fw:      &workingFirewall{},
			checker: &workingLoginChecker{vpnErr: errors.New("test error")},
			resp:    internal.CodeTokenRenewError,
		},
		{
			name:        "Dedicated IP succesfull connect",
			serverGroup: "Dedicated_IP",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},
			netw:    &testnetworker.Mock{},
			fw:      &workingFirewall{},
			checker: &workingLoginChecker{isDedicatedIPExpired: false},
			resp:    internal.CodeConnected,
		},
		{
			name:        "Dedicated IP expired",
			serverGroup: "Dedicated_IP",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},
			netw:    &testnetworker.Mock{},
			fw:      &workingFirewall{},
			checker: &workingLoginChecker{isDedicatedIPExpired: true},
			resp:    internal.CodeDedicatedIPRenewError,
		},
		{
			name:        "Dedicated IP check fails",
			serverGroup: "Dedicated_IP",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},
			netw:    &testnetworker.Mock{},
			fw:      &workingFirewall{},
			checker: &workingLoginChecker{isDedicatedIPExpired: true, dedicatedIPErr: errors.New("test error")},
			resp:    internal.CodeDedicatedIPRenewError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
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
				test.checker,
				cm,
				dm,
				api,
				&mockServersAPI{},
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
			err := rpc.Connect(&pb.ConnectRequest{ServerGroup: "Dedicated_IP"}, server)
			assert.NoError(t, err)
			assert.Equal(t, server.msg.Type, test.resp)
		})
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
