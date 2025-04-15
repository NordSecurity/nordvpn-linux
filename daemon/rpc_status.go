package daemon

import (
	"context"
	"log"
	"time"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/tunnel"
)

// Status of daemon and connection
func (r *RPC) Status(context.Context, *pb.Empty) (*pb.StatusResponse, error) {
	status := r.netw.ConnectionStatus()
	//exhaustive:ignore
	switch status.State {
	case pb.ConnectionState_UNKNOWN_STATE, pb.ConnectionState_DISCONNECTED:
		return &pb.StatusResponse{
			State:  pb.ConnectionState_DISCONNECTED,
			Uptime: -1,
		}, nil
	case pb.ConnectionState_CONNECTING:
		return &pb.StatusResponse{
			State:  pb.ConnectionState_CONNECTING,
			Uptime: -1,
		}, nil
	}

	stats, err := tunnel.GetTransferRates(status.TunnelName)
	if err != nil {
		log.Printf(internal.WarningPrefix+" failed to get transfer rates of '%s': %+v", status.TunnelName, err)
		stats = tunnel.Statistics{}
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
		CountryCode:     status.CountryCode,
		City:            status.City,
		Download:        stats.Rx,
		Upload:          stats.Tx,
		Uptime:          calculateUptime(status.StartTime),
		VirtualLocation: status.VirtualLocation,
		Parameters: &pb.ConnectionParameters{
			Source:      requestedConnParams.ConnectionSource,
			Country:     requestedConnParams.Country,
			City:        requestedConnParams.City,
			Group:       requestedConnParams.Group,
			ServerName:  requestedConnParams.ServerName,
			CountryCode: requestedConnParams.CountryCode,
		},

		PostQuantum: status.PostQuantum,
		Obfuscated:  status.Obfuscated,
	}, nil
}

func calculateUptime(startTime *time.Time) int64 {
	var uptime int64
	if startTime != nil {
		uptime = int64(time.Since(*startTime))
	} else {
		uptime = -1
	}
	return uptime
}
