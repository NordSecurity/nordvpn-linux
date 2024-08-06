package daemon

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestFilterServers(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name          string
		servers       core.Servers
		tech          config.Technology
		proto         config.Protocol
		group         config.ServerGroup
		expectedCount int
		hasError      bool
	}{
		{
			name:          "nil servers",
			servers:       nil,
			expectedCount: 0,
			hasError:      true,
		},
		{
			name:          "0 servers",
			servers:       core.Servers{},
			expectedCount: 0,
			hasError:      true,
		},
		{
			name: "NordLynx 0 servers",
			servers: core.Servers{
				core.Server{
					Status: core.Online,
					Technologies: core.Technologies{
						core.Technology{
							ID:    core.OpenVPNUDP,
							Pivot: core.Pivot{Status: core.Online},
						},
					},
					Groups: core.Groups{
						core.Group{ID: config.StandardVPNServers},
					},
				},
			},
			tech:          config.Technology_NORDLYNX,
			proto:         config.Protocol_UDP,
			expectedCount: 0,
			hasError:      true,
		},
		{
			name: "NordLynx several online servers",
			servers: core.Servers{
				core.Server{
					Status: core.Online,
					Technologies: core.Technologies{
						core.Technology{
							ID:    core.WireguardTech,
							Pivot: core.Pivot{Status: core.Online},
						},
					},
					Groups: core.Groups{
						core.Group{ID: config.StandardVPNServers},
					},
				},
				core.Server{
					Status: core.Online,
					Technologies: core.Technologies{
						core.Technology{
							ID:    core.WireguardTech,
							Pivot: core.Pivot{Status: core.Online},
						},
					},
				},
			},
			tech:          config.Technology_NORDLYNX,
			proto:         config.Protocol_UDP,
			expectedCount: 1,
			hasError:      false,
		},
		{
			name: "NordLynx several mixed status servers",
			servers: core.Servers{
				core.Server{
					Status: core.Online,
					Technologies: core.Technologies{
						core.Technology{
							ID:    core.WireguardTech,
							Pivot: core.Pivot{Status: core.Online},
						},
					},
					Groups: core.Groups{
						core.Group{ID: config.StandardVPNServers},
					},
				},
				core.Server{
					Status: core.Offline,
					Technologies: core.Technologies{
						core.Technology{
							ID:    core.WireguardTech,
							Pivot: core.Pivot{Status: core.Online},
						},
					},
					Groups: core.Groups{
						core.Group{ID: config.StandardVPNServers},
					},
				},
				core.Server{
					Status: core.Maintenance,
					Technologies: core.Technologies{
						core.Technology{
							ID:    core.WireguardTech,
							Pivot: core.Pivot{Status: core.Online},
						},
					},
					Groups: core.Groups{
						core.Group{ID: config.StandardVPNServers},
					},
				},
			},
			tech:          config.Technology_NORDLYNX,
			proto:         config.Protocol_UDP,
			expectedCount: 1,
			hasError:      false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			servers, err := filterServers(
				test.servers,
				test.tech,
				test.proto,
				"",
				test.group,
				false,
			)
			assert.Equal(t, test.hasError, err != nil)
			assert.Equal(t, test.expectedCount, len(servers))
		})
	}
}

func TestResolveServerGroup(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		flag  string
		tag   string
		group config.ServerGroup
		err   error
	}{
		{
			"",
			"",
			config.UndefinedGroup,
			nil,
		},
		{
			"p2p",
			"",
			config.P2P,
			nil,
		},
		{
			"",
			"p2p",
			config.P2P,
			nil,
		},
		{
			"p2p",
			"p2p",
			config.UndefinedGroup,
			internal.ErrDoubleGroup,
		},
		{
			"quantum_vpn",
			"",
			config.UndefinedGroup,
			internal.ErrGroupDoesNotExist,
		},
		{
			"",
			"quantum_vpn",
			config.UndefinedGroup,
			nil,
		},
		{
			"quantum_vpn",
			"quantum_vpn",
			config.UndefinedGroup,
			internal.ErrGroupDoesNotExist,
		},
	}

	for _, tt := range tests {
		group, err := resolveServerGroup(tt.flag, tt.tag)
		assert.Equal(t, tt.group, group)
		assert.Equal(t, tt.err, err)
	}
}

func TestGroupConvert(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		input  string
		output config.ServerGroup
	}{
		{
			"Double VPN",
			config.DoubleVPN,
		},
		{
			"Onion Over VPN",
			config.OnionOverVPN,
		},
		{
			"DEDICATED IP",
			config.DedicatedIP,
		},
		{
			"sTANDARD vPn SeRvErS",
			config.StandardVPNServers,
		},
		{
			"P2P",
			config.P2P,
		},
		{
			"Europe",
			config.Europe,
		},
		{
			"The Americas",
			config.TheAmericas,
		},
		{
			"Asia Pacific",
			config.AsiaPacific,
		},
		{
			"Africa The Middle East And India",
			config.AfricaMiddleEastIndia,
		},
		{
			"neflix & chill",
			config.UndefinedGroup,
		},
	}

	for _, tt := range tests {
		got := groupConvert(tt.input)
		assert.Equal(t, tt.output, got)
	}
}

func TestServerTagToServerBy(t *testing.T) {
	category.Set(t, category.Unit)

	t.Run("action fastest", func(t *testing.T) {
		tag := "double_vpn"
		server := core.Server{}
		groups := core.Groups{}
		locs := core.Locations{}

		groups = append(groups, core.Group{Title: "Double VPN"})
		server.Groups = groups

		locs = append(locs, core.Location{Country: core.Country{Name: "China", City: core.City{Name: "Beijing"}}})
		server.Locations = locs

		got := serverTagToServerBy(tag, server)
		assert.Equal(t, core.ServerBySpeed, got)
	})

	t.Run("action server", func(t *testing.T) {
		tag := "de666"
		server := core.Server{Hostname: "de666.nordvpn.com"}
		locs := core.Locations{}

		locs = append(locs, core.Location{Country: core.Country{Name: "Germany", City: core.City{Name: "Frankfurt"}}})
		server.Locations = locs

		got := serverTagToServerBy(tag, server)
		assert.Equal(t, core.ServerByName, got)
	})

	t.Run("action country", func(t *testing.T) {
		tag := "Japan"
		server := core.Server{}
		locs := core.Locations{}

		locs = append(locs, core.Location{Country: core.Country{Name: "Japan"}})
		server.Locations = locs

		got := serverTagToServerBy(tag, server)
		assert.Equal(t, core.ServerByCountry, got)
	})

	t.Run("action fastest", func(t *testing.T) {
		tag := ""
		server := core.Server{}
		locs := core.Locations{}

		locs = append(locs, core.Location{Country: core.Country{}})
		server.Locations = locs

		got := serverTagToServerBy(tag, server)
		assert.Equal(t, core.ServerBySpeed, got)
	})

	t.Run("action city", func(t *testing.T) {
		tag := "London"
		server := core.Server{}
		locs := core.Locations{}

		locs = append(locs, core.Location{Country: core.Country{City: core.City{Name: "London"}}})
		server.Locations = locs

		got := serverTagToServerBy(tag, server)
		assert.Equal(t, core.ServerByCity, got)
	})
}

func TestServerTagFromString(t *testing.T) {
	category.Set(t, category.File)
	defer testsCleanup()

	tests := []struct {
		name      string
		countries core.Countries
		servers   core.Servers
		tag       string
		group     config.ServerGroup
		groupFlag bool
		expected  core.ServerTag
		hasError  bool
	}{
		{
			name:      "empty tag",
			countries: core.Countries{},
			servers:   core.Servers{},
			tag:       "",
			group:     config.UndefinedGroup,
			expected:  core.ServerTag{Action: core.ServerByUnknown, ID: 0},
			hasError:  false,
		},
		{
			name:      "group tag",
			countries: core.Countries{},
			servers:   core.Servers{},
			tag:       "Europe",
			group:     config.Europe,
			expected:  core.ServerTag{Action: core.ServerBySpeed, ID: 19},
			hasError:  false,
		},
		{
			name: "country tag",
			countries: core.Countries{
				core.Country{
					ID:   202,
					Name: "Spain",
					Code: "ES",
					Cities: []core.City{
						{
							ID:        2619989,
							Name:      "Madrid",
							Latitude:  40.408566,
							Longitude: -3.69222,
						},
					},
				},
			},
			servers:  core.Servers{},
			tag:      "Spain",
			group:    config.UndefinedGroup,
			expected: core.ServerTag{Action: core.ServerByCountry, ID: 202},
			hasError: false,
		},
		{
			name: "city tag",
			countries: core.Countries{
				core.Country{
					ID:   202,
					Name: "Spain",
					Code: "ES",
					Cities: []core.City{
						{
							ID:        2619989,
							Name:      "Madrid",
							Latitude:  40.408566,
							Longitude: -3.69222,
						},
					},
				},
			},
			servers:  core.Servers{},
			tag:      "Madrid",
			group:    config.UndefinedGroup,
			expected: core.ServerTag{Action: core.ServerByCity, ID: 2619989},
			hasError: false,
		},
		{
			name: "country & city tag",
			countries: core.Countries{
				core.Country{
					ID:   202,
					Name: "Spain",
					Code: "ES",
					Cities: []core.City{
						{
							ID:        2619989,
							Name:      "Madrid",
							Latitude:  40.408566,
							Longitude: -3.69222,
						},
					},
				},
			},
			servers:  core.Servers{},
			tag:      "Spain Madrid",
			group:    config.UndefinedGroup,
			expected: core.ServerTag{Action: core.ServerByCity, ID: 2619989},
			hasError: false,
		},
		{
			name: "specific tag",
			countries: core.Countries{
				core.Country{
					ID:   202,
					Name: "Spain",
					Code: "ES",
					Cities: []core.City{
						{
							ID:        2619989,
							Name:      "Madrid",
							Latitude:  40.408566,
							Longitude: -3.69222,
						},
					},
				},
			},
			servers: core.Servers{
				core.Server{
					ID:       929912,
					Name:     "Canada #944",
					Hostname: "ca944.nordvpn.com",
				},
			},
			tag:      "ca944",
			group:    config.UndefinedGroup,
			expected: core.ServerTag{Action: core.ServerByName, ID: 929912},
			hasError: false,
		},
		{
			name: "group flag",
			countries: core.Countries{
				core.Country{
					ID:   202,
					Name: "Spain",
					Code: "ES",
					Cities: []core.City{
						{
							ID:        2619989,
							Name:      "Madrid",
							Latitude:  40.408566,
							Longitude: -3.69222,
						},
					},
				},
			},
			servers: core.Servers{
				core.Server{
					ID:       929912,
					Name:     "Canada #944",
					Hostname: "ca944.nordvpn.com",
				},
			},
			tag:       "",
			group:     config.Europe,
			groupFlag: true,
			expected:  core.ServerTag{Action: core.ServerByUnknown, ID: 0},
			hasError:  false,
		},
		{
			name: "country tag & group flag",
			countries: core.Countries{
				core.Country{
					ID:   202,
					Name: "Spain",
					Code: "ES",
					Cities: []core.City{
						{
							ID:        2619989,
							Name:      "Madrid",
							Latitude:  40.408566,
							Longitude: -3.69222,
						},
					},
				},
			},
			servers: core.Servers{
				core.Server{
					ID:       929912,
					Name:     "Canada #944",
					Hostname: "ca944.nordvpn.com",
				},
			},
			tag:       "Spain",
			group:     config.Europe,
			groupFlag: true,
			expected:  core.ServerTag{Action: core.ServerByCountry, ID: 202},
			hasError:  false,
		},
		{
			name: "non-existing tag",
			countries: core.Countries{
				core.Country{
					ID:   202,
					Name: "Spain",
					Code: "ES",
					Cities: []core.City{
						{
							ID:        2619989,
							Name:      "Madrid",
							Latitude:  40.408566,
							Longitude: -3.69222,
						},
					},
				},
			},
			servers: core.Servers{
				core.Server{
					ID:       929912,
					Name:     "Canada #944",
					Hostname: "ca944.nordvpn.com",
				},
			},
			tag:      "matrix",
			group:    config.UndefinedGroup,
			expected: core.ServerTag{},
			hasError: true,
		},
		{
			name: "wrong country & city tag",
			countries: core.Countries{
				core.Country{
					ID:   202,
					Name: "Spain",
					Code: "ES",
					Cities: []core.City{
						{
							ID:        2619989,
							Name:      "Madrid",
							Latitude:  40.408566,
							Longitude: -3.69222,
						},
					},
				},
				core.Country{
					ID:   175,
					Name: "Portugal",
					Code: "PT",
					Cities: []core.City{
						{
							ID:        6906665,
							Name:      "Lisbon",
							Latitude:  38.716667,
							Longitude: -9.133333,
						},
					},
				},
			},
			servers:  core.Servers{},
			tag:      "Spain Lisbon",
			group:    config.UndefinedGroup,
			expected: core.ServerTag{},
			hasError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := serverTagFromString(test.countries, &mockServersAPI{}, test.tag, test.group, test.servers, test.groupFlag)
			assert.Equal(t, test.hasError, err != nil)
			assert.EqualValues(t, test.expected, got)
		})
	}
}

func TestTechToServerTech(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name       string
		tech       config.Technology
		protocol   config.Protocol
		obfuscated bool
		expected   core.ServerTechnology
	}{
		{
			name:       "wireguard",
			tech:       config.Technology_NORDLYNX,
			protocol:   config.Protocol_UDP,
			obfuscated: false,
			expected:   core.WireguardTech,
		},
		{
			name:       "obfuscated tpc",
			tech:       config.Technology_OPENVPN,
			protocol:   config.Protocol_TCP,
			obfuscated: true,
			expected:   core.OpenVPNTCPObfuscated,
		},
		{
			name:       "obfuscated udp",
			tech:       config.Technology_OPENVPN,
			protocol:   config.Protocol_UDP,
			obfuscated: true,
			expected:   core.OpenVPNUDPObfuscated,
		},
		{
			name:       "openvpn tcp",
			tech:       config.Technology_OPENVPN,
			protocol:   config.Protocol_TCP,
			obfuscated: false,
			expected:   core.OpenVPNTCP,
		},
		{
			name:       "openvpn udp",
			tech:       config.Technology_OPENVPN,
			protocol:   config.Protocol_UDP,
			obfuscated: false,
			expected:   core.OpenVPNUDP,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := techToServerTech(test.tech, test.protocol, test.obfuscated)
			assert.Equal(t, test.expected, got)
		})
	}
}

func TestGetSystemInfo(t *testing.T) {
	category.Set(t, category.Integration)
	str := getSystemInfo("0.0.0")
	assert.Contains(t, str, "App Version:")
	assert.Contains(t, str, "OS Info:")
	assert.Contains(t, str, "System Info:")
}

func TestGetNetworkInfo(t *testing.T) {
	category.Set(t, category.Route, category.Firewall)
	str := getNetworkInfo()
	assert.Contains(t, str, "Routes for ipv4")
	assert.Contains(t, str, "Routes for ipv6")
	assert.Contains(t, str, "IP rules for ipv4")
	assert.Contains(t, str, "IP rules for ipv6")
	assert.Contains(t, str, "IP tables for ipv4")
	assert.Contains(t, str, "IP tables for ipv6")
}

func TestPickServer(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		name                 string
		api                  core.ServersAPI
		servers              core.Servers
		longitude            float64
		latitude             float64
		tech                 config.Technology
		protocol             config.Protocol
		obfuscated           bool
		tag                  string
		groupFlag            string
		onlyPhysicServers    bool
		expectedServerName   string
		expectedRemoteServer bool
		expectedError        error
	}{
		{
			name:               "find server using country code",
			api:                mockFailingServersAPI{},
			servers:            serversList(),
			tech:               config.Technology_NORDLYNX,
			tag:                "de",
			expectedServerName: "Germany #3",
		},
		{
			name:               "find server using country name",
			api:                mockFailingServersAPI{},
			servers:            serversList(),
			tech:               config.Technology_NORDLYNX,
			tag:                "germany",
			expectedServerName: "Germany #3",
		},
		{
			name:               "find server using city name",
			api:                mockFailingServersAPI{},
			servers:            serversList(),
			tech:               config.Technology_NORDLYNX,
			tag:                "berlin",
			expectedServerName: "Germany #3",
		},
		{
			name:               "find server using country + city name",
			api:                mockFailingServersAPI{},
			servers:            serversList(),
			tech:               config.Technology_NORDLYNX,
			tag:                "germany berlin",
			expectedServerName: "Germany #3",
		},
		{
			name:               "find server using country code + city name",
			api:                mockFailingServersAPI{},
			servers:            serversList(),
			tech:               config.Technology_NORDLYNX,
			tag:                "de berlin",
			expectedServerName: "Germany #3",
		},
		{
			name:              "find server when virtual locations are disabled",
			api:               mockFailingServersAPI{},
			servers:           serversList(),
			tech:              config.Technology_NORDLYNX,
			onlyPhysicServers: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server, remote, err := PickServer(test.api,
				countriesList(), test.servers, test.longitude, test.latitude, test.tech, test.protocol, test.obfuscated, test.tag, test.groupFlag, !test.onlyPhysicServers)

			assert.Equal(t, test.expectedError, err)
			assert.Equal(t, test.expectedRemoteServer, remote)
			if len(test.expectedServerName) > 0 {
				assert.Equal(t, test.expectedServerName, server.Name)
			}
		})
	}
}
