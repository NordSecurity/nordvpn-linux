package core

import (
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/google/uuid"
)

type DedicatedServersAPIMock struct {
	DedicatedServersResponse core.DedicatedServers
	ConnectResponse          core.DedicatedServerConnectResponse

	// DedicatedServerConnectCheck holds serverUUID provided to DedicatedServerConnectCheck
	ConnectServerUUID string
	// ConnectRequest holds connectRequest provided to DedicatedServerConnectCheck
	ConnectRequest core.DedicatedServerConnectRequest

	DedicatedServerErr error
	ConnectErr         error
	// GetConnectErrFunc is called when it is not nil. The returned error will be returned by
	// DedicatedServerConnectCheck. Used to test handling of device not existing error fallbacks.
	GetConnectErrFunc func() error
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

func (d *DedicatedServersAPIMock) DedicatedServerConnectCheck(serverUUID string, connectRequest core.DedicatedServerConnectRequest) (core.DedicatedServerConnectResponse, error) {
	d.ConnectRequest = connectRequest
	d.ConnectServerUUID = serverUUID

	var connectErr error
	if d.ConnectErr != nil {
		connectErr = d.ConnectErr
	} else if d.GetConnectErrFunc != nil {
		connectErr = d.GetConnectErrFunc()
	}

	return d.ConnectResponse, connectErr
}
