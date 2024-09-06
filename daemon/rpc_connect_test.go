package daemon

import (
	"errors"
	"net/netip"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/internal"
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

type mockAuthenticationAPI struct{}

func (mockAuthenticationAPI) Login() (string, error) {
	return "", nil
}

func (mockAuthenticationAPI) Token(string) (*core.LoginResponse, error) {
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

type mockEndpointResolver struct{ ip netip.Addr }

func newEndpointResolverMock(ip netip.Addr) mockEndpointResolver {
	return mockEndpointResolver{ip: ip}
}

func (g mockEndpointResolver) Resolve(netip.Addr) ([]netip.Addr, error) {
	return []netip.Addr{g.ip}, nil
}

func TestRpcConnect(t *testing.T) {
	category.Set(t, category.Unit)

	defer testsCleanup()
	tests := []struct {
		name        string
		serverGroup string
		serverTag   string
		factory     FactoryFunc
		resp        int64
		setup       func(*RPC)
	}{
		{
			name: "Quick connect works",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},
			resp: internal.CodeConnected,
		},
		{
			name: "Fail for broken Networker and VPN",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.FailingVPN{}, nil
			},
			setup: func(rpc *RPC) {
				rpc.netw = testnetworker.Failing{}
			},
			resp: internal.CodeFailure,
		},
		{
			name: "Fail when VPN subscription is expired",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},
			setup: func(rpc *RPC) {
				rpc.ac = &workingLoginChecker{isVPNExpired: true}
			},
			resp: internal.CodeAccountExpired,
		},
		{
			name: "Fail when VPN subscription API calls fails",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},
			setup: func(rpc *RPC) {
				rpc.ac = &workingLoginChecker{vpnErr: errors.New("test error")}
			},
			resp: internal.CodeTokenRenewError,
		},
		{
			name:      "Connects using country name",
			serverTag: "germany",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},
			resp: internal.CodeConnected,
		},
		{
			name:      "Connects using country name + city name",
			serverTag: "germany berlin",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},
			resp: internal.CodeConnected,
		},
		{
			name:      "Connects for city name",
			serverTag: "berlin",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},
			resp: internal.CodeConnected,
		},
		{
			name:      "Connects using country code + city name",
			serverTag: "de berlin",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},
			resp: internal.CodeConnected,
		},
		{
			name:      "Connects using country code",
			serverTag: "de",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},
			resp: internal.CodeConnected,
		},
		{
			name:        "Dedicated IP group connect works",
			serverGroup: "Dedicated_IP",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},
			setup: func(rpc *RPC) {
				rpc.ac = &workingLoginChecker{
					isDedicatedIPExpired: false,
					dedicatedIPService: []auth.DedicatedIPService{
						{ExpiresAt: "", ServerID: 7},
					},
				}
			},
			resp: internal.CodeConnected,
		},
		{
			name:      "Dedicated IP with server name works",
			serverTag: "lt7",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},
			setup: func(rpc *RPC) {
				rpc.ac = &workingLoginChecker{
					isDedicatedIPExpired: false,
					dedicatedIPService: []auth.DedicatedIPService{
						{ExpiresAt: "", ServerID: 7},
					},
				}
			},
			resp: internal.CodeConnected,
		},
		{
			name:      "fails when Dedicated IP subscription is expired",
			serverTag: "lt7",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},
			setup: func(rpc *RPC) {
				rpc.ac = &workingLoginChecker{isDedicatedIPExpired: true}
			},
			resp: internal.CodeDedicatedIPRenewError,
		},
		{
			name:      "fails for Dedicated IP when API fails",
			serverTag: "lt7",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},
			setup: func(rpc *RPC) {
				rpc.ac = &workingLoginChecker{
					dedicatedIPErr: errors.New("error"),
				}
			},
		},
		{
			name:      "fails when server not into Dedicated IP servers list",
			serverTag: "lt8",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},

			setup: func(rpc *RPC) {
				rpc.ac = &workingLoginChecker{
					isDedicatedIPExpired: false,
					dedicatedIPService: []auth.DedicatedIPService{
						{ExpiresAt: "", ServerID: 7},
					},
				}
			},
			resp: internal.CodeDedicatedIPNoServer,
		},
		{
			name:      "fails because Dedicated IP servers list is empty",
			serverTag: "lt7",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},
			setup: func(rpc *RPC) {
				rpc.ac = &workingLoginChecker{
					isDedicatedIPExpired: false,
					dedicatedIPService: []auth.DedicatedIPService{
						{ExpiresAt: "", ServerID: auth.NoServerSelected},
					},
				}
			},
			resp: internal.CodeDedicatedIPServiceButNoServers,
		},
	}

	for _, test := range tests {
		// run each test using working API for servers list and using local cached servers
		// list
		servers := map[string]core.ServersAPI{
			"Remote": mockServersAPI{},
			"Local":  mockFailingServersAPI{},
		}
		for key, serversAPI := range servers {
			t.Run(test.name+" "+key, func(t *testing.T) {
				rpc := testRPC()
				rpc.serversAPI = serversAPI
				if test.setup != nil {
					test.setup(rpc)
				}
				server := &mockRPCServer{}
				err := rpc.Connect(&pb.ConnectRequest{
					ServerGroup: test.serverGroup,
					ServerTag:   test.serverTag,
				}, server)
				if test.resp == internal.CodeConnected {
					assert.NoError(t, err)
				} else if test.resp == 0 {
					assert.ErrorIs(t, internal.ErrUnhandled, err)
				} else {
					assert.Equal(t, test.resp, server.msg.Type)
				}
			})
		}
	}
}

func TestRpcReconnect(t *testing.T) {
	category.Set(t, category.Route)

	cm := newMockConfigManager()
	tokenData := cm.c.TokensData[cm.c.AutoConnectData.ID]
	tokenData.TokenExpiry = time.Now().Add(time.Hour * 1).Format(internal.ServerDateFormat)
	tokenData.ServiceExpiry = time.Now().Add(time.Hour * 1).Format(internal.ServerDateFormat)
	cm.c.TokensData[cm.c.AutoConnectData.ID] = tokenData

	rpc := testRPC()
	err := rpc.Connect(&pb.ConnectRequest{}, &mockRPCServer{})
	assert.NoError(t, err)

	rpc.netw = testnetworker.Failing{} // second connect has to fail
	err = rpc.Connect(&pb.ConnectRequest{}, &mockRPCServer{})
	assert.NoError(t, err)

	rpc.netw = &testnetworker.Mock{}
	err = rpc.Connect(&pb.ConnectRequest{}, &mockRPCServer{})
	assert.NoError(t, err)
}
