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
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/response"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/fileshare/service"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/networker"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	testnetworker "github.com/NordSecurity/nordvpn-linux/test/mock/networker"

	"github.com/stretchr/testify/assert"
)

type mockRPCServer struct {
	pb.Daemon_ConnectServer
	msg *pb.Payload
}

func (m *mockRPCServer) Send(p *pb.Payload) error { m.msg = p; return nil }

type mockNameservers []string

func (m mockNameservers) Get(bool, bool) []string {
	return m
}

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
	isVPNExpired bool
	vpnErr       error
}

func (*workingLoginChecker) IsLoggedIn() bool              { return true }
func (c *workingLoginChecker) IsVPNExpired() (bool, error) { return c.isVPNExpired, c.vpnErr }

type mockAnalytics struct{}

func (*mockAnalytics) Enable() error  { return nil }
func (*mockAnalytics) Disable() error { return nil }

func TestRpcConnect(t *testing.T) {
	category.Set(t, category.Route)

	defer testsCleanup()
	tests := []struct {
		name    string
		factory FactoryFunc
		netw    networker.Networker
		fw      firewall.Service
		checker auth.Checker
		resp    int64
	}{
		{
			name: "successfull connect",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},
			netw:    &testnetworker.Mock{},
			fw:      &workingFirewall{},
			checker: &workingLoginChecker{},
			resp:    internal.CodeConnected,
		},
		{
			name: "failed connect",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.FailingVPN{}, nil
			},
			netw:    testnetworker.Failing{},
			fw:      &workingFirewall{},
			checker: &workingLoginChecker{},
			resp:    internal.CodeFailure,
		},
		{
			name: "VPN expired",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},
			netw:    &testnetworker.Mock{},
			fw:      &workingFirewall{},
			checker: &workingLoginChecker{isVPNExpired: true},
			resp:    internal.CodeAccountExpired,
		},
		{
			name: "VPN expiration check fails",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},
			netw:    &testnetworker.Mock{},
			fw:      &workingFirewall{},
			checker: &workingLoginChecker{vpnErr: errors.New("test error")},
			resp:    internal.CodeTokenRenewError,
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
				NewEvents(
					&subs.Subject[bool]{},
					&subs.Subject[bool]{},
					&subs.Subject[events.DataDNS]{},
					&subs.Subject[bool]{},
					&subs.Subject[config.Protocol]{},
					&subs.Subject[events.DataAllowlist]{},
					&subs.Subject[config.Technology]{},
					&subs.Subject[bool]{},
					&subs.Subject[bool]{},
					&subs.Subject[bool]{},
					&subs.Subject[bool]{},
					&subs.Subject[bool]{},
					&subs.Subject[bool]{},
					&subs.Subject[bool]{},
					&subs.Subject[any]{},
					&subs.Subject[events.DataConnect]{},
					&subs.Subject[events.DataDisconnect]{},
					&subs.Subject[any]{},
					&subs.Subject[core.ServicesResponse]{},
					&subs.Subject[events.ServerRating]{},
					&subs.Subject[int]{},
				),
				test.factory,
				newEndpointResolverMock(netip.MustParseAddr("127.0.0.1")),
				test.netw,
				&subs.Subject[string]{},
				mockNameservers([]string{"1.1.1.1"}),
				nil,
				&mockAnalytics{},
				service.NoopFileshare{},
				&RegistryMock{},
			)
			server := &mockRPCServer{}
			err := rpc.Connect(&pb.ConnectRequest{}, server)
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
		NewEvents(
			&subs.Subject[bool]{},
			&subs.Subject[bool]{},
			&subs.Subject[events.DataDNS]{},
			&subs.Subject[bool]{},
			&subs.Subject[config.Protocol]{},
			&subs.Subject[events.DataAllowlist]{},
			&subs.Subject[config.Technology]{},
			&subs.Subject[bool]{},
			&subs.Subject[bool]{},
			&subs.Subject[bool]{},
			&subs.Subject[bool]{},
			&subs.Subject[bool]{},
			&subs.Subject[bool]{},
			&subs.Subject[bool]{},
			&subs.Subject[any]{},
			&subs.Subject[events.DataConnect]{},
			&subs.Subject[events.DataDisconnect]{},
			&subs.Subject[any]{},
			&subs.Subject[core.ServicesResponse]{},
			&subs.Subject[events.ServerRating]{},
			&subs.Subject[int]{},
		),
		factory,
		newEndpointResolverMock(netip.MustParseAddr("127.0.0.1")),
		&testnetworker.Mock{},
		&subs.Subject[string]{},
		mockNameservers([]string{"1.1.1.1"}),
		nil,
		&mockAnalytics{},
		service.NoopFileshare{},
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
