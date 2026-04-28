package core

import (
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/google/uuid"
)

type DedicatedServersAPIMock struct {
	DedicatedServersResponse core.DedicatedServers
	ConnectResponse          core.ConnectResponse

	DedicatedServerErr error
	ConnectErr         error
}

func (*DedicatedServersAPIMock) RegisterDevice(core.DevicesRequest) (core.DevicesResponse, error) {
	return core.DevicesResponse{}, nil
}

func (*DedicatedServersAPIMock) UpdateDevice(uuid.UUID, core.UpdateDeviceRequest) (core.DevicesResponse, error) {
	return core.DevicesResponse{}, nil
}

func (d *DedicatedServersAPIMock) DedicatedServers() (core.DedicatedServers, error) {
	return d.DedicatedServersResponse, d.DedicatedServerErr
}

func (d *DedicatedServersAPIMock) Connect(serverUUID string, connectRequest core.ConnectRequest) (core.ConnectResponse, error) {
	return d.ConnectResponse, d.ConnectErr
}
