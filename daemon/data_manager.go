package daemon

import (
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/internal"

	"github.com/coreos/go-semver/semver"
	mapset "github.com/deckarep/golang-set"
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

func (dm *DataManager) SetAppData(
	countryNames map[bool]map[config.Protocol]mapset.Set,
	cityNames map[bool]map[config.Protocol]map[string]mapset.Set,
	groupNames map[bool]map[config.Protocol]mapset.Set,
) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.appData.CountryNames = countryNames
	dm.appData.CityNames = cityNames
	dm.appData.GroupNames = groupNames
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
