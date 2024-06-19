package daemon

import (
	"context"
	"log"
	"strconv"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func (r *RPC) SetTechnology(ctx context.Context, in *pb.SetTechnologyRequest) (*pb.Payload, error) {
	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
	}

	if cfg.Technology == in.GetTechnology() {
		return &pb.Payload{
			Type: internal.CodeNothingToDo,
			Data: []string{in.GetTechnology().String()},
		}, nil
	}

	v, err := r.factory(in.GetTechnology())
	if err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.Payload{
			Type: internal.CodeConfigError,
		}, nil
	}
	r.netw.SetVPN(v)

	payload := &pb.Payload{}
	// payload.Type gets overridden in case of failure
	//
	// Previously it was at the end of the function, which overrode any failures
	// and user was not given any error messages because of it. Most notably,
	// internal.CodeSuccessWithoutAC was overridden with generic internal.CodeSuccess
	payload.Type = internal.CodeSuccess

	protocol := cfg.AutoConnectData.Protocol
	obfuscate := cfg.AutoConnectData.Obfuscate
	if in.GetTechnology() == config.Technology_NORDLYNX {
		protocol = config.Protocol_UDP
		obfuscate = false
	}

	if err := r.cm.SaveWith(func(c config.Config) config.Config {
		c.Technology = in.GetTechnology()
		c.AutoConnectData.Protocol = protocol
		c.AutoConnectData.Obfuscate = obfuscate
		return c
	}); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.Payload{
			Type: internal.CodeConfigError,
		}, nil
	}

	r.events.Settings.Technology.Publish(in.GetTechnology())

	payload.Data = []string{strconv.FormatBool(r.netw.IsVPNActive()), in.GetTechnology().String()}
	return payload, nil
}
