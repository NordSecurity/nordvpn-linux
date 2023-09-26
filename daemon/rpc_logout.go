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
	log.Printf("Start Logout RPC")
	if !r.ac.IsLoggedIn() {
		return nil, internal.ErrNotLoggedIn
	}

	log.Printf("Checking config")
	var cfg config.Config
	err := r.cm.Load(&cfg)
	if err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.Payload{Type: internal.CodeFailure}, nil
	}

	log.Printf("Disabling fileshare")
	if err := r.fileshare.Disable(cfg.Meshnet.EnabledByUID, cfg.Meshnet.EnabledByGID); err != nil {
		log.Println(internal.ErrorPrefix, "disabling fileshare: ", err)
	}

	log.Printf("Stopping networker")
	if err := r.netw.Stop(); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.Payload{Type: internal.CodeFailure}, nil
	}

	log.Printf("Unsetting mesh")
	if err := r.netw.UnSetMesh(); err != nil && !errors.Is(err, networker.ErrMeshNotActive) {
		log.Println(internal.ErrorPrefix, err)
		return &pb.Payload{Type: internal.CodeFailure}, nil
	}

	tokenData, ok := cfg.TokensData[cfg.AutoConnectData.ID]
	if !ok {
		return &pb.Payload{Type: internal.CodeFailure}, nil
	}

	log.Printf("Before persist token check")

	if !in.GetPersistToken() {
		log.Printf("Deleting token")
		if err := r.api.DeleteToken(tokenData.Token); err != nil {
			log.Println(internal.ErrorPrefix, "deleting token: ", err)
			switch err {
			// This means that token is invalid anyway
			case core.ErrUnauthorized:
			case core.ErrBadRequest:
			case core.ErrServerInternal:
				return &pb.Payload{
					Type: internal.CodeInternalError,
				}, nil
			default:
				return &pb.Payload{
					Type: internal.CodeFailure,
				}, nil
			}
		}

		log.Printf("Logging out")
		if err := r.api.Logout(tokenData.Token); err != nil {
			log.Println(internal.ErrorPrefix, "logging out: ", err)
			switch err {
			// This means that token is invalid anyway
			case core.ErrUnauthorized:
			case core.ErrBadRequest:
			case core.ErrServerInternal:
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

	log.Printf("After API calls")
	if err := r.cm.SaveWith(func(c config.Config) config.Config {
		delete(c.TokensData, c.AutoConnectData.ID)
		c.AutoConnectData.ID = 0
		c.Mesh = false
		return c
	}); err != nil {
		return nil, err
	}

	log.Printf("Stopping nc client")
	if err := r.ncClient.Stop(); err != nil {
		log.Println(internal.WarningPrefix, err)
	}
	r.publisher.Publish("user logged out")

	log.Printf("Before returns")
	// RenewToken being empty means user logged in using Access Token
	if !in.GetPersistToken() && tokenData.RenewToken == "" {
		return &pb.Payload{Type: internal.CodeTokenInvalidated}, nil
	}

	log.Printf("Returning success")
	return &pb.Payload{Type: internal.CodeSuccess}, nil
}
