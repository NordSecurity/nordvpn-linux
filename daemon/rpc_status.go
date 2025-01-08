package daemon

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
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

	connectionParameters, err := r.ConnectionParameters.GetConnectionParameters()
	if err != nil {
		log.Println(internal.WarningPrefix, "failed to read connection parameters:", err)
	}

	postQuantum := false
	if connectionParameters, ok := r.netw.GetConnectionParameters(); ok {
		postQuantum = connectionParameters.PostQuantum
	}

	insights := r.dm.GetInsightsData()
	var ip string
	if insights.Insights.IP != nil {
		ip = *insights.Insights.IP
	} else {
		ip = status.IP.String()
	}

	return &pb.StatusResponse{
		State:           string(status.State),
		Technology:      status.Technology,
		Protocol:        status.Protocol,
		Ip:              ip,
		Hostname:        status.Hostname,
		Name:            status.Name,
		Country:         status.Country,
		City:            status.City,
		Download:        status.Download,
		Upload:          status.Upload,
		Uptime:          uptime,
		VirtualLocation: status.VirtualLocation,
		Parameters: &pb.ConnectionParameters{
			Source:  connectionParameters.ConnectionSource,
			Country: connectionParameters.Parameters.Country,
			City:    connectionParameters.Parameters.City,
			Group:   connectionParameters.Parameters.Group,
		},
		PostQuantum: postQuantum,
	}, nil
}
