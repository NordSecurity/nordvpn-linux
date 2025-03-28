package daemon

import (
	"sync"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
)

type ConnectionParameters struct {
	// ConnectionString is a string used to establish the connection(if any), for example lt1111
	ConnectionString string
	ConnectionSource pb.ConnectionSource
	Parameters       ServerParameters
}

type ParametersStorage struct {
	mu         sync.Mutex
	parameters ConnectionParameters
}

func (c *ParametersStorage) SetConnectionParameters(connectionSource pb.ConnectionSource, parameters ServerParameters) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.parameters = ConnectionParameters{
		ConnectionSource: connectionSource,
		Parameters:       parameters,
	}
}

func (c *ParametersStorage) GetConnectionParameters() (ConnectionParameters, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.parameters, nil
}
