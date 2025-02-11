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
	"github.com/stretchr/testify/assert"
)

func addToServersMap(serversMap []*pb.ServerCountry,
	countryCode string,
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
				{
					ID:    config.ServerGroup_EUROPE,
					Title: "Europe",
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
					ID:    config.ServerGroup_EUROPE,
					Title: "Europe",
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
					ID:    config.ServerGroup_EUROPE,
					Title: "Europe",
				},
			},
			[]core.ServerTechnology{
				core.L2TP,
				core.OpenVPNUDPObfuscated,
				core.OpenVPNTCPObfuscated,
				core.OpenVPNUDP,
				core.OpenVPNTCP,
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
	expectedServersOpenVPNTCP = addToServersMap(expectedServersOpenVPNTCP, "de", "Berlin", &expectedServer1)
	expectedServersOpenVPNTCP = addToServersMap(expectedServersOpenVPNTCP, "fr", "Paris", &expectedServer2)
	expectedServersOpenVPNTCP = addToServersMap(expectedServersOpenVPNTCP, "lt", "Vilnius", &expectedServer3)

	expectedServer4 := pb.Server{
		Id:           int64(server4ID),
		HostName:     server4Hostname,
		Virtual:      true,
		ServerGroups: []config.ServerGroup{config.ServerGroup_OBFUSCATED, config.ServerGroup_STANDARD_VPN_SERVERS},
		Technologies: []pb.Technology{
			pb.Technology_OBFUSCATED_OPENVPN_UDP,
			pb.Technology_OBFUSCATED_OPENVPN_TCP,
			pb.Technology_OPENVPN_UDP,
			pb.Technology_OPENVPN_TCP,
			pb.Technology_NORDLYNX,
		},
	}
	expectedServer5 := pb.Server{
		Id:           int64(server5ID),
		HostName:     server5Hostname,
		Virtual:      false,
		ServerGroups: []config.ServerGroup{config.ServerGroup_OBFUSCATED, config.ServerGroup_STANDARD_VPN_SERVERS},
		Technologies: []pb.Technology{
			pb.Technology_OBFUSCATED_OPENVPN_UDP,
			pb.Technology_OBFUSCATED_OPENVPN_TCP,
			pb.Technology_OPENVPN_UDP,
			pb.Technology_OPENVPN_TCP,
		},
	}

	expectedServersOpenVPNUDPObfuscated := []*pb.ServerCountry{}
	expectedServersOpenVPNUDPObfuscated = addToServersMap(
		expectedServersOpenVPNUDPObfuscated,
		"pl",
		"Warsaw",
		&expectedServer4)
	expectedServersOpenVPNUDPObfuscated = addToServersMap(
		expectedServersOpenVPNUDPObfuscated,
		"is",
		"Reykjavik",
		&expectedServer5)

	expectedServersWireguardNonVirtual := []*pb.ServerCountry{}
	expectedServersWireguardNonVirtual = addToServersMap(
		expectedServersWireguardNonVirtual,
		"lt",
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
		obfuscate        bool
		allowVirtual     bool
		technology       config.Technology
		protocol         config.Protocol
		configErr        error
		expectedResponse *pb.ServersResponse
	}{
		{
			name:         "success openvpn TCP",
			serversList:  servers,
			obfuscate:    false,
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
			name:         "success openvpn UDP obfuscated",
			serversList:  servers,
			obfuscate:    true,
			allowVirtual: true,
			technology:   config.Technology_OPENVPN,
			protocol:     config.Protocol_UDP,
			expectedResponse: &pb.ServersResponse{
				Response: &pb.ServersResponse_Servers{Servers: &pb.ServersMap{
					ServersByCountry: expectedServersOpenVPNUDPObfuscated,
				}},
			},
		},
		{
			name:         "success wireguard non virtual",
			serversList:  servers,
			obfuscate:    false,
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
			serversList: core.Servers{}, // servers will return an error because it will fails to find available servers
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
			cfgManager.Cfg.AutoConnectData.Obfuscate = test.obfuscate
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

			sorterdExpectedServers := sortServersMap(test.expectedResponse.GetServers().GetServersByCountry())
			sortedActuall := sortServersMap(resp.GetServers().GetServersByCountry())
			assert.Equal(t, sorterdExpectedServers, sortedActuall)
		})
	}
}
