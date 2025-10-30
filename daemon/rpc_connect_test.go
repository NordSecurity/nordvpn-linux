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
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/recents"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	mockconfig "github.com/NordSecurity/nordvpn-linux/test/mock/config"
	testnetworker "github.com/NordSecurity/nordvpn-linux/test/mock/networker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRPCServer struct {
	pb.Daemon_ConnectServer
	msg *pb.Payload
}

func (m *mockRPCServer) Send(p *pb.Payload) error { m.msg = p; return nil }

// deterministicServersAPI provides deterministic server selection for connect tests
// while reusing the existing serversList() and countriesList() functions
type deterministicServersAPI struct {
	counter int64
}

func (deterministicServersAPI) Servers() (core.Servers, http.Header, error) {
	return serversList(), nil, nil
}

func (d *deterministicServersAPI) RecommendedServers(filter core.ServersFilter, _ float64, _ float64) (core.Servers, http.Header, error) {
	allServers := serversList()

	if filter.Tag.Action == core.ServerByUnknown && filter.Group == config.ServerGroup_UNDEFINED {
		return getServersByID(allServers, 1), nil, nil
	}

	if filter.Tag.Action == core.ServerByName {
		switch filter.Tag.ID {
		case 3:
			return getServersByID(allServers, 2), nil, nil
		case 7:
			return getServersByID(allServers, 7), nil, nil
		case 8:
			return getServersByID(allServers, 8), nil, nil
		}
	}

	if filter.Tag.Action == core.ServerByCountry {
		switch filter.Tag.ID {
		case 4:
			if filter.Group == config.ServerGroup_P2P {
				return getServersByID(allServers, 2), nil, nil
			}
			return getServersByID(allServers, 2), nil, nil
		case 2:
			return getServersByID(allServers, 1), nil, nil
		case 3:
			return getServersByID(allServers, 5), nil, nil
		case 5:
			return getServersByID(allServers, 10), nil, nil
		default:

			d.counter++
			if d.counter%2 == 1 {
				return getServersByID(allServers, 2), nil, nil
			}
			return getServersByID(allServers, 1), nil, nil
		}
	}

	if filter.Tag.Action == core.ServerByCity {
		if filter.Group == config.ServerGroup_P2P {
			return getServersByID(allServers, 2), nil, nil
		}
		return getServersByID(allServers, 2), nil, nil
	}

	if filter.Tag.Action == core.ServerByUnknown && filter.Group != config.ServerGroup_UNDEFINED {
		switch filter.Group {
		case config.ServerGroup_P2P:

			return getServersByID(allServers, 3), nil, nil
		case config.ServerGroup_DEDICATED_IP:

			return getServersByID(allServers, 7), nil, nil
		case config.ServerGroup_UNDEFINED,
			config.ServerGroup_DOUBLE_VPN,
			config.ServerGroup_ONION_OVER_VPN,
			config.ServerGroup_ULTRA_FAST_TV,
			config.ServerGroup_ANTI_DDOS,
			config.ServerGroup_STANDARD_VPN_SERVERS,
			config.ServerGroup_NETFLIX_USA,
			config.ServerGroup_OBFUSCATED,
			config.ServerGroup_EUROPE,
			config.ServerGroup_THE_AMERICAS,
			config.ServerGroup_ASIA_PACIFIC,
			config.ServerGroup_AFRICA_THE_MIDDLE_EAST_AND_INDIA:

			return getServersByID(allServers, 1), nil, nil
		}
	}

	return getServersByID(allServers, 1), nil, nil
}

func (deterministicServersAPI) Server(serverID int64) (*core.Server, error) {
	allServers := serversList()
	for _, server := range allServers {
		if server.ID == serverID {
			return &server, nil
		}
	}
	return nil, fmt.Errorf("server not found")
}

func (deterministicServersAPI) ServersCountries() (core.Countries, http.Header, error) {
	return countriesList(), nil, nil
}

func (deterministicServersAPI) ServersTechnologiesConfigurations(string, int64, core.ServerTechnology) ([]byte, error) {
	return nil, nil
}

// getServersByID helper function reuses existing servers list
func getServersByID(servers core.Servers, id int64) core.Servers {
	for _, server := range servers {
		if server.ID == id {
			return core.Servers{server}
		}
	}
	return core.Servers{}
}

func testRPCLocal(t *testing.T) *RPC {
	rpc := testRPC()

	fs := mockconfig.NewFilesystemMock(t)
	recentStore := recents.NewRecentConnectionsStore("/test/recents_"+t.Name()+".dat", &fs)
	rpc.recentVPNConnStore = recentStore

	rpc.serversAPI = &deterministicServersAPI{}

	rpc.dm.SetCountryData(time.Now(), countriesList(), "")

	return rpc
}

type workingLoginChecker struct {
	isVPNExpired         bool
	vpnErr               error
	isDedicatedIPExpired bool
	dedicatedIPErr       error
	dedicatedIPService   []auth.DedicatedIPService
}

func (*workingLoginChecker) IsLoggedIn() (bool, error)     { return true, nil }
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
						{ExpiresAt: "", ServerIDs: []int64{7}},
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
						{ExpiresAt: "", ServerIDs: []int64{7}},
					},
				}
			},
			resp: internal.CodeConnected,
		},
		{
			name:      "Dedicated IP with server name multiple servers in service works",
			serverTag: "lt7",
			factory: func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			},
			setup: func(rpc *RPC) {
				rpc.ac = &workingLoginChecker{
					isDedicatedIPExpired: false,
					dedicatedIPService: []auth.DedicatedIPService{
						{ExpiresAt: "", ServerIDs: []int64{7, 8}},
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
						{ExpiresAt: "", ServerIDs: []int64{7}},
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
						{ExpiresAt: "", ServerIDs: []int64{}},
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
				rpc := testRPCLocal(t)
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

func TestRpcConnect_RecentConnections(t *testing.T) {
	category.Set(t, category.Unit)

	defer testsCleanup()

	tests := []struct {
		name               string
		serverTag          string
		serverGroup        string
		expectedRecentConn *recents.Model
		shouldAddToRecent  bool
	}{
		{
			name:      "Country connection adds to recent",
			serverTag: "germany",
			expectedRecentConn: &recents.Model{
				Country:            "Germany",
				CountryCode:        "DE",
				ConnectionType:     config.ServerSelectionRule_COUNTRY,
				Group:              config.ServerGroup_UNDEFINED,
				ServerTechnologies: []core.ServerTechnology{core.OpenVPNUDP},
			},
			shouldAddToRecent: true,
		},
		{
			name:      "City connection adds to recent",
			serverTag: "germany berlin",
			expectedRecentConn: &recents.Model{
				Country:            "Germany",
				CountryCode:        "DE",
				City:               "Berlin",
				ConnectionType:     config.ServerSelectionRule_CITY,
				Group:              config.ServerGroup_UNDEFINED,
				ServerTechnologies: []core.ServerTechnology{core.OpenVPNUDP},
			},
			shouldAddToRecent: true,
		},
		{
			name:      "Specific server connection adds to recent",
			serverTag: "de3",
			expectedRecentConn: &recents.Model{
				Country:            "Germany",
				CountryCode:        "DE",
				City:               "Berlin",
				SpecificServer:     "de3",
				SpecificServerName: "Germany #3",
				ConnectionType:     config.ServerSelectionRule_SPECIFIC_SERVER,
				Group:              config.ServerGroup_UNDEFINED,
				ServerTechnologies: []core.ServerTechnology{core.OpenVPNUDP},
			},
			shouldAddToRecent: true,
		},
		{
			name:        "Group connection adds to recent",
			serverGroup: "P2P",
			expectedRecentConn: &recents.Model{
				// Group connections only store the group, no geographic data
				Group:              config.ServerGroup_P2P,
				ConnectionType:     config.ServerSelectionRule_GROUP,
				ServerTechnologies: []core.ServerTechnology{core.OpenVPNUDP},
			},
			shouldAddToRecent: true,
		},
		{
			name:              "Quick connect (recommended) does not add to recent",
			serverTag:         "",
			serverGroup:       "",
			shouldAddToRecent: false,
		},
		{
			name:        "Country with group adds to recent",
			serverTag:   "germany",
			serverGroup: "P2P",
			expectedRecentConn: &recents.Model{
				Country:            "Germany",
				CountryCode:        "DE",
				Group:              config.ServerGroup_P2P,
				ConnectionType:     config.ServerSelectionRule_COUNTRY_WITH_GROUP,
				ServerTechnologies: []core.ServerTechnology{core.OpenVPNUDP},
			},
			shouldAddToRecent: true,
		},
		{
			name:        "Country with city with group adds to recent",
			serverTag:   "germany berlin",
			serverGroup: "P2P",
			expectedRecentConn: &recents.Model{
				Country:            "Germany",
				CountryCode:        "DE",
				City:               "Berlin",
				Group:              config.ServerGroup_P2P,
				ConnectionType:     config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP,
				ServerTechnologies: []core.ServerTechnology{core.OpenVPNUDP},
			},
			shouldAddToRecent: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fs := mockconfig.NewFilesystemMock(t)
			recentStore := recents.NewRecentConnectionsStore("/test/recents_"+t.Name()+".dat", &fs)

			rpc := testRPCLocal(t)
			rpc.recentVPNConnStore = recentStore
			rpc.factory = func(config.Technology) (vpn.VPN, error) {
				return &mock.WorkingVPN{}, nil
			}

			server := &mockRPCServer{}
			err := rpc.Connect(&pb.ConnectRequest{
				ServerTag:   test.serverTag,
				ServerGroup: test.serverGroup,
			}, server)
			assert.NoError(t, err)

			assert.Equal(t, internal.CodeConnected, server.msg.Type)

			recentConns, err := recentStore.Get()
			require.NoError(t, err)

			if test.shouldAddToRecent {
				require.Len(t, recentConns, 1, "Expected one recent connection")

				recent := recentConns[0]
				assert.Equal(t, test.expectedRecentConn.ConnectionType, recent.ConnectionType)
				assert.Equal(t, test.expectedRecentConn.Group, recent.Group)

				assert.Equal(t, test.expectedRecentConn.Country, recent.Country)
				assert.Equal(t, test.expectedRecentConn.CountryCode, recent.CountryCode)
				assert.Equal(t, test.expectedRecentConn.City, recent.City)

				if test.expectedRecentConn.SpecificServer != "" {
					assert.Equal(t, test.expectedRecentConn.SpecificServer, recent.SpecificServer)
					assert.Equal(t, test.expectedRecentConn.SpecificServerName, recent.SpecificServerName)
					assert.Equal(t, test.expectedRecentConn.City, recent.City)
				}
			} else {
				assert.Empty(t, recentConns, "Expected no recent connections for recommended server")
			}
		})
	}
}

func TestRpcConnect_RecentConnectionsMultiple(t *testing.T) {
	category.Set(t, category.Unit)

	defer testsCleanup()

	rpc := testRPCLocal(t)
	rpc.factory = func(config.Technology) (vpn.VPN, error) {
		return &mock.WorkingVPN{}, nil
	}

	rpc.dm.SetCountryData(time.Now(), countriesList(), "")

	server := &mockRPCServer{}
	err := rpc.Connect(&pb.ConnectRequest{ServerTag: "germany"}, server)
	assert.NoError(t, err)

	server = &mockRPCServer{}
	err = rpc.Connect(&pb.ConnectRequest{ServerTag: "france"}, server)
	assert.NoError(t, err)

	server = &mockRPCServer{}
	err = rpc.Connect(&pb.ConnectRequest{ServerTag: "germany"}, server)
	assert.NoError(t, err)

	recentConns, err := rpc.recentVPNConnStore.Get()
	assert.NoError(t, err)
	assert.Len(t, recentConns, 2)

	assert.Equal(t, "Germany", recentConns[0].Country)
	assert.Equal(t, "DE", recentConns[0].CountryCode)

	assert.Equal(t, "France", recentConns[1].Country)
	assert.Equal(t, "FR", recentConns[1].CountryCode)
}

func TestRpcReconnect(t *testing.T) {
	category.Set(t, category.Route)

	cm := newMockConfigManager()
	tokenData := cm.c.TokensData[cm.c.AutoConnectData.ID]
	tokenData.TokenExpiry = time.Now().Add(time.Hour * 1).Format(internal.ServerDateFormat)
	tokenData.ServiceExpiry = time.Now().Add(time.Hour * 1).Format(internal.ServerDateFormat)
	cm.c.TokensData[cm.c.AutoConnectData.ID] = tokenData

	rpc := testRPCLocal(t)
	err := rpc.Connect(&pb.ConnectRequest{}, &mockRPCServer{})
	assert.NoError(t, err)

	rpc.netw = testnetworker.Failing{} // second connect has to fail
	err = rpc.Connect(&pb.ConnectRequest{}, &mockRPCServer{})
	assert.NoError(t, err)

	rpc.netw = &testnetworker.Mock{}
	err = rpc.Connect(&pb.ConnectRequest{}, &mockRPCServer{})
	assert.NoError(t, err)
}

func Test_determineServerSelectionRule(t *testing.T) {
	tests := []struct {
		name   string
		params ServerParameters
		want   config.ServerSelectionRule
	}{
		{
			name:   "All empty params returns RECOMMENDED",
			params: ServerParameters{},
			want:   config.ServerSelectionRule_RECOMMENDED,
		},
		{
			name: "Country, country-code, city is set returns CITY",
			params: ServerParameters{
				Country:     "Germany",
				City:        "Berlin",
				CountryCode: "DE",
			},
			want: config.ServerSelectionRule_CITY,
		},
		{
			name: "Country, country-code set, group undefined returns COUNTRY",
			params: ServerParameters{
				Country:     "Lithuania",
				Group:       config.ServerGroup_UNDEFINED,
				CountryCode: "LT",
			},
			want: config.ServerSelectionRule_COUNTRY,
		},
		{
			name: "Country, country code, group set returns COUNTRY_WITH_GROUP",
			params: ServerParameters{
				Country:     "Lithuania",
				Group:       config.ServerGroup_OBFUSCATED,
				CountryCode: "LT",
			},
			want: config.ServerSelectionRule_COUNTRY_WITH_GROUP,
		},
		{
			name: "ServerName set, group undefined returns SPECIFIC_SERVER",
			params: ServerParameters{
				ServerName: "lt11",
				Group:      config.ServerGroup_UNDEFINED,
			},
			want: config.ServerSelectionRule_SPECIFIC_SERVER,
		},
		{
			name: "ServerName set, group set returns SPECIFIC_SERVER_WITH_GROUP",
			params: ServerParameters{
				ServerName: "lt11",
				Group:      config.ServerGroup_OBFUSCATED,
			},
			want: config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP,
		},
		{
			name: "Group set returns GROUP",
			params: ServerParameters{
				Group: config.ServerGroup_OBFUSCATED,
			},
			want: config.ServerSelectionRule_GROUP,
		},
		{
			name: "Unknown combination returns RECOMMENDED",
			params: ServerParameters{
				Country:     "",
				City:        "",
				Group:       config.ServerGroup_UNDEFINED,
				CountryCode: "",
				ServerName:  "",
			},
			want: config.ServerSelectionRule_RECOMMENDED,
		},
		{
			name: "All fields set (should not match anything)",
			params: ServerParameters{
				Country:     "Germany",
				City:        "Berlin",
				Group:       config.ServerGroup_OBFUSCATED,
				CountryCode: "DE",
				ServerName:  "de123",
			},
			want: config.ServerSelectionRule_NONE,
		},
		{
			name: "Only ServerName set, others empty/undefined",
			params: ServerParameters{
				ServerName: "us123",
				Group:      config.ServerGroup_UNDEFINED,
			},
			want: config.ServerSelectionRule_SPECIFIC_SERVER,
		},
		{
			name: "Only Group set, others empty/undefined",
			params: ServerParameters{
				Group: config.ServerGroup_DOUBLE_VPN,
			},
			want: config.ServerSelectionRule_GROUP,
		},
		{
			name: "Country and ServerName set, group undefined",
			params: ServerParameters{
				Country:    "France",
				ServerName: "fr123",
				Group:      config.ServerGroup_UNDEFINED,
			},
			want: config.ServerSelectionRule_NONE,
		},
		{
			name: "Country, City, ServerName, group undefined",
			params: ServerParameters{
				Country:    "France",
				City:       "Paris",
				ServerName: "fr123",
				Group:      config.ServerGroup_UNDEFINED,
			},
			want: config.ServerSelectionRule_NONE,
		},
		{
			name: "Country, City, ServerName, group set",
			params: ServerParameters{
				Country:    "France",
				City:       "Paris",
				ServerName: "fr123",
				Group:      config.ServerGroup_OBFUSCATED,
			},
			want: config.ServerSelectionRule_NONE,
		},
		{
			name: "Country set, group set to UNDEFINED, ServerName set",
			params: ServerParameters{
				Country:    "Italy",
				Group:      config.ServerGroup_UNDEFINED,
				ServerName: "it123",
			},
			want: config.ServerSelectionRule_NONE,
		},
		{
			name: "Country set, group set, ServerName set",
			params: ServerParameters{
				Country:    "Italy",
				Group:      config.ServerGroup_DOUBLE_VPN,
				ServerName: "it123",
			},
			want: config.ServerSelectionRule_NONE,
		},
		{
			name: "ServerName set, group set to undefined, City set",
			params: ServerParameters{
				ServerName: "es123",
				Group:      config.ServerGroup_UNDEFINED,
				City:       "Madrid",
			},
			want: config.ServerSelectionRule_NONE,
		},
		{
			name: "ServerName set, group set, City set",
			params: ServerParameters{
				ServerName: "es123",
				Group:      config.ServerGroup_DOUBLE_VPN,
				City:       "Madrid",
			},
			want: config.ServerSelectionRule_NONE,
		},
		{
			name: "Group is UNDEFINED, all other fields empty",
			params: ServerParameters{
				Group: config.ServerGroup_UNDEFINED,
			},
			want: config.ServerSelectionRule_RECOMMENDED,
		},
		{
			name: "Edge: Group is invalid (not in enum), should fallback to invalid/empty",
			params: ServerParameters{
				Group: config.ServerGroup(9999),
			},
			want: config.ServerSelectionRule_NONE,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := determineServerSelectionRule(tt.params)
			if got != tt.want {
				t.Errorf("determineServerSelectionRule() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_determineServerGroup(t *testing.T) {
	tests := []struct {
		name   string
		server core.Server
		params ServerParameters
		want   string
	}{
		{
			name: "Group is UNDEFINED returns first group title",
			server: core.Server{Groups: []core.Group{
				{ID: config.ServerGroup_STANDARD_VPN_SERVERS, Title: "Standard VPN servers"},
				{ID: config.ServerGroup_DOUBLE_VPN, Title: "Double VPN"},
			}},
			params: ServerParameters{},
			want:   "Standard VPN servers",
		},
		{
			name: "Group is DOUBLE_VPN returns matching group title",
			server: core.Server{Groups: []core.Group{
				{ID: config.ServerGroup_STANDARD_VPN_SERVERS, Title: "Standard VPN servers"},
				{ID: config.ServerGroup_DOUBLE_VPN, Title: "Double VPN"},
			}},
			params: ServerParameters{Group: config.ServerGroup_DOUBLE_VPN},
			want:   "Double VPN",
		},
		{
			name: "Group is OBFUSCATED returns matching group title",
			server: core.Server{Groups: []core.Group{
				{ID: config.ServerGroup_STANDARD_VPN_SERVERS, Title: "Standard VPN servers"},
				{ID: config.ServerGroup_OBFUSCATED, Title: "Obfuscated"},
			}},
			params: ServerParameters{Group: config.ServerGroup_OBFUSCATED},
			want:   "Obfuscated",
		},
		{
			name: "Group is DEDICATED_IP returns matching group title",
			server: core.Server{Groups: []core.Group{
				{ID: config.ServerGroup_DEDICATED_IP, Title: "Dedicated IP"},
				{ID: config.ServerGroup_STANDARD_VPN_SERVERS, Title: "Standard VPN servers"},
			}},
			params: ServerParameters{Group: config.ServerGroup_DEDICATED_IP},
			want:   "Dedicated IP",
		},
		{
			name: "Group is NETFLIX_USA returns matching group title",
			server: core.Server{Groups: []core.Group{
				{ID: config.ServerGroup_NETFLIX_USA, Title: "Netflix USA"},
				{ID: config.ServerGroup_STANDARD_VPN_SERVERS, Title: "Standard VPN servers"},
			}},
			params: ServerParameters{Group: config.ServerGroup_NETFLIX_USA},
			want:   "Netflix USA",
		},
		{
			name: "Group is P2P returns matching group title",
			server: core.Server{Groups: []core.Group{
				{ID: config.ServerGroup_P2P, Title: "P2P"},
				{ID: config.ServerGroup_STANDARD_VPN_SERVERS, Title: "Standard VPN servers"},
			}},
			params: ServerParameters{Group: config.ServerGroup_P2P},
			want:   "P2P",
		},
		{
			name: "Group is ULTRA_FAST_TV returns matching group title",
			server: core.Server{Groups: []core.Group{
				{ID: config.ServerGroup_ULTRA_FAST_TV, Title: "Ultra Fast TV"},
				{ID: config.ServerGroup_STANDARD_VPN_SERVERS, Title: "Standard VPN servers"},
			}},
			params: ServerParameters{Group: config.ServerGroup_ULTRA_FAST_TV},
			want:   "Ultra Fast TV",
		},
		{
			name: "Group is ANTI_DDOS returns matching group title",
			server: core.Server{Groups: []core.Group{
				{ID: config.ServerGroup_ANTI_DDOS, Title: "Anti DDoS"},
				{ID: config.ServerGroup_STANDARD_VPN_SERVERS, Title: "Standard VPN servers"},
			}},
			params: ServerParameters{Group: config.ServerGroup_ANTI_DDOS},
			want:   "Anti DDoS",
		},
		{
			name: "Group is EUROPE returns matching group title",
			server: core.Server{Groups: []core.Group{
				{ID: config.ServerGroup_EUROPE, Title: "Europe"},
				{ID: config.ServerGroup_STANDARD_VPN_SERVERS, Title: "Standard VPN servers"},
			}},
			params: ServerParameters{Group: config.ServerGroup_EUROPE},
			want:   "Europe",
		},
		{
			name: "Group is THE_AMERICAS returns matching group title",
			server: core.Server{Groups: []core.Group{
				{ID: config.ServerGroup_THE_AMERICAS, Title: "The Americas"},
				{ID: config.ServerGroup_STANDARD_VPN_SERVERS, Title: "Standard VPN servers"},
			}},
			params: ServerParameters{Group: config.ServerGroup_THE_AMERICAS},
			want:   "The Americas",
		},
		{
			name: "Group is ASIA_PACIFIC returns matching group title",
			server: core.Server{Groups: []core.Group{
				{ID: config.ServerGroup_ASIA_PACIFIC, Title: "Asia Pacific"},
				{ID: config.ServerGroup_STANDARD_VPN_SERVERS, Title: "Standard VPN servers"},
			}},
			params: ServerParameters{Group: config.ServerGroup_ASIA_PACIFIC},
			want:   "Asia Pacific",
		},
		{
			name: "Group is AFRICA_THE_MIDDLE_EAST_AND_INDIA returns matching group title",
			server: core.Server{Groups: []core.Group{
				{ID: config.ServerGroup_AFRICA_THE_MIDDLE_EAST_AND_INDIA, Title: "Africa, the Middle East and India"},
				{ID: config.ServerGroup_STANDARD_VPN_SERVERS, Title: "Standard VPN servers"},
			}},
			params: ServerParameters{Group: config.ServerGroup_AFRICA_THE_MIDDLE_EAST_AND_INDIA},
			want:   "Africa, the Middle East and India",
		},
		{
			name:   "Server has no groups returns empty string",
			server: core.Server{Groups: []core.Group{}},
			params: ServerParameters{Group: config.ServerGroup_DOUBLE_VPN},
			want:   "",
		},
		{
			name: "Server has only one group, params group matches",
			server: core.Server{Groups: []core.Group{
				{ID: config.ServerGroup_OBFUSCATED, Title: "Obfuscated"},
			}},
			params: ServerParameters{Group: config.ServerGroup_OBFUSCATED},
			want:   "Obfuscated",
		},
		{
			name: "Group is not set (zero value), server has multiple groups",
			server: core.Server{Groups: []core.Group{
				{ID: config.ServerGroup_P2P, Title: "P2P"},
				{ID: config.ServerGroup_STANDARD_VPN_SERVERS, Title: "Standard VPN servers"},
			}},
			params: ServerParameters{},
			want:   "Standard VPN servers",
		},
		{
			name:   "Group is not set (zero value), server has no groups",
			server: core.Server{Groups: []core.Group{}},
			params: ServerParameters{},
			want:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := determineTargetServerGroup(&tt.server, tt.params); got != tt.want {
				t.Errorf("determineServerGroup() = %v, want %v", got, tt.want)
			}
		})
	}
}
