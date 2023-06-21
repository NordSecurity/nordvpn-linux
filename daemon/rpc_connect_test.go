package daemon

import (
	"context"
	"net/http"
	"net/netip"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/response"
	"github.com/NordSecurity/nordvpn-linux/daemon/routes"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/meshnet/mock"
	"github.com/NordSecurity/nordvpn-linux/networker"
	"github.com/NordSecurity/nordvpn-linux/request"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"
	"google.golang.org/grpc/metadata"
)

type mockRPCServer struct{}

func (mockRPCServer) SetHeader(metadata.MD) error  { return nil }
func (mockRPCServer) SendHeader(metadata.MD) error { return nil }
func (mockRPCServer) SetTrailer(metadata.MD)       {}
func (mockRPCServer) Context() context.Context     { return nil }
func (mockRPCServer) SendMsg(m interface{}) error  { return nil }
func (mockRPCServer) RecvMsg(m interface{}) error  { return nil }
func (mockRPCServer) Send(*pb.Payload) error       { return nil }

type mockNameservers []string

func (m mockNameservers) Get(bool, bool) []string {
	return m
}

type mockVault struct{}

func (mockVault) Get(string) (ssh.PublicKey, error) {
	return nil, nil
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

func (validCredentialsAPI) SetTransport(request.MetaTransport) {}

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

type workingLoginChecker struct{}

func (workingLoginChecker) IsLoggedIn() bool { return true }

type mockAnalytics struct{}

func (*mockAnalytics) Enable() error  { return nil }
func (*mockAnalytics) Disable() error { return nil }

func TestRpcConnect(t *testing.T) {
	category.Set(t, category.Route)

	defer testsCleanup()
	tests := []struct {
		name      string
		factory   FactoryFunc
		netw      networker.Networker
		retriever routes.GatewayRetriever
		fw        firewall.Service
	}{
		{
			name: "successfull connect",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &workingVPN{}, nil
			},
			netw:      workingNetworker{},
			retriever: newGatewayMock(netip.Addr{}),
			fw:        &workingFirewall{},
		},
		{
			name: "failed connect",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &failingVPN{}, nil
			},
			netw:      failingNetworker{},
			retriever: newGatewayMock(netip.Addr{}),
			fw:        &workingFirewall{},
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
				"1.0.0",
				"",
				internal.Development,
				&mockVault{},
				&request.HTTPClient{},
				response.ValidateResponseHeaders,
				&subs.Subject[events.DataRequestAPI]{},
			)
			rpc := NewRPC(
				internal.Development,
				workingLoginChecker{},
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
				request.NewHTTPClient(http.DefaultClient, "", nil, nil),
				NewEvents(
					&subs.Subject[bool]{},
					&subs.Subject[bool]{},
					&subs.Subject[events.DataDNS]{},
					&subs.Subject[bool]{},
					&subs.Subject[config.Protocol]{},
					&subs.Subject[events.DataWhitelist]{},
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
				mock.Fileshare{},
			)
			err := rpc.Connect(&pb.ConnectRequest{}, &mockRPCServer{})
			assert.NoError(t, err)
		})
	}
}

func TestRpcReconnect(t *testing.T) {
	category.Set(t, category.Route)

	var fail bool
	factory := func(config.Technology) (vpn.VPN, error) {
		if fail {
			fail = false
			return &failingVPN{}, nil
		}
		fail = true
		return &workingVPN{}, nil
	}

	cm := newMockConfigManager()
	tokenData := cm.c.TokensData[cm.c.AutoConnectData.ID]
	tokenData.TokenExpiry = time.Now().Add(time.Hour * 1).Format(internal.ServerDateFormat)
	tokenData.ServiceExpiry = time.Now().Add(time.Hour * 1).Format(internal.ServerDateFormat)
	cm.c.TokensData[cm.c.AutoConnectData.ID] = tokenData
	dm := testNewDataManager()
	api := core.NewDefaultAPI(
		"1.0.0",
		"",
		internal.Development,
		&mockVault{},
		&request.HTTPClient{},
		response.ValidateResponseHeaders,
		&subs.Subject[events.DataRequestAPI]{},
	)
	rpc := NewRPC(
		internal.Development,
		workingLoginChecker{},
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
		request.NewHTTPClient(http.DefaultClient, "", nil, nil),
		NewEvents(
			&subs.Subject[bool]{},
			&subs.Subject[bool]{},
			&subs.Subject[events.DataDNS]{},
			&subs.Subject[bool]{},
			&subs.Subject[config.Protocol]{},
			&subs.Subject[events.DataWhitelist]{},
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
		workingNetworker{},
		&subs.Subject[string]{},
		mockNameservers([]string{"1.1.1.1"}),
		nil,
		&mockAnalytics{},
		mock.Fileshare{},
	)
	err := rpc.Connect(&pb.ConnectRequest{}, &mockRPCServer{})
	assert.NoError(t, err)

	rpc.netw = failingNetworker{} // second connect has to fail
	err = rpc.Connect(&pb.ConnectRequest{}, &mockRPCServer{})
	assert.NoError(t, err)

	rpc.netw = workingNetworker{}
	err = rpc.Connect(&pb.ConnectRequest{}, &mockRPCServer{})
	assert.NoError(t, err)
}
