package serverpicker

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	core_test "github.com/NordSecurity/nordvpn-linux/test/mock/core"

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
						core.Group{ID: config.ServerGroup_STANDARD_VPN_SERVERS},
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
						core.Group{ID: config.ServerGroup_STANDARD_VPN_SERVERS},
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
						core.Group{ID: config.ServerGroup_STANDARD_VPN_SERVERS},
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
						core.Group{ID: config.ServerGroup_STANDARD_VPN_SERVERS},
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
						core.Group{ID: config.ServerGroup_STANDARD_VPN_SERVERS},
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
			localSelFn := selectFilterForLocalServers("", test.group)
			filterFn := func(s core.Server) bool {
				return core.IsConnectableWithProtocol(test.tech, test.proto)(s) &&
					localSelFn(s)
			}
			servers, err := findServersLocally(test.servers, core.ServerTag{Action: core.ServerByUnknown}, filterFn)
			assert.Equal(t, test.hasError, err != nil)
			assert.Equal(t, test.expectedCount, len(servers))
		})
	}
}

func TestResolveServerGroup(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		input         SearchParams
		tagChanged    bool
		expectedGroup config.ServerGroup
		err           error
	}{
		{
			input:         SearchParams{},
			expectedGroup: config.ServerGroup_UNDEFINED,
			err:           nil,
		},
		{
			input:         NewSearchParams("", "double_vpn", ""),
			expectedGroup: config.ServerGroup_DOUBLE_VPN,
			err:           nil,
		},
		{
			input:         NewSearchParams("double_vpn", "", ""),
			tagChanged:    true,
			expectedGroup: config.ServerGroup_DOUBLE_VPN,
			err:           nil,
		},
		{
			input:         NewSearchParams("double_vpn", "double_vpn", ""),
			expectedGroup: config.ServerGroup_UNDEFINED,
			err:           internal.ErrDoubleGroup,
		},
		{
			input:         NewSearchParams("double_vpn", "quantum_vpn", ""),
			expectedGroup: config.ServerGroup_UNDEFINED,
			err:           internal.ErrGroupDoesNotExist,
		},
		{
			input:         NewSearchParams("quantum_vpn", "double_vpn", ""),
			expectedGroup: config.ServerGroup_DOUBLE_VPN,
			err:           nil,
		},
		{
			input:         NewSearchParams("quantum_vpn", "", ""),
			expectedGroup: config.ServerGroup_UNDEFINED,
			err:           nil,
		},
		{
			input:         NewSearchParams("double_vpn us1234", "", ""),
			expectedGroup: config.ServerGroup_UNDEFINED,
			err:           nil,
		},
		{
			input:         NewSearchParams("quantum_vpn", "quantum_vpn", ""),
			expectedGroup: config.ServerGroup_UNDEFINED,
			err:           internal.ErrGroupDoesNotExist,
		},
		{
			input:         NewSearchParams("", "Obfuscated", ""),
			expectedGroup: config.ServerGroup_NW_OBFUSCATED,
			err:           nil,
		},
	}

	for _, tt := range tests {
		normalized, group, err := resolveServerGroup(tt.input)
		assert.Equal(t, tt.expectedGroup, group)
		assert.Equal(t, tt.err, err)
		assert.Equal(t, tt.tagChanged, normalized.Tag != tt.input.Tag)
		if tt.tagChanged {
			assert.Empty(t, normalized.Tag)
		}
		// only the tag may be rewritten
		assert.Equal(t, tt.input.Group, normalized.Group)
		assert.Equal(t, tt.input.ExcludedServer, normalized.ExcludedServer)
	}
}

func TestIsObfuscatedTech(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		tech     config.Technology
		expected bool
	}{
		{tech: config.Technology_NORDWHISPER, expected: true},
		{tech: config.Technology_NORDLYNX, expected: false},
		{tech: config.Technology_OPENVPN, expected: false},
		{tech: config.Technology_UNKNOWN_TECHNOLOGY, expected: false},
	}

	for _, test := range tests {
		t.Run(test.tech.String(), func(t *testing.T) {
			assert.Equal(t, test.expected, IsObfuscatedTech(test.tech))
		})
	}
}

func TestSearchGroup(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name      string
		requested config.ServerGroup
		tech      config.Technology
		expected  config.ServerGroup
	}{
		{
			name:      "obfuscated over nordwhisper is searched as standard",
			requested: config.ServerGroup_NW_OBFUSCATED,
			tech:      config.Technology_NORDWHISPER,
			expected:  config.ServerGroup_STANDARD_VPN_SERVERS,
		},
		{
			name:      "obfuscated over nordlynx is left alone",
			requested: config.ServerGroup_NW_OBFUSCATED,
			tech:      config.Technology_NORDLYNX,
			expected:  config.ServerGroup_NW_OBFUSCATED,
		},
		{
			name:      "obfuscated over openvpn is left alone",
			requested: config.ServerGroup_NW_OBFUSCATED,
			tech:      config.Technology_OPENVPN,
			expected:  config.ServerGroup_NW_OBFUSCATED,
		},
		{
			name:      "other groups are never touched",
			requested: config.ServerGroup_DOUBLE_VPN,
			tech:      config.Technology_NORDWHISPER,
			expected:  config.ServerGroup_DOUBLE_VPN,
		},
		{
			name:      "standard stays standard",
			requested: config.ServerGroup_STANDARD_VPN_SERVERS,
			tech:      config.Technology_NORDWHISPER,
			expected:  config.ServerGroup_STANDARD_VPN_SERVERS,
		},
		{
			name:      "undefined stays undefined",
			requested: config.ServerGroup_UNDEFINED,
			tech:      config.Technology_NORDWHISPER,
			expected:  config.ServerGroup_UNDEFINED,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, searchGroup(test.requested, test.tech))
		})
	}
}

func TestEffectiveGroups(t *testing.T) {
	category.Set(t, category.Unit)

	standard := core.Group{ID: config.ServerGroup_STANDARD_VPN_SERVERS, Title: "Standard VPN servers"}
	doubleVpn := core.Group{ID: config.ServerGroup_DOUBLE_VPN, Title: "P2P"}
	xor := core.Group{ID: config.ServerGroup_OVPN_OBFUSCATED, Title: "Obfuscated"}
	obfuscated := core.Group{ID: config.ServerGroup_NW_OBFUSCATED, Title: ObfuscatedServersGroupTitle}

	nordWhisper := core.Technologies{{ID: core.NordWhisperTech, Pivot: core.Pivot{Status: core.Online}}}
	nordWhisperOffline := core.Technologies{{ID: core.NordWhisperTech, Pivot: core.Pivot{Status: core.Offline}}}
	wireguardOnly := core.Technologies{{ID: core.WireguardTech, Pivot: core.Pivot{Status: core.Online}}}

	tests := []struct {
		name         string
		groups       core.Groups
		technologies core.Technologies
		status       core.Status
		tech         config.Technology
		expected     core.Groups
	}{
		{
			name:         "standard nordwhisper server over nordwhisper offers the obfuscated group too",
			groups:       core.Groups{standard},
			technologies: nordWhisper,
			status:       core.Online,
			tech:         config.Technology_NORDWHISPER,
			expected:     core.Groups{standard, obfuscated},
		},
		{
			name:         "standard server over nordlynx is left alone",
			groups:       core.Groups{standard},
			technologies: nordWhisper,
			status:       core.Online,
			tech:         config.Technology_NORDLYNX,
			expected:     core.Groups{standard},
		},
		{
			name:         "non-standard nordwhisper server over nordwhisper is not obfuscated",
			groups:       core.Groups{doubleVpn},
			technologies: nordWhisper,
			status:       core.Online,
			tech:         config.Technology_NORDWHISPER,
			expected:     core.Groups{doubleVpn},
		},
		{
			name:         "standard server not reachable over nordwhisper is not obfuscated",
			groups:       core.Groups{standard},
			technologies: wireguardOnly,
			status:       core.Online,
			tech:         config.Technology_NORDWHISPER,
			expected:     core.Groups{standard},
		},
		{
			name:         "standard server with nordwhisper offline is not obfuscated",
			groups:       core.Groups{standard},
			technologies: nordWhisperOffline,
			status:       core.Online,
			tech:         config.Technology_NORDWHISPER,
			expected:     core.Groups{standard},
		},
		{
			name:         "offline standard nordwhisper server is not obfuscated",
			groups:       core.Groups{standard},
			technologies: nordWhisper,
			status:       core.Offline,
			tech:         config.Technology_NORDWHISPER,
			expected:     core.Groups{standard},
		},
		{
			name:         "retired XOR tag is dropped under openvpn",
			groups:       core.Groups{xor, standard},
			technologies: nordWhisper,
			status:       core.Online,
			tech:         config.Technology_OPENVPN,
			expected:     core.Groups{standard},
		},
		{
			name:         "retired XOR tag is replaced under nordwhisper, not duplicated",
			groups:       core.Groups{xor, standard},
			technologies: nordWhisper,
			status:       core.Online,
			tech:         config.Technology_NORDWHISPER,
			expected:     core.Groups{standard, obfuscated},
		},
		{
			name:         "server without groups stays without groups",
			groups:       core.Groups{},
			technologies: nordWhisper,
			status:       core.Online,
			tech:         config.Technology_NORDWHISPER,
			expected:     core.Groups{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := core.Server{
				Groups:       slices.Clone(test.groups),
				Technologies: test.technologies,
				Status:       test.status,
			}

			got := EffectiveGroups(server, test.tech)

			assert.Equal(t, test.expected, got)
			assert.Equal(t, test.groups, server.Groups)
		})
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
			config.ServerGroup_DOUBLE_VPN,
		},
		{
			"Onion Over VPN",
			config.ServerGroup_ONION_OVER_VPN,
		},
		{
			"DEDICATED IP",
			config.ServerGroup_DEDICATED_IP,
		},
		{
			"sTANDARD vPn SeRvErS",
			config.ServerGroup_STANDARD_VPN_SERVERS,
		},
		{
			"P2P",
			config.ServerGroup_UNDEFINED,
		},
		{
			"Europe",
			config.ServerGroup_UNDEFINED,
		},
		{
			"The Americas",
			config.ServerGroup_UNDEFINED,
		},
		{
			"Asia Pacific",
			config.ServerGroup_UNDEFINED,
		},
		{
			"Africa The Middle East And India",
			config.ServerGroup_UNDEFINED,
		},
		{
			"netflix & chill",
			config.ServerGroup_UNDEFINED,
		},
	}

	for _, tt := range tests {
		got := groupConvert(tt.input)
		assert.Equal(t, tt.output, got)
	}
}

func TestServerTagFromString(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name      string
		countries core.Countries
		servers   core.Servers
		tag       string
		group     config.ServerGroup
		expected  core.ServerTag
		hasError  bool
	}{
		{
			name:      "empty tag",
			countries: core.Countries{},
			servers:   core.Servers{},
			tag:       "",
			group:     config.ServerGroup_UNDEFINED,
			expected:  core.ServerTag{Action: core.ServerByUnknown, ID: 0},
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
			group:    config.ServerGroup_UNDEFINED,
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
			group:    config.ServerGroup_UNDEFINED,
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
			group:    config.ServerGroup_UNDEFINED,
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
			group:    config.ServerGroup_UNDEFINED,
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
			tag:      "",
			group:    config.ServerGroup_DOUBLE_VPN,
			expected: core.ServerTag{Action: core.ServerBySpeed, ID: int64(config.ServerGroup_DOUBLE_VPN)},
			hasError: false,
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
			tag:      "Spain",
			group:    config.ServerGroup_DOUBLE_VPN,
			expected: core.ServerTag{Action: core.ServerByCountry, ID: 202},
			hasError: false,
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
			group:    config.ServerGroup_UNDEFINED,
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
			group:    config.ServerGroup_UNDEFINED,
			expected: core.ServerTag{},
			hasError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := serverTagFromString(test.tag, test.group, test.countries, test.servers)
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
			got := TechToServerTech(test.tech, test.protocol)
			assert.Equal(t, test.expected, got)
		})
	}
}

func TestPickServer(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		name                 string
		api                  core.ServersAPI
		servers              core.Servers
		tech                 config.Technology
		tag                  string
		group                string
		excludedServer       string
		expectedServerName   string
		expectedRemoteServer bool
		expectedError        error
	}{
		{
			name:               "find server using country code",
			api:                core_test.NewMockFailingServersAPI(errors.New("500")),
			servers:            core_test.ServersList(),
			tech:               config.Technology_NORDLYNX,
			tag:                "de",
			expectedServerName: "Germany #3",
		},
		{
			name:               "find server using country name",
			api:                core_test.NewMockFailingServersAPI(errors.New("500")),
			servers:            core_test.ServersList(),
			tech:               config.Technology_NORDLYNX,
			tag:                "germany",
			expectedServerName: "Germany #3",
		},
		{
			name:               "find server using city name",
			api:                core_test.NewMockFailingServersAPI(errors.New("500")),
			servers:            core_test.ServersList(),
			tech:               config.Technology_NORDLYNX,
			tag:                "berlin",
			expectedServerName: "Germany #3",
		},
		{
			name:               "find server using country + city name",
			api:                core_test.NewMockFailingServersAPI(errors.New("500")),
			servers:            core_test.ServersList(),
			tech:               config.Technology_NORDLYNX,
			tag:                "germany berlin",
			expectedServerName: "Germany #3",
		},
		{
			name:               "find server using country code + city name",
			api:                core_test.NewMockFailingServersAPI(errors.New("500")),
			servers:            core_test.ServersList(),
			tech:               config.Technology_NORDLYNX,
			tag:                "de berlin",
			expectedServerName: "Germany #3",
		},
		{
			name:                 "server selected from the API is marked as remote",
			api:                  core_test.NewMockServersAPI(),
			servers:              core_test.ServersList(),
			tech:                 config.Technology_NORDLYNX,
			tag:                  "de3",
			expectedServerName:   "Germany #3",
			expectedRemoteServer: true,
		},
		{
			name:          "can't find a server",
			api:           core_test.NewMockFailingServersAPI(errors.New("500")),
			servers:       core.Servers{},
			tech:          config.Technology_NORDLYNX,
			expectedError: internal.ErrServerIsUnavailable,
		},
		{
			name:           "exclude server de3.nordvpn.com",
			api:            core_test.NewMockFailingServersAPI(errors.New("500")),
			servers:        core_test.ServersList(),
			tech:           config.Technology_NORDLYNX,
			tag:            "de berlin",
			excludedServer: "de3.nordvpn.com",
			expectedError:  internal.ErrServerIsUnavailable,
		},
		// NordWhisper as Obfuscated: Germany #4 is the only NordWhisper capable server in the mock
		{
			name:               "obfuscated group over nordwhisper is served by the standard servers",
			api:                core_test.NewMockFailingServersAPI(errors.New("500")),
			servers:            core_test.ServersList(),
			tech:               config.Technology_NORDWHISPER,
			group:              "obfuscated",
			expectedServerName: "Germany #4",
		},
		{
			name:                 "obfuscated group over nordwhisper is translated before the API is asked",
			api:                  core_test.NewMockServersAPI(),
			servers:              core_test.ServersList(),
			tech:                 config.Technology_NORDWHISPER,
			group:                "obfuscated",
			expectedServerName:   "Germany #4",
			expectedRemoteServer: true,
		},
		{
			name:               "obfuscated group given as the tag over nordwhisper",
			api:                core_test.NewMockFailingServersAPI(errors.New("500")),
			servers:            core_test.ServersList(),
			tech:               config.Technology_NORDWHISPER,
			tag:                "obfuscated",
			expectedServerName: "Germany #4",
		},
		{
			name:               "standard group over nordwhisper keeps working",
			api:                core_test.NewMockFailingServersAPI(errors.New("500")),
			servers:            core_test.ServersList(),
			tech:               config.Technology_NORDWHISPER,
			group:              "standard_vpn_servers",
			expectedServerName: "Germany #4",
		},
		{
			name:          "obfuscated group is not connectable over nordlynx",
			api:           core_test.NewMockFailingServersAPI(errors.New("500")),
			servers:       core_test.ServersList(),
			tech:          config.Technology_NORDLYNX,
			group:         "obfuscated",
			expectedError: internal.ErrServerIsUnavailable,
		},
		{
			name:          "legacy XOR only server is not connectable by name",
			api:           core_test.NewMockFailingServersAPI(errors.New("500")),
			servers:       core_test.ServersList(),
			tech:          config.Technology_OPENVPN,
			tag:           "lt17",
			expectedError: internal.ErrServerIsUnavailable,
		},
		{
			name:          "unknown technology",
			api:           core_test.NewMockServersAPI(),
			servers:       core_test.ServersList(),
			tech:          config.Technology_UNKNOWN_TECHNOLOGY,
			expectedError: ErrUnknownTechnology,
		},
		{
			name:          "group given both as the tag and the flag",
			api:           core_test.NewMockServersAPI(),
			servers:       core_test.ServersList(),
			tech:          config.Technology_NORDLYNX,
			tag:           "onion_over_vpn",
			group:         "double_vpn",
			expectedError: internal.ErrDoubleGroup,
		},
		{
			name:          "group flag does not exist",
			api:           core_test.NewMockServersAPI(),
			servers:       core_test.ServersList(),
			tech:          config.Technology_NORDLYNX,
			group:         "not_a_group",
			expectedError: internal.ErrGroupDoesNotExist,
		},
		{
			name:          "dedicated IP group is handled by the caller",
			api:           core_test.NewMockServersAPI(),
			servers:       core_test.ServersList(),
			tech:          config.Technology_NORDLYNX,
			group:         "dedicated_ip",
			expectedError: ErrDedicatedIPServer,
		},
		{
			name:          "dedicated server group is handled by the caller",
			api:           core_test.NewMockServersAPI(),
			servers:       core_test.ServersList(),
			tech:          config.Technology_NORDLYNX,
			group:         "dedicated_server",
			expectedError: ErrDedicatedServer,
		},
		{
			name:          "tag does not exist",
			api:           core_test.NewMockServersAPI(),
			servers:       core_test.ServersList(),
			tech:          config.Technology_NORDLYNX,
			tag:           "not a tag",
			expectedError: internal.ErrTagDoesNotExist,
		},
		{
			name:          "local servers are all dedicated IP servers",
			api:           core_test.NewMockFailingServersAPI(errors.New("500")),
			servers:       core_test.ServersList("lt7.nordvpn.com", "lt8.nordvpn.com"),
			tech:          config.Technology_NORDLYNX,
			expectedError: internal.ErrServerIsUnavailable,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := config.Config{Technology: test.tech}
			if test.tech == config.Technology_OPENVPN {
				cfg.AutoConnectData.Protocol = config.Protocol_TCP
			}

			serverSelection, err := PickServer(
				test.api,
				test.servers,
				core_test.CountriesList(),
				core.Insights{},
				cfg,
				NewSearchParams(test.tag, test.group, test.excludedServer),
			)

			assert.Equal(t, test.expectedError, err)
			assert.Equal(t, test.expectedRemoteServer, serverSelection.Remote)
			if len(test.expectedServerName) > 0 {
				assert.Equal(t, test.expectedServerName, serverSelection.Server.Name)
			}
		})
	}
}

func TestGetServerParameters(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		name     string
		tag      string
		group    string
		expected ServerParameters
	}{
		{
			name:     "group found for group name",
			group:    "double_vpn",
			tag:      "",
			expected: ServerParameters{Group: config.ServerGroup_DOUBLE_VPN},
		},
		{
			name:     "group name is in tag field",
			group:    "",
			tag:      "double_vpn",
			expected: ServerParameters{Group: config.ServerGroup_DOUBLE_VPN},
		},
		{
			name:     "country name",
			group:    "",
			tag:      "germany",
			expected: ServerParameters{Group: config.ServerGroup_UNDEFINED, Country: "Germany", CountryCode: "DE"},
		},
		{
			name:     "country code",
			group:    "",
			tag:      "De",
			expected: ServerParameters{Group: config.ServerGroup_UNDEFINED, Country: "Germany", CountryCode: "DE"},
		},
		{
			name:     "country code + group",
			group:    "double_vpn",
			tag:      "De",
			expected: ServerParameters{Group: config.ServerGroup_DOUBLE_VPN, Country: "Germany", CountryCode: "DE"},
		},
		{
			name:     "city name",
			group:    "",
			tag:      "berlin",
			expected: ServerParameters{Group: config.ServerGroup_UNDEFINED, Country: "Germany", CountryCode: "DE", City: "Berlin"},
		},
		{
			name:     "country name + city",
			group:    "",
			tag:      "germany berlin",
			expected: ServerParameters{Group: config.ServerGroup_UNDEFINED, Country: "Germany", CountryCode: "DE", City: "Berlin"},
		},
		{
			name:     "country code + city",
			group:    "",
			tag:      "de berlin",
			expected: ServerParameters{Group: config.ServerGroup_UNDEFINED, Country: "Germany", CountryCode: "DE", City: "Berlin"},
		},
		{
			name:     "country code + city + group",
			group:    "double_vpn",
			tag:      "de berlin",
			expected: ServerParameters{Group: config.ServerGroup_DOUBLE_VPN, Country: "Germany", CountryCode: "DE", City: "Berlin"},
		},
		{
			name:     "server name",
			group:    "",
			tag:      "de123",
			expected: ServerParameters{Group: config.ServerGroup_UNDEFINED, ServerName: "de123"},
		},
		{
			name:     "server name + group",
			group:    "double_vpn",
			tag:      "de123",
			expected: ServerParameters{Group: config.ServerGroup_DOUBLE_VPN, ServerName: "de123"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			params := GetServerParameters(test.tag, test.group, core_test.CountriesList())

			assert.Equal(t, test.expected, params)
		})
	}
}

func TestRecommendationUUID_ExtractRecommendationUUID(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name                    string
		headerKey               string
		headerValue             string
		expectedError           bool
		expectedRecommendedUUID RecommendationUUID
	}{
		{
			name:                    "valid uuid",
			headerKey:               recommendationUUIDHeader,
			headerValue:             "c0b4c990-3000-457f-8b81-6850b8cdb54e",
			expectedError:           false,
			expectedRecommendedUUID: RecommendationUUID(core_test.TestRecommendedUUID),
		},
		{
			name:                    "missing uuid",
			headerKey:               "NOT-Recommendation-Uuid",
			headerValue:             "random-value",
			expectedError:           true,
			expectedRecommendedUUID: emptyUUID,
		},
		{
			name:                    "bad uuid",
			headerKey:               recommendationUUIDHeader,
			headerValue:             "not-a-uuid",
			expectedError:           true,
			expectedRecommendedUUID: emptyUUID,
		},
	}

	for _, test := range tests {
		header := http.Header{}
		header.Set(test.headerKey, test.headerValue)

		t.Run(test.name, func(t *testing.T) {
			recommendedUUID, err := extractRecommendationUUID(header)
			assert.Equal(t, test.expectedRecommendedUUID, recommendedUUID)
			if test.expectedError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestRecommendationUUID_GetServersRemote(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name                    string
		api                     core.ServersAPI
		longitude               float64
		latitude                float64
		tech                    config.Technology
		protocol                config.Protocol
		obfuscated              bool
		tag                     core.ServerTag
		group                   config.ServerGroup
		expectedError           error
		expectedRecommendedUUID RecommendationUUID
	}{
		{
			name:                    "recommended uuid",
			api:                     core_test.NewMockServersAPI(),
			tech:                    config.Technology_NORDLYNX,
			expectedError:           nil,
			expectedRecommendedUUID: RecommendationUUID(core_test.TestRecommendedUUID),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filter := core.ServersFilter{
				Group: test.group,
				Tech:  TechToServerTech(test.tech, test.protocol),
				Tag:   test.tag,
				Limit: apiServersLimit,
			}
			_, recommendedUUID, err := getRecommendedServers(
				test.api,
				test.longitude,
				test.latitude,
				filter,
				func(core.Server) bool { return true },
			)
			assert.Equal(t, test.expectedError, err)
			assert.Equal(t, test.expectedRecommendedUUID, recommendedUUID)
		})
	}
}

func TestNewSearchParams(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name           string
		tag            string
		group          string
		excludedServer string
		expected       SearchParams
	}{
		{
			name:           "plain values are kept as is",
			tag:            "it rome",
			group:          "Double_VPN",
			excludedServer: "it1.nordvpn.com",
			expected: SearchParams{
				Tag:            "it rome",
				Group:          "Double_VPN",
				ExcludedServer: "it1.nordvpn.com",
			},
		},
		{
			name:           "leading and trailing whitespace is trimmed from tag and group",
			tag:            "  it rome  ",
			group:          "  Double_VPN ",
			excludedServer: "it1.nordvpn.com",
			expected: SearchParams{
				Tag:            "it rome",
				Group:          "Double_VPN",
				ExcludedServer: "it1.nordvpn.com",
			},
		},
		{
			name:           "whitespace-only tag and group become empty",
			tag:            "   ",
			group:          "\t\n ",
			excludedServer: "",
			expected:       SearchParams{},
		},
		{
			name:           "internal whitespace in tag is preserved",
			tag:            "it rome",
			group:          "",
			excludedServer: "",
			expected: SearchParams{
				Tag: "it rome",
			},
		},
		{
			name:           "excluded server is trimmed",
			tag:            "it",
			group:          "",
			excludedServer: " it1.nordvpn.com ",
			expected: SearchParams{
				Tag:            "it",
				ExcludedServer: "it1.nordvpn.com",
			},
		},
		{
			name:           "all empty",
			tag:            "",
			group:          "",
			excludedServer: "",
			expected:       SearchParams{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, NewSearchParams(tt.tag, tt.group, tt.excludedServer))
		})
	}
}

func TestPickServer_ServerSelectionOrder(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name                       string
		apiServers                 core.Servers
		excludedServer             string
		expectedServerName         string
		expectedRemoteServer       bool
		expectedRecommendationUUID RecommendationUUID
	}{
		{
			name:                       "first server from the API is selected",
			apiServers:                 core_test.ServersList("fr1.nordvpn.com", "de3.nordvpn.com", "lt15.nordvpn.com"),
			expectedServerName:         "France #1",
			expectedRemoteServer:       true,
			expectedRecommendationUUID: RecommendationUUID(core_test.TestRecommendedUUID),
		},
		{
			name:                       "API order is respected",
			apiServers:                 core_test.ServersList("lt15.nordvpn.com", "de3.nordvpn.com", "fr1.nordvpn.com"),
			expectedServerName:         "Lithuania #15",
			expectedRemoteServer:       true,
			expectedRecommendationUUID: RecommendationUUID(core_test.TestRecommendedUUID),
		},
		{
			name:                       "next server from the API is selected when the first one is excluded",
			apiServers:                 core_test.ServersList("fr1.nordvpn.com", "de3.nordvpn.com", "lt15.nordvpn.com"),
			excludedServer:             "fr1.nordvpn.com",
			expectedServerName:         "Germany #3",
			expectedRemoteServer:       true,
			expectedRecommendationUUID: RecommendationUUID(core_test.TestRecommendedUUID),
		},
		{
			name:                       "local servers are used when no API server matches the user settings",
			apiServers:                 core_test.ServersList("de4.nordvpn.com"),
			expectedRemoteServer:       false,
			expectedRecommendationUUID: emptyUUID,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			api := &core_test.MockServersAPI{ServersList: test.apiServers}
			cfg := config.Config{Technology: config.Technology_NORDLYNX}

			serverSelection, err := PickServer(
				api,
				core_test.ServersList(),
				core_test.CountriesList(),
				core.Insights{},
				cfg,
				NewSearchParams("", "", test.excludedServer),
			)

			assert.NoError(t, err)
			assert.Equal(t, test.expectedRemoteServer, serverSelection.Remote)
			assert.Equal(t, test.expectedRecommendationUUID, serverSelection.RecommendationUUID)
			if len(test.expectedServerName) > 0 {
				assert.Equal(t, test.expectedServerName, serverSelection.Server.Name)
			}
		})
	}
}

func TestFindServersLocally_VirtualServers(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name             string
		physicalCount    int
		virtualCount     int
		expectedPhysical int
		expectedVirtual  int
	}{
		{
			name:             "only physical servers",
			physicalCount:    5,
			expectedPhysical: 5,
		},
		{
			name:             "enough physical servers, virtual ones are removed",
			physicalCount:    25,
			virtualCount:     5,
			expectedPhysical: 25,
		},
		{
			name:             "exactly enough physical servers, virtual ones are removed",
			physicalCount:    20,
			virtualCount:     5,
			expectedPhysical: 20,
		},
		{
			name:             "not enough physical servers, filled with virtual ones",
			physicalCount:    15,
			virtualCount:     10,
			expectedPhysical: 15,
			expectedVirtual:  5,
		},
		{
			name:             "not enough physical and virtual servers, all are used",
			physicalCount:    5,
			virtualCount:     3,
			expectedPhysical: 5,
			expectedVirtual:  3,
		},
		{
			name:            "no physical servers, virtual ones are used",
			virtualCount:    3,
			expectedVirtual: 3,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			servers := append(newServers(test.physicalCount, false), newServers(test.virtualCount, true)...)

			selected, err := findServersLocally(servers, core.ServerTag{}, func(core.Server) bool { return true })

			assert.NoError(t, err)
			virtualCount := len(internal.Filter(selected, func(s core.Server) bool { return s.IsVirtualLocation() }))
			assert.Equal(t, test.expectedVirtual, virtualCount)
			assert.Equal(t, test.expectedPhysical, len(selected)-virtualCount)
		})
	}
}

func newServers(count int, virtual bool) core.Servers {
	servers := make(core.Servers, 0, count)
	for i := range count {
		server := core.Server{ID: int64(i), Hostname: fmt.Sprintf("physical%d", i)}
		if virtual {
			server.Hostname = fmt.Sprintf("virtual%d", i)
			server = withVirtualLocation(server)
		}
		servers = append(servers, server)
	}
	return servers
}

func withVirtualLocation(server core.Server) core.Server {
	server.Specifications = append(slices.Clone(server.Specifications), core.Specification{
		Identifier: core.VirtualLocation,
		Values: []struct {
			Value string `json:"value"`
		}{{Value: "true"}},
	})
	return server
}

func TestFindServersLocally_DedicatedIPServers(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name              string
		servers           core.Servers
		serverTag         core.ServerTag
		expectedHostnames []string
	}{
		{
			name:              "dedicated IP servers are removed",
			servers:           core_test.ServersList("lt15.nordvpn.com", "lt8.nordvpn.com"),
			serverTag:         core.ServerTag{Action: core.ServerByCity},
			expectedHostnames: []string{"lt15.nordvpn.com"},
		},
		{
			name:              "dedicated IP servers are kept when searching by server name",
			servers:           core_test.ServersList("lt15.nordvpn.com", "lt8.nordvpn.com"),
			serverTag:         core.ServerTag{Action: core.ServerByName},
			expectedHostnames: []string{"lt15.nordvpn.com", "lt8.nordvpn.com"},
		},
		{
			name:              "single dedicated IP server is kept",
			servers:           core_test.ServersList("lt8.nordvpn.com"),
			serverTag:         core.ServerTag{Action: core.ServerByCity},
			expectedHostnames: []string{"lt8.nordvpn.com"},
		},
		{
			name:              "only dedicated IP servers results in an empty list",
			servers:           core_test.ServersList("lt7.nordvpn.com", "lt8.nordvpn.com"),
			serverTag:         core.ServerTag{Action: core.ServerByCity},
			expectedHostnames: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			selected, err := findServersLocally(test.servers, test.serverTag, func(core.Server) bool { return true })

			assert.NoError(t, err)
			hostnames := []string{}
			for _, server := range selected {
				hostnames = append(hostnames, server.Hostname)
			}
			assert.Equal(t, test.expectedHostnames, hostnames)
		})
	}
}

func TestPickServer_Fallback(t *testing.T) {
	category.Set(t, category.Unit)

	// a server with a different technology from the API than in the local cache
	apiDe3 := core_test.ServersList("de3.nordvpn.com")[0]
	apiDe3.Technologies = core_test.ServersList("de4.nordvpn.com")[0].Technologies

	// a server without keys matches every local search
	keylessFr1 := core_test.ServersList("fr1.nordvpn.com")[0]
	keylessFr1.Keys = nil

	tests := []struct {
		name                       string
		api                        *core_test.MockServersAPI
		servers                    core.Servers
		tag                        string
		expectedError              error
		expectedServerName         string
		expectedRemoteServer       bool
		expectedRecommendationUUID RecommendationUUID
		expectedRecommendedCalls   int
		expectedServerCalled       bool
	}{
		{
			name:                     "API error is not returned when local servers are found",
			api:                      core_test.NewMockFailingServersAPI(core.ErrUnauthorized),
			servers:                  core_test.ServersList("fr1.nordvpn.com"),
			expectedServerName:       "France #1",
			expectedRecommendedCalls: 1,
		},
		{
			name:                     "local error is returned when both API and local servers fail",
			api:                      core_test.NewMockFailingServersAPI(core.ErrUnauthorized),
			servers:                  core.Servers{},
			expectedError:            internal.ErrServerIsUnavailable,
			expectedRecommendedCalls: 1,
		},
		{
			name: "server by name from the API not matching the settings falls back to local servers",
			api: &core_test.MockServersAPI{
				ServersList: core.Servers{apiDe3},
			},
			servers:              core_test.ServersList(),
			tag:                  "de3",
			expectedServerName:   "Germany #3",
			expectedServerCalled: true,
		},
		{
			name: "API servers are used without the recommendation UUID header",
			api: &core_test.MockServersAPI{
				ServersList:            core_test.ServersList("fr1.nordvpn.com"),
				DropRecommendationUUID: true,
			},
			servers:                  core_test.ServersList(),
			expectedServerName:       "France #1",
			expectedRemoteServer:     true,
			expectedRecommendedCalls: 1,
		},
		{
			name:                     "unrecognized tag skips the API and uses local servers",
			api:                      core_test.NewMockServersAPI(),
			servers:                  core.Servers{keylessFr1},
			tag:                      "xx99",
			expectedServerName:       "France #1",
			expectedRecommendedCalls: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := config.Config{Technology: config.Technology_NORDLYNX}

			serverSelection, err := PickServer(
				test.api,
				test.servers,
				core_test.CountriesList(),
				core.Insights{},
				cfg,
				NewSearchParams(test.tag, "", ""),
			)

			assert.Equal(t, test.expectedError, err)
			assert.Equal(t, test.expectedRemoteServer, serverSelection.Remote)
			assert.Equal(t, test.expectedRecommendationUUID, serverSelection.RecommendationUUID)
			assert.Equal(t, test.expectedRecommendedCalls, test.api.RecommendedCalls)
			assert.Equal(t, test.expectedServerCalled, test.api.ServerEndpointCalled)
			if test.expectedError != nil {
				return
			}
			if len(test.expectedServerName) > 0 {
				assert.Equal(t, test.expectedServerName, serverSelection.Server.Name)
			}
		})
	}
}
