package daemon

import (
	"context"
	"errors"
	"log"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/networker"
)

// Logout erases user credentials and disconnects completely
func (r *RPC) Logout(ctx context.Context, in *pb.LogoutRequest) (*pb.Payload, error) {
	if !r.ac.IsLoggedIn() {
		return nil, internal.ErrNotLoggedIn
	}

	var cfg config.Config
	err := r.cm.Load(&cfg)
	if err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.Payload{Type: internal.CodeFailure}, nil
	}

	if err := r.netw.Stop(); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.Payload{Type: internal.CodeFailure}, nil
	}

	if err := r.netw.UnSetMesh(); err != nil && !errors.Is(err, networker.ErrMeshNotActive) {
		log.Println(internal.ErrorPrefix, err)
		return &pb.Payload{Type: internal.CodeFailure}, nil
	}

	tokenData, ok := cfg.TokensData[cfg.AutoConnectData.ID]
	if !ok {
		return &pb.Payload{Type: internal.CodeFailure}, nil
	}

	if !in.GetPersistToken() {
		if err := r.api.DeleteToken(tokenData.Token); err != nil {
			log.Println(internal.ErrorPrefix, "deleting token: ", err)
			switch {
			// This means that token is invalid anyway
			case errors.Is(err, core.ErrUnauthorized):
			case errors.Is(err, core.ErrBadRequest):
			case errors.Is(err, core.ErrServerInternal):
				return &pb.Payload{
					Type: internal.CodeInternalError,
				}, nil
			default:
				return &pb.Payload{
					Type: internal.CodeFailure,
				}, nil
			}
		}

		if err := r.api.Logout(tokenData.Token); err != nil {
			log.Println(internal.ErrorPrefix, "logging out: ", err)
			switch {
			// This means that token is invalid anyway
			case errors.Is(err, core.ErrUnauthorized):
			case errors.Is(err, core.ErrBadRequest):
			// NordAccount tokens do not work with Logout endpoint and return ErrNotFound
			case errors.Is(err, core.ErrNotFound):
			case errors.Is(err, core.ErrServerInternal):
				return &pb.Payload{
					Type: internal.CodeInternalError,
				}, nil
			default:
				return &pb.Payload{
					Type: internal.CodeFailure,
				}, nil
			}
		}
	}

	if err := r.cm.SaveWith(func(c config.Config) config.Config {
		delete(c.TokensData, c.AutoConnectData.ID)
		c.AutoConnectData.ID = 0
		c.Mesh = false
		return c
	}); err != nil {
		return nil, err
	}

	if err := r.ncClient.Stop(); err != nil {
		log.Println(internal.WarningPrefix, err)
	}
	r.publisher.Publish("user logged out")

	// RenewToken being empty means user logged in using Access Token
	if !in.GetPersistToken() && tokenData.RenewToken == "" {
		return &pb.Payload{Type: internal.CodeTokenInvalidated}, nil
	}

	return &pb.Payload{Type: internal.CodeSuccess}, nil
}
