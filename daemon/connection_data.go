package daemon

import (
	"sync"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
)

type ConnectionParameters struct {
	ConnectionSource pb.ConnectionSource
	ServerParameters
}

// RequestedConnParamsStorage stores connection parameters as requested by user.
//
// Note that those may not be the same as actual connection parameters.
type RequestedConnParamsStorage struct {
	mu         sync.Mutex
	parameters ConnectionParameters
}

func (c *RequestedConnParamsStorage) Set(connectionSource pb.ConnectionSource, parameters ServerParameters) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.parameters = ConnectionParameters{
		ConnectionSource: connectionSource,
		ServerParameters: parameters,
	}
}

func (c *RequestedConnParamsStorage) Get() ConnectionParameters {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.parameters
}
