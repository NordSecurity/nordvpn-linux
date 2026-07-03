package core_test

import (
	"fmt"
	"net/http"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
)

const TestRecommendedUUID string = "c0b4c990-3000-457f-8b81-6850b8cdb54e"

type MockServersAPI struct {
	Error                error
	ServersList          core.Servers
	CountriesList        core.Countries
	ServerEndpointCalled bool
}

func NewMockServersAPI() *MockServersAPI {
	return &MockServersAPI{
		ServersList:   ServersList(),
		CountriesList: CountriesList(),
	}
}

func NewMockFailingServersAPI(err error) *MockServersAPI {
	return &MockServersAPI{
		Error: err,
	}
}

func (m *MockServersAPI) Servers() (core.Servers, http.Header, error) {
	return m.ServersList, nil, m.Error
}

func (m *MockServersAPI) RecommendedServers(filter core.ServersFilter, _ float64, _ float64) (core.Servers, http.Header, error) {
	if m.Error != nil {
		return nil, nil, m.Error
	}
	if filter.Group == config.ServerGroup_DEDICATED_IP {
		return nil, nil, fmt.Errorf("API must not be called for Dedicated IP")
	}

	var servers core.Servers
	for _, server := range m.ServersList {
		if server.Status != core.Online || core.IsDedicatedIP(server) {
			continue
		}

		servers = append(servers, server)
	}

	header := http.Header{}
	header.Set("X-Recommendation-Uuid", TestRecommendedUUID)

	return servers, header, nil
}

func (m *MockServersAPI) Server(serverID int64) (*core.Server, error) {
	m.ServerEndpointCalled = true
	if m.Error != nil {
		return nil, m.Error
	}
	for _, server := range m.ServersList {
		if server.ID == serverID {
			return &server, nil
		}
	}

	return nil, fmt.Errorf("not found")
}

func (m *MockServersAPI) ServersCountries() (core.Countries, http.Header, error) {
	return m.CountriesList, nil, m.Error
}

func (m *MockServersAPI) ServersTechnologiesConfigurations(string, int64, core.ServerTechnology) ([]byte, error) {
	return nil, m.Error
}
