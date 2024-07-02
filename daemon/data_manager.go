package daemon

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"

	"github.com/coreos/go-semver/semver"
	mapset "github.com/deckarep/golang-set/v2"
)

type InsightsDataManager interface {
	GetInsightsData() InsightsData
	SetInsightsData(core.Insights) error
}

type DataManager struct {
	appData      AppData
	countryData  CountryData
	insightsData InsightsData
	serversData  ServersData
	versionData  VersionData
	mu           sync.Mutex
}

func NewDataManager(insightsFilePath, serversFilePath, countryFilePath, versionFilePath string) *DataManager {
	return &DataManager{
		countryData:  CountryData{filePath: countryFilePath},
		insightsData: InsightsData{filePath: insightsFilePath},
		serversData:  ServersData{filePath: serversFilePath},
		versionData:  VersionData{filePath: versionFilePath},
	}
}

func (dm *DataManager) LoadData() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	// TODO: after Go 1.20, rewrite using error joining
	if err := dm.countryData.load(); err != nil {
		return fmt.Errorf("loading country data: %w", err)
	}
	if err := dm.insightsData.load(); err != nil {
		return fmt.Errorf("loading insights data: %w", err)
	}
	if err := dm.serversData.load(); err != nil {
		return fmt.Errorf("loading servers data: %w", err)
	}
	if err := dm.versionData.load(); err != nil {
		return fmt.Errorf("loading version data: %w", err)
	}
	return nil
}

func (dm *DataManager) GetInsightsData() InsightsData {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	return dm.insightsData
}

func (dm *DataManager) SetInsightsData(insights core.Insights) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.insightsData.Insights = insights
	return dm.insightsData.save()
}

func (dm *DataManager) SetCountryData(updatedAt time.Time, countries core.Countries, hash string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.countryData.UpdatedAt = updatedAt
	dm.countryData.Countries = countries
	dm.countryData.Hash = hash
	return dm.countryData.save()
}

func (dm *DataManager) CountryDataExists() bool {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	return dm.countryData.exists()
}

func (dm *DataManager) IsCountryDataValid() bool {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	return dm.countryData.isValid()
}

func (dm *DataManager) GetCountryData() CountryData {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	return dm.countryData
}

func (dm *DataManager) ServerDataExists() bool {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	return dm.serversData.exists()
}

func (dm *DataManager) IsServersDataValid() bool {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	return dm.serversData.isValid()
}

func (dm *DataManager) GetServersData() ServersData {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	return dm.serversData
}

func (dm *DataManager) SetServersData(updatedAt time.Time, servers core.Servers, hash string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.serversData.UpdatedAt = updatedAt
	dm.serversData.Servers = servers
	dm.serversData.Hash = hash
	return dm.serversData.save()
}

func (dm *DataManager) UpdateServerPenalty(s core.Server) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	servers := dm.serversData.Servers
	for idx, server := range servers {
		if s.ID == server.ID {
			servers[idx].Status = s.Status
			servers[idx].Penalty = server.PartialPenalty + loadPenalty(s.Load)
			break
		}
	}
	sort.SliceStable(servers, func(i, j int) bool {
		return servers[i].Penalty < servers[j].Penalty
	})
	dm.serversData.Servers = servers
	return dm.serversData.save()
}

func (dm *DataManager) SetServerStatus(s core.Server, status core.Status) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	servers := dm.serversData.Servers
	for idx, server := range servers {
		if s.ID == server.ID {
			servers[idx].Status = status
			break
		}
	}
	dm.serversData.Servers = servers
	return dm.serversData.save()
}

func (dm *DataManager) GetAppData() AppData {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	return dm.appData
}

func (dm *DataManager) GetVersionData() VersionData {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	return dm.versionData
}

func (dm *DataManager) SetVersionData(version semver.Version, newerAvailable bool) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.versionData.version = version
	dm.versionData.newerVersionAvailable = newerAvailable
	if err := dm.versionData.save(); err != nil {
		log.Println(internal.WarningPrefix, err)
	}
}

func toServerTechnology(
	technology config.Technology,
	protocol config.Protocol,
	obfuscated bool,
) (core.ServerTechnology, error) {
	var serverTechnology core.ServerTechnology
	switch technology {
	case config.Technology_NORDLYNX:
		serverTechnology = core.WireguardTech
	case config.Technology_OPENVPN:
		switch protocol {
		case config.Protocol_TCP:
			if obfuscated {
				serverTechnology = core.OpenVPNTCPObfuscated
			} else {
				serverTechnology = core.OpenVPNTCP
			}
		case config.Protocol_UDP:
			if obfuscated {
				serverTechnology = core.OpenVPNUDPObfuscated
			} else {
				serverTechnology = core.OpenVPNUDP
			}
		case config.Protocol_UNKNOWN_PROTOCOL:
			return 0, errors.New("invalid protocol")
		}
	case config.Technology_UNKNOWN_TECHNOLOGY:
		return 0, errors.New("invalid technology")
	}
	return serverTechnology, nil
}

func (dm *DataManager) Countries(
	technology config.Technology,
	protocol config.Protocol,
	obfuscated bool,
	virtualLocation bool,
) ([]*pb.ServerGroup, error) {
	serverTechnology, err := toServerTechnology(technology, protocol, obfuscated)
	if err != nil {
		return nil, err
	}

	dm.mu.Lock()
	defer dm.mu.Unlock()
	countriesSet := mapset.NewSet[string]()
	result := []*pb.ServerGroup{}

	for _, server := range dm.serversData.Servers {
		if !core.IsConnectableVia(serverTechnology)(server) {
			continue
		}

		if !virtualLocation && server.IsVirtualLocation() {
			continue
		}

		country := server.Country()
		if country == nil {
			continue
		}

		if countriesSet.Contains(country.Code) {
			continue
		}

		countriesSet.Add(country.Code)
		group := &pb.ServerGroup{Name: internal.Title(country.Name), VirtualLocation: server.IsVirtualLocation()}
		result = append(result, group)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result, nil
}

func (dm *DataManager) Cities(
	countryName string,
	technology config.Technology,
	protocol config.Protocol,
	obfuscated bool,
	virtualLocation bool,
) ([]*pb.ServerGroup, error) {
	serverTechnology, err := toServerTechnology(technology, protocol, obfuscated)
	if err != nil {
		return nil, err
	}
	countryCode := strings.ToUpper(countryName)
	countryName = strings.ToLower(countryName)

	dm.mu.Lock()
	defer dm.mu.Unlock()
	citiesSet := mapset.NewSet[string]()
	result := []*pb.ServerGroup{}
	for _, server := range dm.serversData.Servers {
		if !core.IsConnectableVia(serverTechnology)(server) {
			continue
		}

		if !virtualLocation && server.IsVirtualLocation() {
			continue
		}

		country := server.Country()
		if country == nil {
			continue
		}

		if citiesSet.Contains(country.City.Name) {
			continue
		}

		if countryCode == country.Code || countryName == strings.ToLower(internal.Title(country.Name)) {
			citiesSet.Add(country.City.Name)
			group := &pb.ServerGroup{Name: internal.Title(country.City.Name), VirtualLocation: server.IsVirtualLocation()}
			result = append(result, group)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result, nil
}

func (dm *DataManager) Groups(
	technology config.Technology,
	protocol config.Protocol,
	obfuscated bool,
	virtualLocation bool,
) ([]*pb.ServerGroup, error) {
	serverTechnology, err := toServerTechnology(technology, protocol, obfuscated)
	if err != nil {
		return nil, err
	}
	dm.mu.Lock()
	defer dm.mu.Unlock()
	groupsSet := mapset.NewSet[string]()
	result := []*pb.ServerGroup{}
	for _, server := range dm.serversData.Servers {
		if !core.IsConnectableVia(serverTechnology)(server) {
			continue
		}

		if !virtualLocation && server.IsVirtualLocation() {
			continue
		}

		for _, group := range server.Groups {
			if groupsSet.Contains(group.Title) {
				continue
			}

			groupsSet.Add(group.Title)
			// special server groups contain both virtual and physical
			// display them always as physical servers
			item := &pb.ServerGroup{Name: internal.Title(group.Title), VirtualLocation: false}
			result = append(result, item)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result, nil
}
