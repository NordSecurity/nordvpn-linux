package daemon

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
)

// Status of daemon and connection
func (r *RPC) Status(context.Context, *pb.Empty) (*pb.StatusResponse, error) {
	if !r.netw.IsVPNActive() {
		return &pb.StatusResponse{
			State:  "Disconnected",
			Uptime: -1,
		}, nil
	}

	status, _ := r.netw.ConnectionStatus()

	var uptime int64
	if status.Uptime != nil {
		uptime = int64(*status.Uptime)
	} else {
		uptime = -1
	}

	switch status.State { //nolint:exhaustive
	case "EXITING":
		status.State = "Disconnecting"
	case "EXITED":
		status.State = "Disconnected"
	case "RECONNECTING":
		status.State = "Reconnecting"
	case "CONNECTED":
		status.State = "Connected"
	default:
		status.State = "Connecting"
	}

	return &pb.StatusResponse{
		State:           string(status.State),
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
	}, nil
}
