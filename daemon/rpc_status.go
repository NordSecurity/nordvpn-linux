package daemon

import (
	"context"
	"time"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Status of daemon and connection
func (r *RPC) Status(context.Context, *pb.Empty) (*pb.StatusResponse, error) {
	status := r.connectionInfo.Status()
	//exhaustive:ignore
	switch status.State {
	case pb.ConnectionState_PAUSED:
		return &pb.StatusResponse{
			State:                     pb.ConnectionState_PAUSED,
			Uptime:                    -1,
			PausedAt:                  timestamppb.New(status.PausedAt),
			PauseRemainingDurationSec: uint32(status.PauseRemainingTimeSec),
		}, nil
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
		Download:        status.Rx,
		Upload:          status.Tx,
		Uptime:          calculateUptime(status.StartTime),
		VirtualLocation: status.IsVirtualLocation,
		Parameters: &pb.ConnectionParameters{
			Source:      requestedConnParams.ConnectionSource,
			Country:     requestedConnParams.Country,
			City:        requestedConnParams.City,
			Group:       requestedConnParams.Group,
			ServerName:  requestedConnParams.ServerName,
			CountryCode: requestedConnParams.CountryCode,
		},

		PostQuantum:               status.IsPostQuantum,
		Obfuscated:                status.IsObfuscated,
		PausedAt:                  timestamppb.New(status.PausedAt),
		PauseRemainingDurationSec: uint32(status.PauseRemainingTimeSec),
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
