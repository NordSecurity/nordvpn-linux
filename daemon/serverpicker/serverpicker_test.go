package serverpicker

import (
	"errors"
	"net/http"
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
			localSelFn := selectFilterForLocalServers("", test.group, false)
			filterFn := func(s core.Server) bool {
				return core.IsConnectableWithProtocol(test.tech, test.proto)(s) &&
					!core.IsObfuscated()(s) &&
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
			input:         NewSearchParams("", "p2p", ""),
			expectedGroup: config.ServerGroup_P2P,
			err:           nil,
		},
		{
			input:         NewSearchParams("p2p", "", ""),
			tagChanged:    true,
			expectedGroup: config.ServerGroup_P2P,
			err:           nil,
		},
		{
			input:         NewSearchParams("p2p", "p2p", ""),
			expectedGroup: config.ServerGroup_UNDEFINED,
			err:           internal.ErrDoubleGroup,
		},
		{
			input:         NewSearchParams("p2p", "quantum_vpn", ""),
			expectedGroup: config.ServerGroup_UNDEFINED,
			err:           internal.ErrGroupDoesNotExist,
		},
		{
			input:         NewSearchParams("quantum_vpn", "p2p", ""),
			expectedGroup: config.ServerGroup_P2P,
			err:           nil,
		},
		{
			input:         NewSearchParams("quantum_vpn", "", ""),
			expectedGroup: config.ServerGroup_UNDEFINED,
			err:           nil,
		},
		{
			input:         NewSearchParams("p2p us1234", "", ""),
			expectedGroup: config.ServerGroup_UNDEFINED,
			err:           nil,
		},
		{
			input:         NewSearchParams("quantum_vpn", "quantum_vpn", ""),
			expectedGroup: config.ServerGroup_UNDEFINED,
			err:           internal.ErrGroupDoesNotExist,
		},
	}

	for _, tt := range tests {
		tag := tt.input.Tag
		group, err := resolveServerGroup(&tt.input, false)
		assert.Equal(t, tt.expectedGroup, group)
		assert.Equal(t, tt.err, err)
		assert.Equal(t, tt.tagChanged, tt.input.Tag != tag)
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
			config.ServerGroup_P2P,
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
			"neflix & chill",
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
			group:    config.ServerGroup_P2P,
			expected: core.ServerTag{Action: core.ServerBySpeed, ID: int64(config.ServerGroup_P2P)},
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
			group:    config.ServerGroup_P2P,
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
			got := TechToServerTech(test.tech, test.protocol, test.obfuscated)
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
		obfuscated           bool
		tag                  string
		onlyPhysicServers    bool
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
			name:              "find server when virtual locations are disabled",
			api:               core_test.NewMockFailingServersAPI(errors.New("500")),
			servers:           core_test.ServersList(),
			tech:              config.Technology_NORDLYNX,
			onlyPhysicServers: true,
		},
		{
			name:              "virtual location disabled returns error when only virtual servers match",
			api:               core_test.NewMockFailingServersAPI(errors.New("500")),
			servers:           core_test.ServersList(),
			tech:              config.Technology_NORDLYNX,
			tag:               "algeria",
			onlyPhysicServers: true,
			expectedError:     internal.ErrVirtualServerSelected,
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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := config.Config{
				Technology: test.tech,
				AutoConnectData: config.AutoConnectData{
					Obfuscate: test.obfuscated,
				},
			}
			if test.onlyPhysicServers {
				cfg.VirtualLocation.Set(false)
			}

			serverSelection, err := PickServer(
				test.api,
				test.servers,
				core_test.CountriesList(),
				core.Insights{},
				cfg,
				NewSearchParams(test.tag, "", test.excludedServer),
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
			group:    "p2p",
			tag:      "",
			expected: ServerParameters{Group: config.ServerGroup_P2P},
		},
		{
			name:     "group name is in tag field",
			group:    "",
			tag:      "p2p",
			expected: ServerParameters{Group: config.ServerGroup_P2P},
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
			group:    "p2p",
			tag:      "De",
			expected: ServerParameters{Group: config.ServerGroup_P2P, Country: "Germany", CountryCode: "DE"},
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
			group:    "p2p",
			tag:      "de berlin",
			expected: ServerParameters{Group: config.ServerGroup_P2P, Country: "Germany", CountryCode: "DE", City: "Berlin"},
		},
		{
			name:     "server name",
			group:    "",
			tag:      "de123",
			expected: ServerParameters{Group: config.ServerGroup_UNDEFINED, ServerName: "de123"},
		},
		{
			name:     "server name + group",
			group:    "p2p",
			tag:      "de123",
			expected: ServerParameters{Group: config.ServerGroup_P2P, ServerName: "de123"},
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
				Tech:  TechToServerTech(test.tech, test.protocol, test.obfuscated),
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
