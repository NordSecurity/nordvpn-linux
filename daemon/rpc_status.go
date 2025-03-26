package daemon

import (
	"context"
	"time"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
)

// Status of daemon and connection
func (r *RPC) Status(context.Context, *pb.Empty) (*pb.StatusResponse, error) {
	status := r.netw.ConnectionStatus()
	if status.State == pb.ConnectionState_UNKNOWN_STATE || status.State == pb.ConnectionState_DISCONNECTED {
		return &pb.StatusResponse{
			State:  pb.ConnectionState_DISCONNECTED,
			Uptime: -1,
		}, nil
	}

	var uptime int64
	if status.StartTime != nil {
		uptime = int64(time.Since(*status.StartTime))
	} else {
		uptime = -1
	}

	requestedConnParams := r.RequestedConnParams.Get()

	return &pb.StatusResponse{
		State:           status.State,
		Technology:      status.Technology,
		Protocol:        status.Protocol,
		Ip:              status.IP.String(),
		Hostname:        status.Hostname,
		Name:            status.Name,
		Country:         status.Country,
		City:            status.City,
		Download:        status.Download,
		Upload:          status.Upload,
		Uptime:          uptime,
		VirtualLocation: status.VirtualLocation,
		Parameters: &pb.ConnectionParameters{
			Source:  requestedConnParams.ConnectionSource,
			Country: requestedConnParams.Country,
			City:    requestedConnParams.City,
			Group:   requestedConnParams.Group,
		},
		PostQuantum: status.PostQuantum,
	}, nil
}
