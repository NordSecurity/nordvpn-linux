package core

import (
	"encoding/json"
	"net/netip"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"golang.org/x/exp/slices"

	"github.com/stretchr/testify/assert"
)

const inputTest = `{
  "id": 956504,
  "created_at": "2020-02-26 13:19:09",
  "updated_at": "2020-06-12 09:56:27",
  "name": "Latvia #40",
  "station": "185.176.222.52",
  "hostname": "lv40.nordvpn.com",
  "load": 10,
  "status": "online",
  "locations": [
    {
      "id": 237,
      "created_at": "2017-06-15 14:06:47",
      "updated_at": "2017-06-15 14:06:47",
      "latitude": 56.95,
      "longitude": 24.1,
      "country": {
        "id": 119,
        "name": "Latvia",
        "code": "LV",
        "city": {
          "id": 5192828,
          "name": "Riga",
          "latitude": 56.95,
          "longitude": 24.1,
          "dns_name": "riga",
          "hub_score": 0
        }
      }
    }
  ],
  "services": [
    {
      "id": 1,
      "name": "VPN",
      "identifier": "vpn",
      "created_at": "2017-03-21 12:00:45",
      "updated_at": "2017-05-25 13:12:31"
    },
    {
      "id": 5,
      "name": "Proxy",
      "identifier": "proxy",
      "created_at": "2017-05-29 19:38:30",
      "updated_at": "2017-05-29 19:38:30"
    }
  ],
  "technologies": [
    {
      "id": 1,
      "name": "IKEv2/IPSec",
      "identifier": "ikev2",
      "created_at": "2017-03-21 12:00:24",
      "updated_at": "2017-09-05 14:20:16",
      "metadata": [],
      "pivot": {
        "technology_id": 1,
        "server_id": 956504,
        "status": "online"
      }
    },
    {
      "id": 3,
      "name": "OpenVPN UDP",
      "identifier": "openvpn_udp",
      "created_at": "2017-05-04 08:03:24",
      "updated_at": "2017-05-09 19:27:37",
      "metadata": [
        {
		  "name": "ports",
		  "value": [1111, 2222, 3333]
		}
      ],
      "pivot": {
        "technology_id": 3,
        "server_id": 956504,
        "status": "online"
      }
    },
    {
      "id": 5,
      "name": "OpenVPN TCP",
      "identifier": "openvpn_tcp",
      "created_at": "2017-05-09 19:28:14",
      "updated_at": "2017-05-09 19:28:14",
      "metadata": [],
      "pivot": {
        "technology_id": 5,
        "server_id": 956504,
        "status": "online"
      }
    },
    {
      "id": 9,
      "name": "HTTP Proxy",
      "identifier": "proxy",
      "created_at": "2017-05-09 19:29:09",
      "updated_at": "2017-06-13 14:25:29",
      "metadata": [],
      "pivot": {
        "technology_id": 9,
        "server_id": 956504,
        "status": "maintenance"
      }
    },
    {
      "id": 19,
      "name": "HTTP CyberSec Proxy",
      "identifier": "proxy_cybersec",
      "created_at": "2017-08-22 12:44:49",
      "updated_at": "2017-08-22 12:44:49",
      "metadata": [],
      "pivot": {
        "technology_id": 19,
        "server_id": 956504,
        "status": "maintenance"
      }
    },
    {
      "id": 21,
      "name": "HTTP Proxy (SSL)",
      "identifier": "proxy_ssl",
      "created_at": "2017-10-02 12:45:14",
      "updated_at": "2017-10-02 12:45:14",
      "metadata": [],
      "pivot": {
        "technology_id": 21,
        "server_id": 956504,
        "status": "online"
      }
    },
    {
      "id": 23,
      "name": "HTTP CyberSec Proxy (SSL)",
      "identifier": "proxy_ssl_cybersec",
      "created_at": "2017-10-02 12:50:49",
      "updated_at": "2017-10-02 12:50:49",
      "metadata": [],
      "pivot": {
        "technology_id": 23,
        "server_id": 956504,
        "status": "online"
      }
    },
    {
      "id": 35,
      "name": "Wireguard",
      "identifier": "wireguard_udp",
      "created_at": "2019-02-14 14:08:43",
      "updated_at": "2019-02-14 14:08:43",
      "metadata": [
        {
          "name": "public_key",
          "value": "\tZid1YfpCDPeeyWzEEmiZLPcmwNopke/B/Pa/DtiViiw="
        }
      ],
      "pivot": {
        "technology_id": 35,
        "server_id": 956504,
        "status": "online"
      }
    },
	{
      "id": 51,
      "name": "NordWhisper",
      "identifier": "quench",
      "created_at": "2024-12-1 15:08:23",
      "updated_at": "2024-12-1 15:08:23",
      "metadata": [
        {
          "name": "port",
          "value": "12345"
        }
      ],
      "pivot": {
        "technology_id": 51,
        "server_id": 956504,
        "status": "online"
      }
    }
  ],
  "groups": [
    {
      "id": 11,
      "created_at": "2017-06-13 13:43:00",
      "updated_at": "2017-06-13 13:43:00",
      "title": "Standard VPN servers",
      "type": {
        "id": 3,
        "created_at": "2017-06-13 13:40:17",
        "updated_at": "2017-06-13 13:40:23",
        "title": "Legacy category",
        "identifier": "legacy_group_category"
      }
    },
    {
      "id": 15,
      "created_at": "2017-06-13 13:43:38",
      "updated_at": "2017-06-13 13:43:38",
      "title": "P2P",
      "type": {
        "id": 3,
        "created_at": "2017-06-13 13:40:17",
        "updated_at": "2017-06-13 13:40:23",
        "title": "Legacy category",
        "identifier": "legacy_group_category"
      }
    },
    {
      "id": 19,
      "created_at": "2017-10-27 14:17:17",
      "updated_at": "2017-10-27 14:17:17",
      "title": "Europe",
      "type": {
        "id": 5,
        "created_at": "2017-10-27 14:16:30",
        "updated_at": "2017-10-27 14:16:30",
        "title": "Regions",
        "identifier": "regions"
      }
    }
  ],
  "specifications": [
    {
      "id": 8,
      "title": "Version",
      "identifier": "version",
      "values": [
        {
          "id": 257,
          "value": "2.1.0"
        }
      ]
    }
  ],
  "ips": [
    {
      "id": 180629,
      "created_at": "2020-02-26 13:19:09",
      "updated_at": "2020-02-26 13:19:09",
      "server_id": 956504,
      "ip_id": 102536,
      "type": "entry",
      "ip": {
        "id": 102536,
        "ip": "185.176.222.52",
        "version": 4
      }
    },
    {
      "id": 180630,
      "created_at": "2020-02-26 13:19:09",
      "updated_at": "2020-02-26 13:19:09",
      "server_id": 956504,
      "ip_id": 102537,
      "type": "entry",
      "ip": {
        "id": 102536,
        "ip": "::1",
        "version": 6
      }
    },
    {
      "id": 180631,
      "created_at": "2020-02-26 13:19:09",
      "updated_at": "2020-02-26 13:19:09",
      "server_id": 956504,
      "ip_id": 102537,
      "type": "entry",
      "ip": {
        "id": 102536,
        "ip": "2a02:5740:1:9::11",
        "version": 6
      }
    },
    {
      "id": 180632,
      "created_at": "2020-02-26 13:19:09",
      "updated_at": "2020-02-26 13:19:09",
      "server_id": 956504,
      "ip_id": 102538,
      "type": "entry",
      "ip": {
        "id": 102536,
        "ip": "196.245.151.3",
        "version": 4
      }
    }
  ]
}`

func TestIsOnline(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		server   Server
		expected bool
	}{
		{
			name:     "online server",
			server:   Server{Status: Online},
			expected: true,
		},
		{
			name:     "offline server",
			server:   Server{Status: Offline},
			expected: false,
		},
		{
			name:     "maintenance server",
			server:   Server{Status: Maintenance},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, IsOnline()(test.server))
		})
	}
}

func TestIsObfuscated(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		server   Server
		expected bool
	}{
		{
			name: "openvpn both obfuscated technologies online and server online",
			server: Server{
				Status: Online,
				Technologies: Technologies{
					Technology{
						ID:    OpenVPNUDPObfuscated,
						Pivot: Pivot{Status: Online},
					},
					Technology{
						ID:    OpenVPNTCPObfuscated,
						Pivot: Pivot{Status: Online},
					},
				},
			},
			expected: true,
		},
		{
			name: "openvpn both obfuscated technologies online but server offline",
			server: Server{
				Status: Offline,
				Technologies: Technologies{
					Technology{
						ID:    OpenVPNUDPObfuscated,
						Pivot: Pivot{Status: Online},
					},
					Technology{
						ID:    OpenVPNTCPObfuscated,
						Pivot: Pivot{Status: Online},
					},
				},
			},
			expected: false,
		},
		{
			name: "openvpn one obfuscate technology online and server online",
			server: Server{
				Status: Online,
				Technologies: Technologies{
					Technology{
						ID:    OpenVPNUDPObfuscated,
						Pivot: Pivot{Status: Online},
					},
					Technology{
						ID:    OpenVPNTCPObfuscated,
						Pivot: Pivot{Status: Offline},
					},
				},
			},
			expected: true,
		},
		{
			name: "not obfuscated online technology with online server",
			server: Server{
				Status: Online,
				Technologies: Technologies{
					Technology{
						ID:    WireguardTech,
						Pivot: Pivot{Status: Online},
					},
				},
			},
			expected: false,
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, IsObfuscated()(test.server))
	}
}

func TestIsConnectableVia(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		server   Server
		tech     ServerTechnology
		expected bool
	}{
		{
			name: "both server and technology are online",
			server: Server{
				Status: Online,
				Technologies: Technologies{
					Technology{
						ID:    WireguardTech,
						Pivot: Pivot{Status: Online},
					},
				},
			},
			tech:     WireguardTech,
			expected: true,
		},
		{
			name: "different technology",
			server: Server{
				Status: Online,
				Technologies: Technologies{
					Technology{
						ID:    WireguardTech,
						Pivot: Pivot{Status: Online},
					},
				},
			},
			tech:     OpenVPNUDP,
			expected: false,
		},
		{
			name: "only technology is online",
			server: Server{
				Status: Offline,
				Technologies: Technologies{
					Technology{
						ID:    WireguardTech,
						Pivot: Pivot{Status: Online},
					},
				},
			},
			tech:     OpenVPNUDP,
			expected: false,
		},
		{
			name: "only server is online",
			server: Server{
				Status: Online,
				Technologies: Technologies{
					Technology{
						ID:    WireguardTech,
						Pivot: Pivot{Status: Offline},
					},
				},
			},
			tech:     WireguardTech,
			expected: false,
		},
		{
			name: "both server and technology are offline",
			server: Server{
				Status: Offline,
				Technologies: Technologies{
					Technology{
						ID:    WireguardTech,
						Pivot: Pivot{Status: Offline},
					},
				},
			},
			tech:     WireguardTech,
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, IsConnectableVia(test.tech)(test.server))
		})
	}
}

func TestIsConnectableWithProtocol(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		server   Server
		tech     config.Technology
		proto    config.Protocol
		expected bool
	}{
		{
			name: "openvpn udp matching tech and proto",
			server: Server{
				Status: Online,
				Technologies: Technologies{{
					ID:    OpenVPNUDP,
					Pivot: Pivot{Status: Online},
				}},
			},
			tech:     config.Technology_OPENVPN,
			proto:    config.Protocol_UDP,
			expected: true,
		},
		{
			name: "openvpn udp matching tech only",
			server: Server{
				Status: Online,
				Technologies: Technologies{{
					ID:    OpenVPNUDP,
					Pivot: Pivot{Status: Online},
				}},
			},
			tech:     config.Technology_OPENVPN,
			proto:    config.Protocol_TCP,
			expected: false,
		},
		{
			name: "openvpn udp matching proto only",
			server: Server{
				Status: Online,
				Technologies: Technologies{{
					ID:    OpenVPNUDP,
					Pivot: Pivot{Status: Online},
				}},
			},
			tech:     config.Technology_NORDLYNX,
			proto:    config.Protocol_UDP,
			expected: false,
		},
		{
			name: "nordlynx ignores protocol",
			server: Server{
				Status: Online,
				Technologies: Technologies{{
					ID:    WireguardTech,
					Pivot: Pivot{Status: Online},
				}},
			},
			tech:     config.Technology_NORDLYNX,
			proto:    config.Protocol_TCP,
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected,
				IsConnectableWithProtocol(test.tech, test.proto)(test.server),
			)
		})
	}
}

func TestByTag(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		tag      string
		group    Group
		expected bool
	}{
		{
			name:     "has tag",
			group:    Group{Title: "Double VPN"},
			tag:      "double_vpn",
			expected: true,
		},
		{
			name:     "doesn't have tag",
			group:    Group{Title: "Double VPN"},
			tag:      "triple_vpn",
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, ByTag(test.tag)(test.group))
		})
	}
}

func TestByGroup(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		group    Group
		sg       config.ServerGroup
		expected bool
	}{
		{
			name:     "has group",
			group:    Group{ID: config.ServerGroup_DOUBLE_VPN},
			sg:       config.ServerGroup_DOUBLE_VPN,
			expected: true,
		},
		{
			name:     "doesn't have group",
			group:    Group{ID: config.ServerGroup_DOUBLE_VPN},
			sg:       config.ServerGroup_ONION_OVER_VPN,
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, ByGroup(test.sg)(test.group))
		})
	}
}

func TestServerUnmarshal(t *testing.T) {
	category.Set(t, category.Unit)

	var server Server
	err := json.Unmarshal([]byte(inputTest), &server)
	assert.NoError(t, err)
	assert.False(t, strings.HasPrefix(server.NordLynxPublicKey, "\t"))
}

func TestServerGroupsString(t *testing.T) {
	category.Set(t, category.Unit)

	var server Server
	err := json.Unmarshal([]byte(inputTest), &server)
	assert.NoError(t, err)
	groupIDs := []int64{}
	for _, g := range server.Groups {
		groupIDs = append(groupIDs, int64(g.ID))
	}
	assert.True(t, slices.Equal([]int64{11, 15, 19}, groupIDs))
}

func TestServer_SupportsIPv6(t *testing.T) {
	category.Set(t, category.Unit)

	var server Server
	err := json.Unmarshal([]byte(inputTest), &server)
	assert.NoError(t, err)
	assert.True(t, server.SupportsIPv6())
}

func TestServer_IPs(t *testing.T) {
	category.Set(t, category.Unit)

	var server Server
	err := json.Unmarshal([]byte(inputTest), &server)
	assert.NoError(t, err)
	assert.Equal(
		t,
		[]netip.Addr{
			netip.MustParseAddr("185.176.222.52"),
			netip.MustParseAddr("::1"),
			netip.MustParseAddr("2a02:5740:1:9::11"),
			netip.MustParseAddr("196.245.151.3"),
		},
		server.IPs(),
	)
}

func TestServer_IPv4(t *testing.T) {
	category.Set(t, category.Unit)

	var server Server
	err := json.Unmarshal([]byte(inputTest), &server)
	assert.NoError(t, err)
	got, err := server.IPv4()
	assert.NoError(t, err)
	assert.Equal(t, netip.MustParseAddr("185.176.222.52"), got)
}

func TestServerVersion(t *testing.T) {
	category.Set(t, category.Unit)

	var server Server
	err := json.Unmarshal([]byte(inputTest), &server)
	assert.NoError(t, err)
	assert.Equal(t, "2.1.0", server.Version())
}

func TestLocationsCountry(t *testing.T) {
	category.Set(t, category.Unit)

	t.Run("empty slice", func(t *testing.T) {
		locations := Locations{}
		_, err := locations.Country()
		assert.Error(t, err)
	})

	t.Run("nil slice", func(t *testing.T) {
		var locations Locations = nil
		_, err := locations.Country()
		assert.Error(t, err)
	})

	t.Run("single element", func(t *testing.T) {
		country := Country{}
		locations := Locations{{country}}
		cntr, err := locations.Country()
		assert.NoError(t, err)
		assert.EqualValues(t, country, cntr)
	})

	t.Run("multiple elements", func(t *testing.T) {
		first := Country{Name: "First"}
		second := Country{Name: "Second"}
		third := Country{Name: "Third"}
		locations := Locations{
			{first},
			{second},
			{third},
		}
		country, err := locations.Country()
		assert.NoError(t, err)
		assert.EqualValues(t, first, country)
	})
}

func TestPayment_UnmarshalJSON(t *testing.T) {
	category.Set(t, category.Unit)
	for _, tt := range []struct {
		name    string
		json    string
		payment Payment
		errType error
	}{
		{
			name: "valid payment",
			json: `{"created_at": "2001-01-01 00:00:00", "amount": "1.23", "status": "done"}`,
			payment: Payment{
				CreatedAt: time.Date(2001, time.January, 1, 0, 0, 0, 0, time.UTC),
				Amount:    1.23,
				Status:    "done",
			},
		},
		{
			name:    "invalid JSON",
			errType: &json.SyntaxError{},
		},
		{
			name:    "invalid created_at",
			errType: &time.ParseError{},
			json:    `{"created_at": "2001-01-01"}`,
		},
		{
			name:    "invalid amount",
			errType: strconv.ErrSyntax,
			json:    `{"created_at": "2001-01-01 00:00:00", "amount": "1.2.3"}`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var p Payment
			err := p.UnmarshalJSON([]byte(tt.json))
			if tt.errType != nil {
				target := reflect.New(reflect.TypeOf(tt.errType)).Interface()
				assert.ErrorAs(t, err, target)
			}
			assert.Equal(t, tt.payment, p)
		})
	}
}

func TestNordWhisperPort_UnmarshalJSON(t *testing.T) {
	const serverJson = `{
	"id": 956504,
	"created_at": "2020-02-26 13:19:09",
	"updated_at": "2020-06-12 09:56:27",
	"name": "Latvia #40",
	"station": "185.176.222.52",
	"hostname": "lv40.nordvpn.com",
	"load": 10,
	"status": "online",
	"locations": [
	],
	"services": [
	],
	"groups": [
	],
	"specifications": [
	],
	"ips": [
	],`

	quenchTechJson := `"technologies": [
		{
		"id": 51,
		"name": "NordWhisper",
		"identifier": "quench",
		"created_at": "2024-12-1 15:08:23",
		"updated_at": "2024-12-1 15:08:23",
		"metadata": [
			{
			"name": "port",
			"value": "12345"
			}
		],
		"pivot": {
			"technology_id": 51,
			"server_id": 956504,
			"status": "online"
		}
		}
	]}`

	quenchTechNoPortJson := `"technologies": [
		{
		"id": 51,
		"name": "NordWhisper",
		"identifier": "quench",
		"created_at": "2024-12-1 15:08:23",
		"updated_at": "2024-12-1 15:08:23",
		"metadata": [
		],
		"pivot": {
			"technology_id": 51,
			"server_id": 956504,
			"status": "online"
		}
		}
	]}`

	quenchInvalidPortJson := `"technologies": [
		{
		"id": 51,
		"name": "NordWhisper",
		"identifier": "quench",
		"created_at": "2024-12-1 15:08:23",
		"updated_at": "2024-12-1 15:08:23",
		"metadata": [
			{
			"name": "port",
			"value": "abcd"
			}
		],
		"pivot": {
			"technology_id": 51,
			"server_id": 956504,
			"status": "online"
		}
		}
	]}`

	noQuenchJson := `"technologies": []}`

	multipleTechnologies := `"technologies": [
		{
		"id": 51,
		"name": "NordWhisper",
		"identifier": "quench",
		"created_at": "2024-12-1 15:08:23",
		"updated_at": "2024-12-1 15:08:23",
		"metadata": [
			{
			"name": "port",
			"value": "12345"
			}
		],
		"pivot": {
			"technology_id": 51,
			"server_id": 956504,
			"status": "online"
		}
		},
		{
		"id": 3,
		"name": "OpenVPN UDP",
		"identifier": "openvpn_udp",
		"created_at": "2017-05-04 08:03:24",
		"updated_at": "2017-05-09 19:27:37",
		"metadata": [
			{
			"name": "ports",
			"value": [1111, 2222, 3333]
			}
		],
		"pivot": {
			"technology_id": 3,
			"server_id": 956504,
			"status": "online"
		}
		},
		{
		"id": 5,
		"name": "OpenVPN TCP",
		"identifier": "openvpn_tcp",
		"created_at": "2017-05-09 19:28:14",
		"updated_at": "2017-05-09 19:28:14",
		"metadata": [],
		"pivot": {
			"technology_id": 5,
			"server_id": 956504,
			"status": "online"
		}
		},
		{
		"id": 9,
		"name": "HTTP Proxy",
		"identifier": "proxy",
		"created_at": "2017-05-09 19:29:09",
		"updated_at": "2017-06-13 14:25:29",
		"metadata": [],
		"pivot": {
			"technology_id": 9,
			"server_id": 956504,
			"status": "maintenance"
		}
		}
	]}`

	tests := []struct {
		name           string
		technologyJson string
		expectedPort   int64
	}{
		{
			name:           "success",
			technologyJson: quenchTechJson,
			expectedPort:   12345,
		},
		{
			name:           "quench tech no port json",
			technologyJson: quenchTechNoPortJson,
			expectedPort:   0,
		},
		{
			name:           "invalid port",
			technologyJson: quenchInvalidPortJson,
			expectedPort:   0,
		},
		{
			name:           "no quench technology",
			technologyJson: noQuenchJson,
			expectedPort:   0,
		},
		{
			name:           "multiple technologies",
			technologyJson: multipleTechnologies,
			expectedPort:   12345,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var server Server
			err := server.UnmarshalJSON([]byte(serverJson + test.technologyJson))
			assert.Nil(t, err, "Unexpected error when deserializing server json.")
			assert.Equal(t, test.expectedPort, server.NordWhisperPort)
		})
	}
}

func TestNewCountryCode_SetsCoutryCodeToLowercase(t *testing.T) {
	category.Set(t, category.Unit)

	codes := []string{"US", "Us", "uS"}
	for _, codeStr := range codes {
		t.Run(codeStr, func(t *testing.T) {
			cc := NewCountryCode(codeStr)
			assert.Equal(t, cc.cc, "us")
		})
	}
}
