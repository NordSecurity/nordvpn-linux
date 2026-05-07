package core

import (
	"net/http"

	"github.com/NordSecurity/nordvpn-linux/core"
)

type ServersAPIMock struct {
	ServerEndpointCalled bool
	ServerResponse       core.Server
}

func (*ServersAPIMock) Servers() (core.Servers, http.Header, error) {
	return core.Servers{}, http.Header{}, nil
}

func (*ServersAPIMock) RecommendedServers(filter core.ServersFilter, longitude, latitude float64) (core.Servers, http.Header, error) {
	return core.Servers{}, http.Header{}, nil
}

func (s *ServersAPIMock) Server(id int64) (*core.Server, error) {
	s.ServerEndpointCalled = true
	return &s.ServerResponse, nil
}

func (*ServersAPIMock) ServersCountries() (core.Countries, http.Header, error) {
	return core.Countries{}, http.Header{}, nil
}
