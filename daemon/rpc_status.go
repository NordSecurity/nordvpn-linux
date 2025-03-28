package daemon

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// Status of daemon and connection
func (r *RPC) Status(context.Context, *pb.Empty) (*pb.StatusResponse, error) {
	if !r.netw.IsVPNActive() {
		return &pb.StatusResponse{
			State:  pb.ConnectionState_DISCONNECTED,
			Uptime: -1,
		}, nil
	}

	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.WarningPrefix, "failed to read config:", err)
	}

	status, _ := r.netw.ConnectionStatus()

	var uptime int64
	if status.Uptime != nil {
		uptime = int64(*status.Uptime)
	} else {
		uptime = -1
	}

	connectionParameters, err := r.ConnectionParameters.GetConnectionParameters()
	if err != nil {
		log.Println(internal.WarningPrefix, "failed to read connection parameters:", err)
	}

	postQuantum := false
	obfuscated := false
	if connectionParameters, ok := r.netw.GetConnectionParameters(); ok {
		postQuantum = connectionParameters.PostQuantum
		obfuscated = true
	}

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
		Download:        status.Download,
		Upload:          status.Upload,
		Uptime:          uptime,
		VirtualLocation: status.VirtualLocation,
		Parameters: &pb.ConnectionParameters{
			Source:      connectionParameters.ConnectionSource,
			Country:     connectionParameters.Parameters.Country,
			City:        connectionParameters.Parameters.City,
			Group:       connectionParameters.Parameters.Group,
			CountryCode: connectionParameters.Parameters.CountryCode,
			ServerName:  connectionParameters.Parameters.ServerName,
		},
		PostQuantum: postQuantum,
		Obfuscated:  obfuscated,
	}, nil
}
