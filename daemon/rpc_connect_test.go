package daemon

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/recents"
	"github.com/NordSecurity/nordvpn-linux/daemon/serverpicker"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	devicekey "github.com/NordSecurity/nordvpn-linux/device_key"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	core_test "github.com/NordSecurity/nordvpn-linux/test/mock/core"
	testdevicekey "github.com/NordSecurity/nordvpn-linux/test/mock/devicekey"
	"github.com/NordSecurity/nordvpn-linux/test/mock/fs"
	testnetworker "github.com/NordSecurity/nordvpn-linux/test/mock/networker"
	"github.com/google/uuid"
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
	return core_test.ServersList(), nil, nil
}

func (d *deterministicServersAPI) RecommendedServers(filter core.ServersFilter, _ float64, _ float64) (core.Servers, http.Header, error) {
	allServers := core_test.ServersList()

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
			config.ServerGroup_OBFUSCATED:

			return getServersByID(allServers, 1), nil, nil
		case config.ServerGroup_DEDICATED_SERVER:
			panic("dedicated servers will never be recommended")
		}
	}

	return getServersByID(allServers, 1), nil, nil
}

func (deterministicServersAPI) Server(serverID int64) (*core.Server, error) {
	allServers := core_test.ServersList()
	for _, server := range allServers {
		if server.ID == serverID {
			return &server, nil
		}
	}
	return nil, fmt.Errorf("server not found")
}

func (deterministicServersAPI) ServersCountries() (core.Countries, http.Header, error) {
	return core_test.CountriesList(), nil, nil
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

	fs := fs.NewSystemFileHandleMock(t)
	recentStore := recents.NewRecentConnectionsStore("/test/recents_"+t.Name()+".dat", &fs, nil)
	rpc.recentVPNConnStore = recentStore

	rpc.serversAPI = &deterministicServersAPI{}

	rpc.dm.SetCountryData(time.Now(), core_test.CountriesList(), "")

	return rpc
}

type workingLoginChecker struct {
	isVPNExpired              bool
	vpnErr                    error
	isDedicatedIPExpired      bool
	dedicatedIPErr            error
	dedicatedIPService        []auth.DedicatedIPService
	isDedicatedServersExpired bool
	dedicatedServerErr        error
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
func (c *workingLoginChecker) GetDedicatedServerService() (auth.DedicatedServerService, error) {
	return auth.DedicatedServerService{Active: !c.isDedicatedServersExpired}, c.dedicatedServerErr
}

func TestRPCConnect(t *testing.T) {
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
			"Remote": core_test.NewMockServersAPI(),
			"Local":  core_test.NewMockFailingServersAPI(errors.New("500")),
		}
		for key, serversAPI := range servers {
			t.Run(test.name+" "+key, func(t *testing.T) {
				rpc := testRPCLocal(t)
				rpc.serversAPI = serversAPI
				if test.setup != nil {
					test.setup(rpc)
				}
				firstOpenEventCouter := 0
				firstOpenListener := func(any) error {
					firstOpenEventCouter++
					return nil
				}
				rpc.events.Service.FirstTimeOpened.Subscribe(firstOpenListener)
				server := &mockRPCServer{}

				mockPauseManager := &mock.PauseSchedulerMock{
					ConnectionScheduled: true,
				}

				rpc.pauseManager = mockPauseManager

				err := rpc.Connect(&pb.ConnectRequest{
					ServerGroup: test.serverGroup,
					ServerTag:   test.serverTag,
				}, server)

				assert.False(t, mockPauseManager.ConnectionScheduled,
					"Paused connection was not cancelled after another connection attempt.")

				switch test.resp {
				case internal.CodeConnected:
					assert.NoError(t, err)
					assert.Equal(t, firstOpenEventCouter, 1)
				case 0:
					assert.ErrorIs(t, internal.ErrUnhandled, err)
					assert.Equal(t, firstOpenEventCouter, 0)
				default:
					assert.Equal(t, test.resp, server.msg.Type)
					assert.Equal(t, firstOpenEventCouter, 0)
				}
			})
		}
	}
}

func TestRPCConnect_RecentConnections(t *testing.T) {
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
				City:               "Berlin",
				ConnectionType:     config.ServerSelectionRule_CITY,
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
				City:               "Berlin",
				Group:              config.ServerGroup_P2P,
				ConnectionType:     config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP,
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
			rpc := testRPCLocal(t)
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

			// Manually store the pending connection (normally happens on disconnect)
			storePendingRecentConnection(rpc.recentVPNConnStore)

			recentConns, err := rpc.recentVPNConnStore.Get()
			require.NoError(t, err)

			if test.shouldAddToRecent {
				require.Len(t, recentConns, 1, "Expected one recent connection")

				recent := recentConns[0]
				assert.Equal(t, test.expectedRecentConn.ConnectionType, recent.ConnectionType)
				assert.Equal(t, test.expectedRecentConn.Group, recent.Group)

				assert.Equal(t, test.expectedRecentConn.Country, recent.Country)
				assert.Equal(t, test.expectedRecentConn.CountryCode, recent.CountryCode)
				// Only check City if it's expected to be set
				if test.expectedRecentConn.ConnectionType == config.ServerSelectionRule_CITY ||
					test.expectedRecentConn.ConnectionType == config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP {
					assert.Equal(t, test.expectedRecentConn.City, recent.City)
				}

				if test.expectedRecentConn.SpecificServer != "" {
					assert.Equal(t, test.expectedRecentConn.SpecificServer, recent.SpecificServer)
					assert.Equal(t, test.expectedRecentConn.SpecificServerName, recent.SpecificServerName)
				}
			} else {
				assert.Empty(t, recentConns, "Expected no recent connections for recommended server")
			}
		})
	}
}

func TestRPCConnect_RecentConnectionsOnFailure_PreviousConnectionStored(t *testing.T) {
	category.Set(t, category.Unit)

	defer testsCleanup()

	rpc := testRPCLocal(t)

	rpc.factory = func(config.Technology) (vpn.VPN, error) {
		return &mock.WorkingVPN{}, nil
	}

	server := &mockRPCServer{}
	err := rpc.Connect(&pb.ConnectRequest{ServerTag: "germany"}, server)
	assert.NoError(t, err)
	assert.Equal(t, internal.CodeConnected, server.msg.Type)

	// Now try to switch to France but it fails (making networker fail)
	rpc.netw = testnetworker.Failing{}
	rpc.factory = func(config.Technology) (vpn.VPN, error) {
		return &mock.FailingVPN{}, nil
	}

	server = &mockRPCServer{}
	err = rpc.Connect(&pb.ConnectRequest{ServerTag: "france"}, server)
	assert.NoError(t, err)
	assert.Equal(t, internal.CodeFailure, server.msg.Type)

	// Should have stored the pending connection (Germany)
	// even though the new connection (France) failed
	recentConns, err := rpc.recentVPNConnStore.Get()
	assert.NoError(t, err)

	// Should have the previous successful connection (Germany)
	assert.Len(t, recentConns, 1, "Expected one recent connection from previous successful connect")
	assert.Equal(t, "Germany", recentConns[0].Country)
	assert.Equal(t, "DE", recentConns[0].CountryCode)
}

func TestRPCConnect_RecentConnectionsOnFailure_MultipleConnectionsPreserved(t *testing.T) {
	category.Set(t, category.Unit)

	defer testsCleanup()

	rpc := testRPCLocal(t)
	rpc.factory = func(config.Technology) (vpn.VPN, error) {
		return &mock.WorkingVPN{}, nil
	}

	server := &mockRPCServer{}
	err := rpc.Connect(&pb.ConnectRequest{ServerTag: "germany"}, server)
	assert.NoError(t, err)

	server = &mockRPCServer{}
	err = rpc.Connect(&pb.ConnectRequest{ServerTag: "france"}, server)
	assert.NoError(t, err)

	// Now try to connect to a P2P server but it fails (making networker fail)
	rpc.netw = testnetworker.Failing{}
	rpc.factory = func(config.Technology) (vpn.VPN, error) {
		return &mock.FailingVPN{}, nil
	}

	server = &mockRPCServer{}
	err = rpc.Connect(&pb.ConnectRequest{ServerGroup: "P2P"}, server)
	assert.NoError(t, err)
	assert.Equal(t, internal.CodeFailure, server.msg.Type)

	// Should have both previous successful connections
	recentConns, err := rpc.recentVPNConnStore.Get()
	assert.NoError(t, err)
	assert.Len(t, recentConns, 2, "Expected two recent connections from previous successful connects")

	// Most recent successful connection should be first (France)
	assert.Equal(t, "France", recentConns[0].Country)
	assert.Equal(t, "Germany", recentConns[1].Country)
}

func TestRPCConnect_RecentConnectionsMultiple(t *testing.T) {
	category.Set(t, category.Unit)

	defer testsCleanup()

	rpc := testRPCLocal(t)
	rpc.factory = func(config.Technology) (vpn.VPN, error) {
		return &mock.WorkingVPN{}, nil
	}

	rpc.dm.SetCountryData(time.Now(), core_test.CountriesList(), "")

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

	// Recents are stored with most recent first
	// Sequence: Germany → France → Germany (reconnect moves Germany to front)
	assert.Equal(t, "France", recentConns[0].Country)
	assert.Equal(t, "FR", recentConns[0].CountryCode)

	assert.Equal(t, "Germany", recentConns[1].Country)
	assert.Equal(t, "DE", recentConns[1].CountryCode)
}

func TestRPCReconnect(t *testing.T) {
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
		params serverpicker.ServerParameters
		want   config.ServerSelectionRule
	}{
		{
			name:   "All empty params returns RECOMMENDED",
			params: serverpicker.ServerParameters{},
			want:   config.ServerSelectionRule_RECOMMENDED,
		},
		{
			name: "Country, country-code, city is set returns CITY",
			params: serverpicker.ServerParameters{
				Country:     "Germany",
				City:        "Berlin",
				CountryCode: "DE",
			},
			want: config.ServerSelectionRule_CITY,
		},
		{
			name: "Country, country-code set, group undefined returns COUNTRY",
			params: serverpicker.ServerParameters{
				Country:     "Lithuania",
				Group:       config.ServerGroup_UNDEFINED,
				CountryCode: "LT",
			},
			want: config.ServerSelectionRule_COUNTRY,
		},
		{
			name: "Country, country code, group set returns COUNTRY_WITH_GROUP",
			params: serverpicker.ServerParameters{
				Country:     "Lithuania",
				Group:       config.ServerGroup_OBFUSCATED,
				CountryCode: "LT",
			},
			want: config.ServerSelectionRule_COUNTRY_WITH_GROUP,
		},
		{
			name: "ServerName set, group undefined returns SPECIFIC_SERVER",
			params: serverpicker.ServerParameters{
				ServerName: "lt11",
				Group:      config.ServerGroup_UNDEFINED,
			},
			want: config.ServerSelectionRule_SPECIFIC_SERVER,
		},
		{
			name: "ServerName set, group set returns SPECIFIC_SERVER_WITH_GROUP",
			params: serverpicker.ServerParameters{
				ServerName: "lt11",
				Group:      config.ServerGroup_OBFUSCATED,
			},
			want: config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP,
		},
		{
			name: "Group set returns GROUP",
			params: serverpicker.ServerParameters{
				Group: config.ServerGroup_OBFUSCATED,
			},
			want: config.ServerSelectionRule_GROUP,
		},
		{
			name: "Unknown combination returns RECOMMENDED",
			params: serverpicker.ServerParameters{
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
			params: serverpicker.ServerParameters{
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
			params: serverpicker.ServerParameters{
				ServerName: "us123",
				Group:      config.ServerGroup_UNDEFINED,
			},
			want: config.ServerSelectionRule_SPECIFIC_SERVER,
		},
		{
			name: "Only Group set, others empty/undefined",
			params: serverpicker.ServerParameters{
				Group: config.ServerGroup_DOUBLE_VPN,
			},
			want: config.ServerSelectionRule_GROUP,
		},
		{
			name: "Country and ServerName set, group undefined",
			params: serverpicker.ServerParameters{
				Country:    "France",
				ServerName: "fr123",
				Group:      config.ServerGroup_UNDEFINED,
			},
			want: config.ServerSelectionRule_NONE,
		},
		{
			name: "Country, City, ServerName, group undefined",
			params: serverpicker.ServerParameters{
				Country:    "France",
				City:       "Paris",
				ServerName: "fr123",
				Group:      config.ServerGroup_UNDEFINED,
			},
			want: config.ServerSelectionRule_NONE,
		},
		{
			name: "Country, City, ServerName, group set",
			params: serverpicker.ServerParameters{
				Country:    "France",
				City:       "Paris",
				ServerName: "fr123",
				Group:      config.ServerGroup_OBFUSCATED,
			},
			want: config.ServerSelectionRule_NONE,
		},
		{
			name: "Country set, group set to UNDEFINED, ServerName set",
			params: serverpicker.ServerParameters{
				Country:    "Italy",
				Group:      config.ServerGroup_UNDEFINED,
				ServerName: "it123",
			},
			want: config.ServerSelectionRule_NONE,
		},
		{
			name: "Country set, group set, ServerName set",
			params: serverpicker.ServerParameters{
				Country:    "Italy",
				Group:      config.ServerGroup_DOUBLE_VPN,
				ServerName: "it123",
			},
			want: config.ServerSelectionRule_NONE,
		},
		{
			name: "ServerName set, group set to undefined, City set",
			params: serverpicker.ServerParameters{
				ServerName: "es123",
				Group:      config.ServerGroup_UNDEFINED,
				City:       "Madrid",
			},
			want: config.ServerSelectionRule_NONE,
		},
		{
			name: "ServerName set, group set, City set",
			params: serverpicker.ServerParameters{
				ServerName: "es123",
				Group:      config.ServerGroup_DOUBLE_VPN,
				City:       "Madrid",
			},
			want: config.ServerSelectionRule_NONE,
		},
		{
			name: "Group is UNDEFINED, all other fields empty",
			params: serverpicker.ServerParameters{
				Group: config.ServerGroup_UNDEFINED,
			},
			want: config.ServerSelectionRule_RECOMMENDED,
		},
		{
			name: "Edge: Group is invalid (not in enum), should fallback to invalid/empty",
			params: serverpicker.ServerParameters{
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
		params serverpicker.ServerParameters
		want   string
	}{
		{
			name: "Group is UNDEFINED returns first group title",
			server: core.Server{Groups: []core.Group{
				{ID: config.ServerGroup_STANDARD_VPN_SERVERS, Title: "Standard VPN servers"},
				{ID: config.ServerGroup_DOUBLE_VPN, Title: "Double VPN"},
			}},
			params: serverpicker.ServerParameters{},
			want:   "Standard VPN servers",
		},
		{
			name: "Group is DOUBLE_VPN returns matching group title",
			server: core.Server{Groups: []core.Group{
				{ID: config.ServerGroup_STANDARD_VPN_SERVERS, Title: "Standard VPN servers"},
				{ID: config.ServerGroup_DOUBLE_VPN, Title: "Double VPN"},
			}},
			params: serverpicker.ServerParameters{Group: config.ServerGroup_DOUBLE_VPN},
			want:   "Double VPN",
		},
		{
			name: "Group is OBFUSCATED returns matching group title",
			server: core.Server{Groups: []core.Group{
				{ID: config.ServerGroup_STANDARD_VPN_SERVERS, Title: "Standard VPN servers"},
				{ID: config.ServerGroup_OBFUSCATED, Title: "Obfuscated"},
			}},
			params: serverpicker.ServerParameters{Group: config.ServerGroup_OBFUSCATED},
			want:   "Obfuscated",
		},
		{
			name: "Group is DEDICATED_IP returns matching group title",
			server: core.Server{Groups: []core.Group{
				{ID: config.ServerGroup_DEDICATED_IP, Title: "Dedicated IP"},
				{ID: config.ServerGroup_STANDARD_VPN_SERVERS, Title: "Standard VPN servers"},
			}},
			params: serverpicker.ServerParameters{Group: config.ServerGroup_DEDICATED_IP},
			want:   "Dedicated IP",
		},
		{
			name: "Group is NETFLIX_USA returns matching group title",
			server: core.Server{Groups: []core.Group{
				{ID: config.ServerGroup_NETFLIX_USA, Title: "Netflix USA"},
				{ID: config.ServerGroup_STANDARD_VPN_SERVERS, Title: "Standard VPN servers"},
			}},
			params: serverpicker.ServerParameters{Group: config.ServerGroup_NETFLIX_USA},
			want:   "Netflix USA",
		},
		{
			name: "Group is P2P returns matching group title",
			server: core.Server{Groups: []core.Group{
				{ID: config.ServerGroup_P2P, Title: "P2P"},
				{ID: config.ServerGroup_STANDARD_VPN_SERVERS, Title: "Standard VPN servers"},
			}},
			params: serverpicker.ServerParameters{Group: config.ServerGroup_P2P},
			want:   "P2P",
		},
		{
			name: "Group is ULTRA_FAST_TV returns matching group title",
			server: core.Server{Groups: []core.Group{
				{ID: config.ServerGroup_ULTRA_FAST_TV, Title: "Ultra Fast TV"},
				{ID: config.ServerGroup_STANDARD_VPN_SERVERS, Title: "Standard VPN servers"},
			}},
			params: serverpicker.ServerParameters{Group: config.ServerGroup_ULTRA_FAST_TV},
			want:   "Ultra Fast TV",
		},
		{
			name: "Group is ANTI_DDOS returns matching group title",
			server: core.Server{Groups: []core.Group{
				{ID: config.ServerGroup_ANTI_DDOS, Title: "Anti DDoS"},
				{ID: config.ServerGroup_STANDARD_VPN_SERVERS, Title: "Standard VPN servers"},
			}},
			params: serverpicker.ServerParameters{Group: config.ServerGroup_ANTI_DDOS},
			want:   "Anti DDoS",
		},
		{
			name:   "Server has no groups returns empty string",
			server: core.Server{Groups: []core.Group{}},
			params: serverpicker.ServerParameters{Group: config.ServerGroup_DOUBLE_VPN},
			want:   "",
		},
		{
			name: "Server has only one group, params group matches",
			server: core.Server{Groups: []core.Group{
				{ID: config.ServerGroup_OBFUSCATED, Title: "Obfuscated"},
			}},
			params: serverpicker.ServerParameters{Group: config.ServerGroup_OBFUSCATED},
			want:   "Obfuscated",
		},
		{
			name: "Group is not set (zero value), server has multiple groups",
			server: core.Server{Groups: []core.Group{
				{ID: config.ServerGroup_P2P, Title: "P2P"},
				{ID: config.ServerGroup_STANDARD_VPN_SERVERS, Title: "Standard VPN servers"},
			}},
			params: serverpicker.ServerParameters{},
			want:   "Standard VPN servers",
		},
		{
			name:   "Group is not set (zero value), server has no groups",
			server: core.Server{Groups: []core.Group{}},
			params: serverpicker.ServerParameters{},
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

func TestConnect_DedicatedServers(t *testing.T) {
	category.Set(t, category.Unit)

	serverUUID := uuid.MustParse("af0bc2b1-785a-4455-bfe0-5397c39c4f4e")
	serverAddress := "1.2.3.4"
	serverPort := 55555
	dedicatedServer := core.DedicatedServer{
		UUID:   serverUUID.String(),
		Name:   "server 1",
		Status: core.DedicatedServerStatusRunning,
		IP:     serverAddress,
	}
	dedicatedServers := core.DedicatedServers{dedicatedServer}

	dedicatedServerNotReady := dedicatedServer
	dedicatedServerNotReady.Status = "not running"

	dedicatedServerStopped := dedicatedServer
	dedicatedServerStopped.Status = core.DedicatedServerStatusStopped

	dedicatedServerStopping := dedicatedServer
	dedicatedServerStopping.Status = core.DedicatedServerStatusStopping

	dedicatedServerUppercaseStopped := dedicatedServer
	dedicatedServerUppercaseStopped.Status = "STOPPED"

	serverPublicKey := "public_key"
	deviceUUID := uuid.MustParse("a25b9415-c77b-4a18-b8ff-ef06042aa450")
	connectResponse := core.DedicatedServerConnectResponse{
		ServerEndpoint:  serverAddress + ":" + strconv.Itoa(serverPort),
		ServerPublicKey: serverPublicKey,
	}

	devicePublicKey := "device public key"

	connectRequest := core.DedicatedServerConnectRequest{
		DeviceUUID:      deviceUUID.String(),
		DevicePublicKey: devicePublicKey,
	}

	tests := []struct {
		name                       string
		isDedicatedServersExpired  bool
		postQuantum                bool
		dedicatedServersResponse   core.DedicatedServers
		connectResponse            core.DedicatedServerConnectResponse
		technology                 config.Technology
		serviceCheckErr            error
		dedicatedServersFetchErr   error
		connectErr                 error
		expectedErr                error
		expectedStatus             int64
		expectedConnectRequestUUID string
		expectedConnectRequest     core.DedicatedServerConnectRequest
	}{
		{
			name:                       "success",
			isDedicatedServersExpired:  false,
			dedicatedServersResponse:   dedicatedServers,
			connectResponse:            connectResponse,
			technology:                 config.Technology_NORDLYNX,
			expectedStatus:             internal.CodeConnected,
			expectedConnectRequestUUID: serverUUID.String(),
			expectedConnectRequest:     connectRequest,
		},
		{
			name:                      "dedicated servers service has expired",
			isDedicatedServersExpired: true,
			technology:                config.Technology_NORDLYNX,
			expectedStatus:            internal.CodeDedicatedServersRenewError,
		},
		{
			name:                     "empty dedicated servers list",
			dedicatedServersResponse: core.DedicatedServers{},
			technology:               config.Technology_NORDLYNX,
			expectedStatus:           internal.CodeDedicatedServersServiceButNoServers,
		},
		{
			name:                     "dedicated server not ready",
			dedicatedServersResponse: core.DedicatedServers{dedicatedServerNotReady},
			technology:               config.Technology_NORDLYNX,
			expectedStatus:           internal.CodeDedicatedServersNotReady,
		},
		{
			name:                     "technology is not nordlynx",
			dedicatedServersResponse: dedicatedServers,
			connectResponse:          connectResponse,
			technology:               config.Technology_OPENVPN,
			expectedStatus:           internal.CodeDedicatedServersNoNordlynx,
		},
		{
			name:            "dedicated server service check fails",
			technology:      config.Technology_NORDLYNX,
			serviceCheckErr: errors.New("error"),
			expectedErr:     internal.ErrUnhandled,
		},
		{
			name:                     "dedicated servers fetch fails",
			technology:               config.Technology_NORDLYNX,
			dedicatedServersFetchErr: errors.New("error"),
			expectedErr:              internal.ErrUnhandled,
		},
		{
			name:                       "connect fails",
			isDedicatedServersExpired:  false,
			dedicatedServersResponse:   dedicatedServers,
			technology:                 config.Technology_NORDLYNX,
			connectErr:                 errors.New("error"),
			expectedErr:                internal.ErrUnhandled,
			expectedConnectRequestUUID: serverUUID.String(),
			expectedConnectRequest: core.DedicatedServerConnectRequest{
				DeviceUUID:      deviceUUID.String(),
				DevicePublicKey: devicePublicKey,
			},
		},
		{
			name:                     "dedicated server is stopped",
			dedicatedServersResponse: core.DedicatedServers{dedicatedServerStopped},
			technology:               config.Technology_NORDLYNX,
			expectedStatus:           internal.CodeDedicatedServersCanNotConnect,
		},
		{
			name:                     "dedicated server is stopping",
			dedicatedServersResponse: core.DedicatedServers{dedicatedServerStopping},
			technology:               config.Technology_NORDLYNX,
			expectedStatus:           internal.CodeDedicatedServersCanNotConnect,
		},
		{
			name:                     "dedicated server status is case-insensitive",
			dedicatedServersResponse: core.DedicatedServers{dedicatedServerUppercaseStopped},
			technology:               config.Technology_NORDLYNX,
			expectedStatus:           internal.CodeDedicatedServersCanNotConnect,
		},
		{
			name:                       "API returns 400 - session limit reached",
			technology:                 config.Technology_NORDLYNX,
			dedicatedServersResponse:   dedicatedServers,
			expectedConnectRequestUUID: serverUUID.String(),
			expectedConnectRequest:     connectRequest,
			connectErr:                 core.ErrDedicatedServersSessionMaxLimitReached,
			expectedStatus:             internal.CodeDedicatedServersSessionMaxLimitReached,
		},
		{
			name:                       "API returns 400 - device not found",
			technology:                 config.Technology_NORDLYNX,
			dedicatedServersResponse:   dedicatedServers,
			expectedConnectRequestUUID: serverUUID.String(),
			expectedConnectRequest:     connectRequest,
			connectErr:                 core.ErrDedicatedServersDeviceNotFound,
			expectedStatus:             internal.CodeDedicatedServersCanNotConnect,
		},
		{
			name:                       "API returns 400 - device not registered",
			technology:                 config.Technology_NORDLYNX,
			dedicatedServersResponse:   dedicatedServers,
			expectedConnectRequestUUID: serverUUID.String(),
			expectedConnectRequest:     connectRequest,
			connectErr:                 core.ErrDedicatedServersDeviceNotRegistered,
			expectedStatus:             internal.CodeDedicatedServersCanNotConnect,
		},
		{
			name:                       "API returns 400 - public key mismatch",
			technology:                 config.Technology_NORDLYNX,
			dedicatedServersResponse:   dedicatedServers,
			expectedConnectRequestUUID: serverUUID.String(),
			expectedConnectRequest:     connectRequest,
			connectErr:                 core.ErrDedicatedServersPublicKeyMismatch,
			expectedStatus:             internal.CodeDedicatedServersCanNotConnect,
		},
		{
			name:                       "API returns 400 - server offline",
			technology:                 config.Technology_NORDLYNX,
			dedicatedServersResponse:   dedicatedServers,
			expectedConnectRequestUUID: serverUUID.String(),
			expectedConnectRequest:     connectRequest,
			connectErr:                 core.ErrDedicatedServersServerOffline,
			expectedStatus:             internal.CodeDedicatedServersCanNotConnect,
		},
		{
			name:                       "API returns 400 - server not found",
			technology:                 config.Technology_NORDLYNX,
			dedicatedServersResponse:   dedicatedServers,
			expectedConnectRequestUUID: serverUUID.String(),
			expectedConnectRequest:     connectRequest,
			connectErr:                 core.ErrDedicatedServersServerNotFound,
			expectedStatus:             internal.CodeDedicatedServersCanNotConnect,
		},
		{
			name:                       "API returns 400 - invalid form data",
			technology:                 config.Technology_NORDLYNX,
			dedicatedServersResponse:   dedicatedServers,
			expectedConnectRequestUUID: serverUUID.String(),
			expectedConnectRequest:     connectRequest,
			connectErr:                 core.ErrDedicatedServersInvalidFormData,
			expectedStatus:             internal.CodeDedicatedServersCanNotConnect,
		},
		{
			name:                      "post quantum is on",
			isDedicatedServersExpired: false,
			postQuantum:               true,
			technology:                config.Technology_NORDLYNX,
			expectedStatus:            internal.CodeDedicatedServersPq,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rpc := testRPCLocal(t)
			mockDedicatedServersAPI := core_test.DedicatedServersAPIMock{
				DedicatedServersResponse: test.dedicatedServersResponse,
				ConnectResponse:          test.connectResponse,
				DedicatedServerErr:       test.dedicatedServersFetchErr,
				ConnectErr:               test.connectErr,
			}
			rpc.dedicatedServersAPI = &mockDedicatedServersAPI
			rpc.ac = &workingLoginChecker{
				isDedicatedServersExpired: test.isDedicatedServersExpired,
				dedicatedServerErr:        test.serviceCheckErr,
			}
			rpc.dedicatedServerKeyManager = &testdevicekey.MockDeviceKeyManager{
				DedicatedServerRegistrationData: &devicekey.DedicatedServersConnectionData{
					DevicePublicKey: devicePublicKey,
					DeviceUUID:      deviceUUID,
				},
			}

			configManagerMock := rpc.cm.(*mockConfigManager)
			configManagerMock.c.Technology = test.technology
			configManagerMock.c.AutoConnectData.PostquantumVpn = test.postQuantum

			mockRPCServer := &mockRPCServer{}
			err := rpc.Connect(&pb.ConnectRequest{ServerTag: "dedicated_server"}, mockRPCServer)

			assert.Equal(t, test.expectedErr, err, "Unexpected error returned by the Connect RPC.")
			assert.Equal(t, test.expectedConnectRequest, mockDedicatedServersAPI.ConnectRequest,
				"Invalid connect request.")
			assert.Equal(t, test.expectedConnectRequestUUID, mockDedicatedServersAPI.ConnectServerUUID,
				"Connect request was sent to invalid server.")

			if test.expectedErr == nil {
				assert.Equal(t, test.expectedStatus, mockRPCServer.msg.Type,
					"Unexpected status in a Connect RPC response.")
			} else {
				assert.Nil(t, mockRPCServer.msg, "Unexpected Connect RPC response.")
			}
		})
	}
}

func TestDedicatedServers_Internals(t *testing.T) {
	category.Set(t, category.Unit)

	serverUUID := uuid.MustParse("af0bc2b1-785a-4455-bfe0-5397c39c4f4e")
	serverAddress := "100.34.50.2"
	serverPort := int64(55555)
	dedicatedServer := core.DedicatedServer{
		UUID:   serverUUID.String(),
		Name:   "server 1",
		Status: core.DedicatedServerStatusRunning,
		IP:     serverAddress,
	}
	dedicatedServers := core.DedicatedServers{dedicatedServer}

	serverPublicKey := "public_key"
	connectResponse := core.DedicatedServerConnectResponse{
		ServerEndpoint:  serverAddress + ":" + strconv.Itoa(int(serverPort)),
		ServerPublicKey: serverPublicKey,
	}

	deviceKey := "device key"
	devicePrivateKey := "device private key"

	rpc := testRPCLocal(t)
	rpc.dedicatedServersAPI = &core_test.DedicatedServersAPIMock{
		DedicatedServersResponse: dedicatedServers,
		ConnectResponse:          connectResponse,
	}
	rpc.ac = &workingLoginChecker{
		isDedicatedServersExpired: false,
	}
	rpc.dedicatedServerKeyManager = &testdevicekey.MockDeviceKeyManager{
		DedicatedServerRegistrationData: &devicekey.DedicatedServersConnectionData{
			DevicePublicKey:  deviceKey,
			DevicePrivateKey: devicePrivateKey,
		},
	}

	configManagerMock := rpc.cm.(*mockConfigManager)
	configManagerMock.c.Technology = config.Technology_NORDLYNX

	networkerMock := rpc.netw.(*testnetworker.Mock)

	mockRPCServer := &mockRPCServer{}
	err := rpc.Connect(&pb.ConnectRequest{ServerTag: "dedicated_server"}, mockRPCServer)
	assert.Nil(t, err, "Unexpected error returned by Connect.")

	assert.Equal(t, devicePrivateKey, networkerMock.ProvidedCredentials.NordLynxPrivateKey,
		"DeviceKey should be used in place of NordlynxPrivateKey in case of dedicated server connections.")
	assert.Equal(t, serverPort, networkerMock.ProvidedServerData.DedicatedServerPort)
}

func TestDedicatedServers_ForceRegistration(t *testing.T) {
	category.Set(t, category.Unit)

	serverUUID := uuid.MustParse("af0bc2b1-785a-4455-bfe0-5397c39c4f4e")
	serverAddress := "44.44.44.44"
	serverPort := int64(55555)
	dedicatedServer := core.DedicatedServer{
		UUID:   serverUUID.String(),
		Name:   "server 1",
		Status: core.DedicatedServerStatusRunning,
		IP:     serverAddress,
	}
	dedicatedServers := core.DedicatedServers{dedicatedServer}

	serverPublicKey := "server_public_key"
	rpc := testRPCLocal(t)

	connectResponse := core.DedicatedServerConnectResponse{
		ServerEndpoint:  serverAddress + ":" + strconv.Itoa(int(serverPort)),
		ServerPublicKey: serverPublicKey,
	}

	connectErr := core.ErrDedicatedServersDeviceNotFound
	getConnectErrFunc := func() error {
		err := connectErr
		// set to nil so that no error will be returned for the next call
		connectErr = nil
		return err
	}

	dedicatedServersAPIMock := core_test.DedicatedServersAPIMock{
		DedicatedServersResponse: dedicatedServers,
		ConnectResponse:          connectResponse,
		GetConnectErrFunc:        getConnectErrFunc,
	}
	rpc.dedicatedServersAPI = &dedicatedServersAPIMock

	rpc.ac = &workingLoginChecker{
		isDedicatedServersExpired: false,
	}

	staleDeviceKey := "stale device key"
	staleDevicePrivateKey := "stale device private key"

	newDeviceKey := "new device key"
	newDevicePrivateKey := "new device private key"

	deviceKeyManagerMock := testdevicekey.MockDeviceKeyManager{
		DedicatedServerRegistrationData: &devicekey.DedicatedServersConnectionData{
			DevicePublicKey:  staleDeviceKey,
			DevicePrivateKey: staleDevicePrivateKey,
		},
		DedicatedServerForcedRegistrationData: &devicekey.DedicatedServersConnectionData{
			DevicePublicKey:  newDeviceKey,
			DevicePrivateKey: newDevicePrivateKey,
		},
	}
	rpc.dedicatedServerKeyManager = &deviceKeyManagerMock

	configManagerMock := rpc.cm.(*mockConfigManager)
	configManagerMock.c.Technology = config.Technology_NORDLYNX

	mockRPCServer := &mockRPCServer{}
	err := rpc.Connect(&pb.ConnectRequest{ServerTag: "dedicated_server"}, mockRPCServer)
	assert.Nil(t, err, "Unexpected error returned by Connect.")

	assert.True(t, deviceKeyManagerMock.WasKeyForceRegistered, "Key should be forcibly registered when API returns %w",
		core.ErrDedicatedServersDeviceNotFound)
	assert.Equal(t, newDeviceKey, dedicatedServersAPIMock.ConnectRequest.DevicePublicKey,
		"Key sent in a connect request should be equal to newly registered key.")

	networkerMock := rpc.netw.(*testnetworker.Mock)
	assert.Equal(t, newDevicePrivateKey, networkerMock.ProvidedCredentials.NordLynxPrivateKey,
		"Key used to connect to the VPN server should be equal to newly registered key.")
}

func Test_serverGroupIDs_ExtractsAllIDs(t *testing.T) {
	category.Set(t, category.Unit)
	server := core.Server{Groups: []core.Group{
		{ID: config.ServerGroup_DEDICATED_IP, Title: "Dedicated IP"},
		{ID: config.ServerGroup_STANDARD_VPN_SERVERS, Title: "Standard VPN servers"},
	}}

	got := determineServerGroupIDs(&server)

	want := []config.ServerGroup{
		config.ServerGroup_DEDICATED_IP,
		config.ServerGroup_STANDARD_VPN_SERVERS,
	}
	assert.Equal(t, want, got)
}

func Test_serverGroupIDs_EmptyGroups_ReturnsEmptySlice(t *testing.T) {
	category.Set(t, category.Unit)
	server := core.Server{Groups: []core.Group{}}

	got := determineServerGroupIDs(&server)

	assert.Equal(t, 0, len(got))
}

func Test_serverGroupIDs_NilGroups_ReturnsEmptySlice(t *testing.T) {
	category.Set(t, category.Unit)
	server := core.Server{}

	got := determineServerGroupIDs(&server)

	assert.Equal(t, 0, len(got))
}
