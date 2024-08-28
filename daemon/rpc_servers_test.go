package daemon

import (
	"context"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestServers(t *testing.T) {
	category.Set(t, category.Unit)

	server1Hostname := "server1"
	server1Country := "Germany"
	server1CountryCode := "de"
	server1City := "Berlin"

	server2Hostname := "server2"
	server2Country := "France"
	server2CountryCode := "fr"
	server2City := "Paris"

	server3Hostname := "server3"
	server3Country := "Lithuania"
	server3CountryCode := "lt"
	server3City := "Vilnius"

	servers := core.Servers{
		// Server 1
		// Virtual/P2P/Germany/Berlin/Nordlynx/OpenVPN TCP Obfuscated
		{
			Hostname: server1Hostname,
			Locations: core.Locations{
				{
					Country: core.Country{
						Name: server1Country,
						Code: server1CountryCode,
						City: core.City{
							Name: server1City,
						},
					}},
			},
			Specifications: []core.Specification{
				{
					Identifier: core.VirtualLocation,
					Values: []struct {
						Value string "json:\"value\""
					}{
						{Value: "True"},
					},
				},
			},
			Groups: core.Groups{
				{
					ID:    config.ServerGroup_P2P,
					Title: "P2P",
				},
				{
					ID:    config.ServerGroup_NETFLIX_USA,
					Title: "Netflix USA",
				},
			},
			Technologies: core.Technologies{
				{ID: core.L2TP},
				{ID: core.HTTPProxy},
				{ID: core.WireguardTech},
				{ID: core.OpenVPNTCPObfuscated},
			},
		},
		// Server 2
		// France/Paris/Standard VPN/OpenVPN TCP/OpenVPN UDP
		{
			Hostname: server2Hostname,
			Locations: core.Locations{
				{
					Country: core.Country{
						Name: server2Country,
						Code: server2CountryCode,
						City: core.City{
							Name: server2City,
						},
					}},
			},
			Specifications: []core.Specification{},
			Groups: core.Groups{
				{
					ID:    config.ServerGroup_STANDARD_VPN_SERVERS,
					Title: "Standard VPN",
				},
			},
			Technologies: core.Technologies{
				{ID: core.L2TP},
				{ID: core.HTTPProxy},
				{ID: core.OpenVPNTCP},
				{ID: core.OpenVPNUDP},
			},
		},
		// Server 3
		// Lithuania/Vilnius/Standard VPN/OpenVPN TCP Obfuscated
		{
			Hostname: server3Hostname,
			Locations: core.Locations{
				{
					Country: core.Country{
						Name: server3Country,
						Code: server3CountryCode,
						City: core.City{
							Name: server3City,
						},
					}},
			},
			Specifications: []core.Specification{},
			Groups: core.Groups{
				{
					ID:    config.ServerGroup_OBFUSCATED,
					Title: "Standard VPN",
				},
				{
					ID:    config.ServerGroup_ANTI_DDOS,
					Title: "Anti DDOS",
				},
				{
					ID:    config.ServerGroup_EUROPE,
					Title: "Europe",
				},
			},
			Technologies: core.Technologies{
				{ID: core.L2TP},
				{ID: core.HTTPProxy},
				{ID: core.OpenVPNTCPObfuscated},
				{ID: core.OpenVPNTCP},
				{ID: core.WireguardTech},
			},
		},
	}

	expectedServers := []*pb.Server{
		{
			CountryCode: server1CountryCode,
			CityName:    server1City,
			HostName:    server1Hostname,
			Virtual:     true,
			ServerGroup: config.ServerGroup_P2P,
			Technologies: []pb.Technology{
				pb.Technology_NORDLYNX,
				pb.Technology_OBFUSCATED_OPENVPN_TCP,
			},
		},
		{
			CountryCode: server2CountryCode,
			CityName:    server2City,
			HostName:    server2Hostname,
			Virtual:     false,
			ServerGroup: config.ServerGroup_STANDARD_VPN_SERVERS,
			Technologies: []pb.Technology{
				pb.Technology_OPENVPN_TCP,
				pb.Technology_OPENVPN_UDP,
			},
		},
		{
			CountryCode: server3CountryCode,
			CityName:    server3City,
			HostName:    server3Hostname,
			Virtual:     false,
			ServerGroup: config.ServerGroup_OBFUSCATED,
			Technologies: []pb.Technology{
				pb.Technology_OBFUSCATED_OPENVPN_TCP,
				pb.Technology_OPENVPN_TCP,
				pb.Technology_NORDLYNX,
			},
		},
	}

	tests := []struct {
		name             string
		serversList      core.Servers
		serversErr       error
		expectedResponse *pb.ServersResponse
	}{
		{
			name:        "success",
			serversList: servers,
			expectedResponse: &pb.ServersResponse{
				Servers: expectedServers,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dm := DataManager{}
			dm.serversData.Servers = servers
			r := RPC{dm: &dm}

			resp, err := r.GetServers(context.Background(), &pb.Empty{})
			assert.Nil(t, err, "Unexpeced error returned by servers RPC.")
			assert.Equal(t, test.expectedResponse, resp)
		})
	}
}
