package daemon

import (
	"context"
	"log"
	"strconv"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func (r *RPC) SetProtocol(ctx context.Context, in *pb.SetProtocolRequest) (*pb.Payload, error) {
	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
	}

	payload := &pb.Payload{}
	if err := r.cm.SaveWith(func(c config.Config) config.Config {
		c.AutoConnectData.Protocol = in.GetProtocol()
		return c
	}); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.Payload{
			Type: internal.CodeConfigError,
		}, nil
	}

	r.events.Settings.Protocol.Publish(in.GetProtocol())
	payload.Type = internal.CodeSuccess
	payload.Data = []string{strconv.FormatBool(r.netw.IsVPNActive())}
	return payload, nil
}
