package core_test

import (
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
)

func CountriesList() core.Countries {
	return core.Countries{
		{
			Name: "Latvia",
			Code: "LV",
			Cities: []core.City{
				{Name: "Riga"},
			},
		},
		{
			Name: "United Kingdom",
			Code: "GB",
			Cities: []core.City{
				{Name: "London"},
				{Name: "Liverpool"},
			},
		},
		{
			Name: "France",
			Code: "FR",
			Cities: []core.City{
				{Name: "Paris"},
				{Name: "Nice"},
			},
		},
		{
			Name: "Lithuania",
			Code: "LT",
			Cities: []core.City{
				{Name: "Vilnius"},
				{Name: "Kaunas"},
			},
		},
		{
			Name: "Germany",
			Code: "DE",
			ID:   133,
			Cities: []core.City{
				{Name: "Berlin", ID: 28},
			},
		},
		{
			Name: "Algeria",
			Code: "DZ",
			Cities: []core.City{
				{Name: "Algiers"},
			},
		},
		{
			Name: "Italy",
			Code: "IT",
			ID:   150,
			Cities: []core.City{
				{ID: 1, Name: "Rome"},
			},
		},
	}
}

func ServersList() core.Servers {
	obfuscatedTechnologies := core.Technologies{
		core.Technology{
			ID:    core.OpenVPNTCPObfuscated,
			Pivot: core.Pivot{Status: core.Online},
		},
		core.Technology{
			ID:    core.OpenVPNUDPObfuscated,
			Pivot: core.Pivot{Status: core.Online},
		},
	}

	technologies := core.Technologies{
		core.Technology{
			ID:    core.OpenVPNTCP,
			Pivot: core.Pivot{Status: core.Online},
		},
		core.Technology{
			ID:    core.OpenVPNUDP,
			Pivot: core.Pivot{Status: core.Online},
		},
		core.Technology{
			ID:    core.WireguardTech,
			Pivot: core.Pivot{Status: core.Online},
		},
	}

	standardGroups := core.Groups{
		core.Group{
			ID:    config.ServerGroup_P2P,
			Title: "P2P",
		},
		core.Group{
			ID:    config.ServerGroup_DOUBLE_VPN,
			Title: "Double VPN",
		},
		core.Group{
			ID:    config.ServerGroup_ONION_OVER_VPN,
			Title: "Double VPN",
		},
		core.Group{
			ID:    config.ServerGroup_STANDARD_VPN_SERVERS,
			Title: "Standard VPN Servers",
		},
	}

	obfuscatedGroups := core.Groups{
		core.Group{
			ID:    config.ServerGroup_OBFUSCATED,
			Title: "Obfuscated Servers",
		},
	}

	dipGroups := core.Groups{
		core.Group{
			ID:    config.ServerGroup_DEDICATED_IP,
			Title: "Dedicated IP",
		},
		core.Group{
			ID:    config.ServerGroup_STANDARD_VPN_SERVERS,
			Title: "Standard VPN Servers",
		},
	}

	virtualServer := []core.Specification{
		{
			Identifier: core.VirtualLocation,
			Values: []struct {
				Value string "json:\"value\""
			}{
				{Value: "true"},
			},
		},
	}

	servers := core.Servers{
		core.Server{
			ID:           1,
			Name:         "France #1",
			Hostname:     "fr1.nordvpn.com",
			Status:       core.Online,
			Technologies: technologies,
			CreatedAt:    "2006-01-02 15:04:05",
			Station:      "127.0.0.1",
			Locations: core.Locations{
				core.Location{
					Country: core.Country{
						Name: "France",
						Code: "FR",
						City: core.City{Name: "Paris"},
					},
				},
			},
			Groups: standardGroups,
		},
		core.Server{
			ID:           2,
			Name:         "Germany #3",
			Hostname:     "de3.nordvpn.com",
			Status:       core.Online,
			Technologies: technologies,
			CreatedAt:    "2006-01-02 15:04:05",
			Station:      "127.0.0.1",
			Locations: core.Locations{
				core.Location{
					Country: core.Country{
						Name: "Germany",
						ID:   133,
						Code: "DE",
						City: core.City{Name: "Berlin", ID: 28},
					},
				},
			},
			Groups: standardGroups,
		},
		core.Server{
			ID:        3,
			Name:      "Lithuania #16",
			Hostname:  "lt16.nordvpn.com",
			CreatedAt: "2006-01-02 15:04:05",
			Station:   "127.0.0.1",
			Technologies: core.Technologies{
				core.Technology{
					ID:    core.WireguardTech,
					Pivot: core.Pivot{Status: core.Online},
				},
			},
			Status: core.Online,
			Locations: core.Locations{
				core.Location{
					Country: core.Country{
						Name: "Lithuania",
						Code: "LT",
						City: core.City{Name: "Vilnius"},
					},
				},
			},
			Groups: standardGroups,
		},
		core.Server{
			ID:           4,
			Name:         "Lithuania #15",
			Hostname:     "lt15.nordvpn.com",
			Status:       core.Online,
			Technologies: technologies,
			CreatedAt:    "2006-01-02 15:04:05",
			Station:      "127.0.0.1",
			Locations: core.Locations{
				core.Location{
					Country: core.Country{
						Name: "Lithuania",
						Code: "LT",
						City: core.City{Name: "Kaunas"},
					},
				},
			},
			Specifications: virtualServer,
			Groups:         standardGroups,
		},
		core.Server{
			ID:           5,
			Name:         "Lithuania #17",
			Hostname:     "lt17.nordvpn.com",
			Status:       core.Online,
			Technologies: obfuscatedTechnologies,
			CreatedAt:    "2006-01-02 15:04:05",
			Station:      "127.0.0.1",
			Locations: core.Locations{
				core.Location{
					Country: core.Country{
						Name: "Lithuania",
						Code: "LT",
						ID:   123,
						City: core.City{Name: "Vilnius", ID: 12},
					},
				},
			},
			Groups: obfuscatedGroups,
		},
		core.Server{
			ID:           7,
			Name:         "Lithuania #7",
			Hostname:     "lt7.nordvpn.com",
			Status:       core.Online,
			Technologies: technologies,
			CreatedAt:    "2006-01-02 15:04:05",
			Station:      "127.0.0.1",
			Locations: core.Locations{
				core.Location{
					Country: core.Country{
						Name: "Lithuania",
						Code: "LT",
						City: core.City{Name: "Vilnius"},
					},
				},
			},
			Groups: dipGroups,
		},
		core.Server{
			ID:           8,
			Name:         "Lithuania #8",
			Hostname:     "lt8.nordvpn.com",
			Status:       core.Online,
			Technologies: technologies,
			CreatedAt:    "2006-01-02 15:04:05",
			Station:      "127.0.0.1",
			Locations: core.Locations{
				core.Location{
					Country: core.Country{
						Name: "Lithuania",
						Code: "LT",
						City: core.City{Name: "Kaunas"},
					},
				},
			},
			Groups: dipGroups,
		},
		core.Server{
			ID:           9,
			Name:         "Lithuania #9",
			Hostname:     "lt9.nordvpn.com",
			Status:       core.Offline,
			Technologies: technologies,
			CreatedAt:    "2006-01-02 15:04:05",
			Station:      "127.0.0.1",
			Locations: core.Locations{
				core.Location{
					Country: core.Country{
						Name: "Lithuania",
						Code: "LT",
						City: core.City{Name: "Kaunas"},
					},
				},
			},
			Groups: dipGroups,
		},
		core.Server{
			ID:           10,
			Name:         "Algeria #1",
			Hostname:     "dz1.nordvpn.com",
			Status:       core.Online,
			Technologies: technologies,
			CreatedAt:    "2006-01-02 15:04:05",
			Station:      "127.0.0.1",
			Locations: core.Locations{
				core.Location{
					Country: core.Country{
						Name: "Algeria",
						Code: "DZ",
						City: core.City{Name: "Algiers"},
					},
				},
			},
			Specifications: virtualServer,
			Groups:         standardGroups,
		},
		core.Server{
			ID:           11,
			Name:         "Algeria #2",
			Hostname:     "dz2.nordvpn.com",
			Status:       core.Online,
			Technologies: obfuscatedTechnologies,
			CreatedAt:    "2006-01-02 15:04:05",
			Station:      "127.0.0.1",
			Locations: core.Locations{
				core.Location{
					Country: core.Country{
						Name: "Algeria",
						Code: "DZ",
						City: core.City{Name: "Algiers"},
					},
				},
			},
			Groups: obfuscatedGroups,
		},
		core.Server{
			ID:           12,
			Name:         "Italy #1",
			Hostname:     "it1.nordvpn.com",
			Status:       core.Online,
			Technologies: technologies,
			CreatedAt:    "2006-01-02 15:04:05",
			Station:      "127.0.0.1",
			Locations: core.Locations{
				core.Location{
					Country: core.Country{
						Name: "Italy",
						Code: "IT",
						ID:   150,
						City: core.City{ID: 1, Name: "Rome"},
					},
				},
			},
			Groups: standardGroups,
		},
		core.Server{
			ID:           13,
			Name:         "Italy #2",
			Hostname:     "it2.nordvpn.com",
			Status:       core.Online,
			Technologies: technologies,
			CreatedAt:    "2006-01-02 15:04:05",
			Station:      "127.0.0.1",
			Locations: core.Locations{
				core.Location{
					Country: core.Country{
						Name: "Italy",
						Code: "IT",
						ID:   150,
						City: core.City{ID: 1, Name: "Rome"},
					},
				},
			},
			Groups: standardGroups,
		},
	}

	// if at least one record is not valid - reject whole list, assuming something wrong is with whole list
	if err := servers.Validate(); err != nil {
		return nil
	}

	for i, server := range servers {
		servers[i].Keys = server.GenerateKeys()
	}

	return servers
}
