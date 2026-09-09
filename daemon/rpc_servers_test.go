package daemon

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	core_test "github.com/NordSecurity/nordvpn-linux/test/mock/core"
	"github.com/stretchr/testify/assert"
)

func addToServersMap(serversMap []*pb.ServerCountry,
	countryCode string,
	countryName string,
	city string,
	server *pb.Server) []*pb.ServerCountry {
	for _, serverCountry := range serversMap {
		if serverCountry.CountryCode == countryCode {
			for _, serverCity := range serverCountry.Cities {
				if serverCity.CityName == city {
					serverCity.Servers = append(serverCity.Servers, server)
					return serversMap
				}
			}

			serverCity := pb.ServerCity{
				CityName: city,
				Servers:  []*pb.Server{server},
			}
			serverCountry.Cities = append(serverCountry.Cities, &serverCity)
			return serversMap
		}
	}

	serverCity := pb.ServerCity{
		CityName: city,
		Servers:  []*pb.Server{server},
	}

	serverCountry := pb.ServerCountry{
		CountryCode: countryCode,
		CountryName: countryName,
		Cities:      []*pb.ServerCity{&serverCity},
	}

	serversMap = append(serversMap, &serverCountry)
	return serversMap
}

func sortServersMap(serversMap []*pb.ServerCountry) []*pb.ServerCountry {
	slices.SortFunc(serversMap, func(a *pb.ServerCountry, b *pb.ServerCountry) int {
		return strings.Compare(a.CountryCode, b.CountryCode)
	})

	for _, serversCountry := range serversMap {
		slices.SortFunc(serversCountry.Cities, func(a *pb.ServerCity, b *pb.ServerCity) int {
			return strings.Compare(a.CityName, b.CityName)
		})
		for _, serversCity := range serversCountry.Cities {
			slices.SortFunc(serversCity.Servers, func(a *pb.Server, b *pb.Server) int {
				if a.Id < b.Id {
					return -1
				}

				if a.Id > b.Id {
					return 1
				}

				return 0
			})
		}
	}

	return serversMap
}

func getServer(id int,
	name string,
	country string,
	countryCode string,
	city string,
	virtual bool,
	groups core.Groups,
	technologyIDs []core.ServerTechnology) core.Server {
	technologies := core.Technologies{}
	for _, techID := range technologyIDs {
		technologies = append(technologies, core.Technology{
			ID:    techID,
			Pivot: core.Pivot{Status: core.Online},
		})
	}

	specifications := []core.Specification{}
	if virtual {
		specifications = append(specifications, core.Specification{
			Identifier: core.VirtualLocation,
			Values: []struct {
				Value string "json:\"value\""
			}{
				{Value: "True"},
			},
		})
	}

	return core.Server{
		ID:       int64(id),
		Status:   core.Online,
		Hostname: name,
		Locations: core.Locations{
			{
				Country: core.Country{
					Name: country,
					Code: countryCode,
					City: core.City{
						Name: city,
					},
				}},
		},
		Specifications: specifications,
		Groups:         groups,
		Technologies:   technologies,
	}
}

func TestServers(t *testing.T) {
	category.Set(t, category.Unit)

	server1ID := 1
	server1Hostname := "server1"
	server1Country := "Germany"
	server1CountryCode := "de"
	server1City := "Berlin"

	server2ID := 2
	server2Hostname := "server2"
	server2Country := "France"
	server2CountryCode := "fr"
	server2City := "Paris"

	server3ID := 3
	server3Hostname := "server3"
	server3Country := "Lithuania"
	server3CountryCode := "lt"
	server3City := "Vilnius"

	server4ID := 4
	server4Hostname := "server4"
	server4Country := "Poland"
	server4CountryCode := "pl"
	server4City := "Warsaw"

	server5ID := 5
	server5Hostname := "server4"
	server5Country := "Iceland"
	server5CountryCode := "is"
	server5City := "Reykjavik"

	// legacy XOR only server: must not be listed on any technology
	server6ID := 6
	server6Hostname := "server6"
	server6Country := "Canada"
	server6CountryCode := "ca"
	server6City := "Toronto"

	servers := core.Servers{
		getServer(server1ID,
			server1Hostname,
			server1Country,
			server1CountryCode,
			server1City,
			true,
			core.Groups{
				{
					ID:    config.ServerGroup_STANDARD_VPN_SERVERS,
					Title: "P2P",
				},
				{
					ID:    config.ServerGroup_P2P,
					Title: "P2P",
				},
				{
					ID:    config.ServerGroup_NETFLIX_USA,
					Title: "Netflix USA",
				},
			},
			[]core.ServerTechnology{
				core.L2TP,
				core.HTTPProxy,
				core.WireguardTech,
				core.OpenVPNTCP,
			}),
		getServer(server2ID,
			server2Hostname,
			server2Country,
			server2CountryCode,
			server2City,
			false,
			core.Groups{
				{
					ID:    config.ServerGroup_STANDARD_VPN_SERVERS,
					Title: "Standard VPN",
				},
			},
			[]core.ServerTechnology{
				core.L2TP,
				core.HTTPProxy,
				core.OpenVPNTCP,
				core.OpenVPNUDP,
			}),
		getServer(server3ID,
			server3Hostname,
			server3Country,
			server3CountryCode,
			server3City,
			false,
			core.Groups{
				{
					ID:    config.ServerGroup_STANDARD_VPN_SERVERS,
					Title: "Standard VPN",
				},
				{
					ID:    config.ServerGroup_ANTI_DDOS,
					Title: "Anti DDOS",
				},
			},
			[]core.ServerTechnology{
				core.L2TP,
				core.HTTPProxy,
				core.OpenVPNTCP,
				core.WireguardTech,
			}),
		getServer(server4ID,
			server4Hostname,
			server4Country,
			server4CountryCode,
			server4City,
			true,
			core.Groups{
				{
					ID:    config.ServerGroup_STANDARD_VPN_SERVERS,
					Title: "Standard VPN",
				},
				{
					ID:    config.ServerGroup_OBFUSCATED,
					Title: "Obfuscated",
				},
				{
					ID:    config.ServerGroup_NETFLIX_USA,
					Title: "Anti DDOS",
				},
				{
					ID:    config.ServerGroup_ANTI_DDOS,
					Title: "Anti DDOS",
				},
			},
			[]core.ServerTechnology{
				core.L2TP,
				core.HTTPProxy,
				core.OpenVPNUDPObfuscated,
				core.OpenVPNTCPObfuscated,
				core.OpenVPNUDP,
				core.OpenVPNTCP,
				core.WireguardTech,
			}),
		getServer(server5ID,
			server5Hostname,
			server5Country,
			server5CountryCode,
			server5City,
			false,
			core.Groups{
				{
					ID:    config.ServerGroup_STANDARD_VPN_SERVERS,
					Title: "Standard VPN",
				},
				{
					ID:    config.ServerGroup_OBFUSCATED,
					Title: "Obfuscated",
				},
				{
					ID:    config.ServerGroup_NETFLIX_USA,
					Title: "Anti DDOS",
				},
				{
					ID:    config.ServerGroup_ANTI_DDOS,
					Title: "Anti DDOS",
				},
			},
			[]core.ServerTechnology{
				core.L2TP,
				core.OpenVPNUDPObfuscated,
				core.OpenVPNTCPObfuscated,
				core.OpenVPNUDP,
				core.OpenVPNTCP,
			}),
		getServer(server6ID,
			server6Hostname,
			server6Country,
			server6CountryCode,
			server6City,
			false,
			core.Groups{
				{
					ID:    config.ServerGroup_OBFUSCATED,
					Title: "Obfuscated",
				},
			},
			[]core.ServerTechnology{
				core.OpenVPNUDPObfuscated,
				core.OpenVPNTCPObfuscated,
			}),
	}

	expectedServer1 := pb.Server{
		Id:           int64(server1ID),
		HostName:     server1Hostname,
		Virtual:      true,
		ServerGroups: []config.ServerGroup{config.ServerGroup_P2P, config.ServerGroup_STANDARD_VPN_SERVERS},
		Technologies: []pb.Technology{
			pb.Technology_NORDLYNX,
			pb.Technology_OPENVPN_TCP,
		},
	}
	expectedServer2 := pb.Server{
		Id:           int64(server2ID),
		HostName:     server2Hostname,
		Virtual:      false,
		ServerGroups: []config.ServerGroup{config.ServerGroup_STANDARD_VPN_SERVERS},
		Technologies: []pb.Technology{
			pb.Technology_OPENVPN_TCP,
			pb.Technology_OPENVPN_UDP,
		},
	}
	expectedServer3 := pb.Server{
		Id:           int64(server3ID),
		HostName:     server3Hostname,
		Virtual:      false,
		ServerGroups: []config.ServerGroup{config.ServerGroup_STANDARD_VPN_SERVERS},
		Technologies: []pb.Technology{
			pb.Technology_OPENVPN_TCP,
			pb.Technology_NORDLYNX,
		},
	}

	expectedServersOpenVPNTCP := []*pb.ServerCountry{}
	expectedServersOpenVPNTCP = addToServersMap(expectedServersOpenVPNTCP, "de", "Germany", "Berlin", &expectedServer1)
	expectedServersOpenVPNTCP = addToServersMap(expectedServersOpenVPNTCP, "fr", "France", "Paris", &expectedServer2)
	expectedServersOpenVPNTCP = addToServersMap(expectedServersOpenVPNTCP, "lt", "Lithuania", "Vilnius", &expectedServer3)

	expectedServer4 := pb.Server{
		Id:           int64(server4ID),
		HostName:     server4Hostname,
		Virtual:      true,
		ServerGroups: []config.ServerGroup{config.ServerGroup_STANDARD_VPN_SERVERS},
		Technologies: []pb.Technology{
			pb.Technology_OPENVPN_UDP,
			pb.Technology_OPENVPN_TCP,
			pb.Technology_NORDLYNX,
		},
	}
	expectedServer5 := pb.Server{
		Id:           int64(server5ID),
		HostName:     server5Hostname,
		Virtual:      false,
		ServerGroups: []config.ServerGroup{config.ServerGroup_STANDARD_VPN_SERVERS},
		Technologies: []pb.Technology{
			pb.Technology_OPENVPN_UDP,
			pb.Technology_OPENVPN_TCP,
		},
	}

	expectedServersOpenVPNUDP := []*pb.ServerCountry{}
	expectedServersOpenVPNUDP = addToServersMap(expectedServersOpenVPNUDP, "fr", "France", "Paris", &expectedServer2)
	expectedServersOpenVPNUDP = addToServersMap(expectedServersOpenVPNUDP, "pl", "Poland", "Warsaw", &expectedServer4)
	expectedServersOpenVPNUDP = addToServersMap(expectedServersOpenVPNUDP, "is", "Iceland", "Reykjavik", &expectedServer5)
	expectedServersOpenVPNTCP = addToServersMap(expectedServersOpenVPNTCP, "pl", "Poland", "Warsaw", &expectedServer4)
	expectedServersOpenVPNTCP = addToServersMap(expectedServersOpenVPNTCP, "is", "Iceland", "Reykjavik", &expectedServer5)

	expectedServersWireguardNonVirtual := []*pb.ServerCountry{}
	expectedServersWireguardNonVirtual = addToServersMap(
		expectedServersWireguardNonVirtual,
		"lt",
		"Lithuania",
		"Vilnius",
		&pb.Server{
			Id:           int64(server3ID),
			HostName:     server3Hostname,
			Virtual:      false,
			ServerGroups: []config.ServerGroup{config.ServerGroup_STANDARD_VPN_SERVERS},
			Technologies: []pb.Technology{
				pb.Technology_OPENVPN_TCP,
				pb.Technology_NORDLYNX,
			},
		})

	tests := []struct {
		name             string
		serversList      core.Servers
		serversErr       error
		allowVirtual     bool
		technology       config.Technology
		protocol         config.Protocol
		configErr        error
		expectedResponse *pb.ServersResponse
	}{
		{
			name:         "success openvpn TCP",
			serversList:  servers,
			allowVirtual: true,
			technology:   config.Technology_OPENVPN,
			protocol:     config.Protocol_TCP,
			expectedResponse: &pb.ServersResponse{
				Response: &pb.ServersResponse_Servers{Servers: &pb.ServersMap{
					ServersByCountry: expectedServersOpenVPNTCP,
				}},
			},
		},
		{
			name:         "success openvpn UDP ignores legacy XOR technologies and tag",
			serversList:  servers,
			allowVirtual: true,
			technology:   config.Technology_OPENVPN,
			protocol:     config.Protocol_UDP,
			expectedResponse: &pb.ServersResponse{
				Response: &pb.ServersResponse_Servers{Servers: &pb.ServersMap{
					ServersByCountry: expectedServersOpenVPNUDP,
				}},
			},
		},
		{
			name:         "success wireguard non virtual",
			serversList:  servers,
			allowVirtual: false,
			technology:   config.Technology_NORDLYNX,
			expectedResponse: &pb.ServersResponse{
				Response: &pb.ServersResponse_Servers{Servers: &pb.ServersMap{
					ServersByCountry: expectedServersWireguardNonVirtual,
				}},
			},
		},
		{
			name:      "failure because of config error",
			configErr: fmt.Errorf("failed to load config"),
			expectedResponse: &pb.ServersResponse{
				Response: &pb.ServersResponse_Error{
					Error: pb.ServersError_GET_CONFIG_ERROR,
				},
			},
		},
		{
			name:        "failure because of filter error",
			serversList: core.Servers{}, // servers will return an error because it fails to find available servers
			expectedResponse: &pb.ServersResponse{
				Response: &pb.ServersResponse_Error{
					Error: pb.ServersError_FILTER_SERVERS_ERROR,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfgManager := mock.NewMockConfigManager()
			cfgManager.LoadErr = test.configErr
			cfgManager.Cfg.Technology = test.technology
			cfgManager.Cfg.VirtualLocation.Set(test.allowVirtual)
			cfgManager.Cfg.AutoConnectData.Protocol = test.protocol

			dm := DataManager{}
			dm.serversData.Servers = test.serversList
			r := RPC{dm: &dm, cm: cfgManager}

			resp, err := r.GetServers(context.Background(), &pb.Empty{})
			assert.Nil(t, err, "Unexpected error returned by servers RPC.")
			assert.IsType(t, test.expectedResponse, resp)

			if test.configErr != nil || test.serversErr != nil {
				assert.Equal(t, test.expectedResponse, resp)
				return
			}

			sortedExpectedServers := sortServersMap(test.expectedResponse.GetServers().GetServersByCountry())
			sortedActual := sortServersMap(resp.GetServers().GetServersByCountry())
			assert.Equal(t, sortedExpectedServers, sortedActual)
		})
	}
}

func TestLegacyXORServersNeverSurface(t *testing.T) {
	category.Set(t, category.Unit)

	standard := getServer(1, "standard1", "Germany", "de", "Berlin", false,
		core.Groups{{ID: config.ServerGroup_STANDARD_VPN_SERVERS, Title: "Standard VPN"}},
		[]core.ServerTechnology{core.OpenVPNTCP, core.OpenVPNUDP, core.WireguardTech, core.NordWhisperTech})
	legacyXOR := getServer(2, "xor1", "Canada", "ca", "Toronto", false,
		core.Groups{{ID: config.ServerGroup_OBFUSCATED, Title: "Obfuscated Servers"}},
		[]core.ServerTechnology{core.OpenVPNUDPObfuscated, core.OpenVPNTCPObfuscated})

	tests := []struct {
		tech  config.Technology
		proto config.Protocol
	}{
		{tech: config.Technology_NORDLYNX, proto: config.Protocol_UDP},
		{tech: config.Technology_OPENVPN, proto: config.Protocol_TCP},
		{tech: config.Technology_OPENVPN, proto: config.Protocol_UDP},
		{tech: config.Technology_NORDWHISPER, proto: config.Protocol_Webtunnel},
	}

	for _, test := range tests {
		t.Run(test.tech.String()+"/"+test.proto.String(), func(t *testing.T) {
			dm := DataManager{serversData: ServersData{Servers: core.Servers{standard, legacyXOR}}}
			cfgManager := mock.NewMockConfigManager()
			cfgManager.Cfg.Technology = test.tech
			cfgManager.Cfg.AutoConnectData.Protocol = test.proto
			cfgManager.Cfg.VirtualLocation.Set(true)
			r := RPC{dm: &dm, cm: cfgManager}

			resp, err := r.GetServers(context.Background(), &pb.Empty{})
			assert.NoError(t, err)
			countries := resp.GetServers().GetServersByCountry()
			assert.Len(t, countries, 1, "only the standard server's country is expected")
			for _, country := range countries {
				for _, city := range country.Cities {
					for _, server := range city.Servers {
						assert.NotEqual(t, legacyXOR.Hostname, server.HostName)
					}
				}
			}

			groups, err := dm.Groups(test.tech, test.proto, true)
			assert.NoError(t, err)
			names := make([]string, 0, len(groups))
			for _, group := range groups {
				names = append(names, group.Name)
			}
			assert.NotContains(t, names, "Obfuscated_Servers")
		})
	}

	t.Run("a fleet of only XOR servers lists nothing", func(t *testing.T) {
		dm := DataManager{serversData: ServersData{Servers: core.Servers{legacyXOR}}}
		groups, err := dm.Groups(config.Technology_NORDWHISPER, config.Protocol_Webtunnel, true)
		assert.NoError(t, err)
		assert.Empty(t, groups)
	})
}

func TestObfuscatedGroupNeverComesFromAServerTag(t *testing.T) {
	category.Set(t, category.Unit)

	tagged := getServer(1, "tagged1", "Germany", "de", "Berlin", false,
		core.Groups{
			{ID: config.ServerGroup_STANDARD_VPN_SERVERS, Title: "Standard VPN servers"},
			{ID: config.ServerGroup_OBFUSCATED, Title: "Obfuscated Servers"},
		},
		[]core.ServerTechnology{
			core.OpenVPNTCP,
			core.OpenVPNUDP,
			core.OpenVPNUDPObfuscated,
			core.OpenVPNTCPObfuscated,
			core.NordWhisperTech,
		})

	tests := []struct {
		tech  config.Technology
		proto config.Protocol
	}{
		{tech: config.Technology_NORDLYNX, proto: config.Protocol_UDP},
		{tech: config.Technology_OPENVPN, proto: config.Protocol_TCP},
		{tech: config.Technology_OPENVPN, proto: config.Protocol_UDP},
		{tech: config.Technology_NORDWHISPER, proto: config.Protocol_Webtunnel},
	}

	for _, test := range tests {
		t.Run(test.tech.String()+"/"+test.proto.String(), func(t *testing.T) {
			dm := DataManager{serversData: ServersData{Servers: core.Servers{tagged}}}

			groups, err := dm.Groups(test.tech, test.proto, true)
			assert.NoError(t, err)
			names := make([]string, 0, len(groups))
			for _, group := range groups {
				names = append(names, group.Name)
			}
			assert.NotContains(t, names, "Obfuscated_Servers")

			cfgManager := mock.NewMockConfigManager()
			cfgManager.Cfg.Technology = test.tech
			cfgManager.Cfg.AutoConnectData.Protocol = test.proto
			cfgManager.Cfg.VirtualLocation.Set(true)
			r := RPC{dm: &dm, cm: cfgManager}

			resp, err := r.GetServers(context.Background(), &pb.Empty{})
			assert.NoError(t, err)
			for _, country := range resp.GetServers().GetServersByCountry() {
				for _, city := range country.Cities {
					for _, server := range city.Servers {
						assert.NotContains(t, server.ServerGroups, config.ServerGroup_OBFUSCATED)
					}
				}
			}
		})
	}
}

func TestServersValidation(t *testing.T) {
	category.Set(t, category.Unit)

	// here expecting valid list of servers
	servers := core_test.ServersList()

	assert.NotNil(t, servers)
	assert.NoError(t, servers.Validate())

	// adding invalid record to the servers list
	servers = append(servers, core.Server{})

	assert.Error(t, servers.Validate())
}
