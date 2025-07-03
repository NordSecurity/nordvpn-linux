package daemon

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/daemon/access"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
)

func (r *RPC) Logout(ctx context.Context, in *pb.LogoutRequest) (*pb.Payload, error) {
	result := access.Logout(access.LogoutInput{
		AuthChecker:    r.ac,
		CredentialsAPI: r.credentialsAPI,
		Netw:           r.netw,
		NcClient:       r.ncClient,
		ConfigManager:  r.cm,
		Events:         r.events,
		Publisher:      r.publisher,
		PersistToken:   in.GetPersistToken(),
		DisconnectAll:  r.DoDisconnect,
	})

	if result.Status == 0 {
		return nil, result.Err
	}

	return &pb.Payload{Type: result.Status}, result.Err
}

// Logout erases user credentials and disconnects completely
// func (r *RPC) Logout(ctx context.Context, in *pb.LogoutRequest) (payload *pb.Payload, retErr error) {
// 	if !r.ac.IsLoggedIn() {
// 		return nil, internal.ErrNotLoggedIn
// 	}

// 	logoutStartTime := time.Now()
// 	r.events.User.Logout.Publish(events.DataAuthorization{DurationMs: -1, EventTrigger: events.TriggerUser, EventStatus: events.StatusAttempt})

// 	defer func() {
// 		eventStatus := events.StatusSuccess
// 		if retErr != nil || payload != nil && payload.Type != internal.CodeSuccess && payload.Type != internal.CodeTokenInvalidated {
// 			eventStatus = events.StatusFailure
// 		}
// 		r.events.User.Logout.Publish(
// 			events.DataAuthorization{
// 				DurationMs:   max(int(time.Since(logoutStartTime).Milliseconds()), 1),
// 				EventTrigger: events.TriggerUser,
// 				EventStatus:  eventStatus},
// 		)
// 	}()

// 	log.Println("loading config")
// 	var cfg config.Config
// 	err := r.cm.Load(&cfg)
// 	if err != nil {
// 		log.Println(internal.ErrorPrefix, err)
// 		return &pb.Payload{Type: internal.CodeFailure}, nil
// 	}

// 	log.Println("performing disconnect")
// 	if _, err := r.DoDisconnect(); err != nil {
// 		log.Println(internal.ErrorPrefix, "Error while disconnecting", err)
// 		return &pb.Payload{Type: internal.CodeFailure}, nil
// 	}

// 	log.Println("unsetting mesh")
// 	if err := r.netw.UnSetMesh(); err != nil && !errors.Is(err, networker.ErrMeshNotActive) {
// 		log.Println(internal.ErrorPrefix, err)
// 		return &pb.Payload{Type: internal.CodeFailure}, nil
// 	}

// 	log.Println("stopping nc")
// 	if err := r.ncClient.Stop(); err != nil {
// 		log.Println(internal.WarningPrefix, err)
// 	}

// 	log.Println("reading access token")
// 	tokenData, ok := cfg.TokensData[cfg.AutoConnectData.ID]
// 	if !ok {
// 		return &pb.Payload{Type: internal.CodeFailure}, nil
// 	}

// 	if !r.ncClient.Revoke() {
// 		log.Println(internal.WarningPrefix, "error revoking NC token")
// 	}

// 	log.Println("deleting token")
// 	if !in.GetPersistToken() {
// 		if err := r.credentialsAPI.DeleteToken(tokenData.Token); err != nil {
// 			log.Println(internal.ErrorPrefix, "deleting token: ", err)
// 			switch {
// 			// This means that token is invalid anyway
// 			case errors.Is(err, core.ErrUnauthorized):
// 			case errors.Is(err, core.ErrBadRequest):
// 			case errors.Is(err, core.ErrServerInternal):
// 				return &pb.Payload{
// 					Type: internal.CodeInternalError,
// 				}, nil
// 			default:
// 				return &pb.Payload{
// 					Type: internal.CodeFailure,
// 				}, nil
// 			}
// 		}

// 		log.Println("loging out")
// 		if err := r.credentialsAPI.Logout(tokenData.Token); err != nil {
// 			log.Println(internal.ErrorPrefix, "logging out: ", err)
// 			switch {
// 			// This means that token is invalid anyway
// 			case errors.Is(err, core.ErrUnauthorized):
// 			case errors.Is(err, core.ErrBadRequest):
// 			// NordAccount tokens do not work with Logout endpoint and return ErrNotFound
// 			case errors.Is(err, core.ErrNotFound):
// 			case errors.Is(err, core.ErrServerInternal):
// 				return &pb.Payload{
// 					Type: internal.CodeInternalError,
// 				}, nil
// 			default:
// 				return &pb.Payload{
// 					Type: internal.CodeFailure,
// 				}, nil
// 			}
// 		}
// 	}

// 	log.Println("reseting config")
// 	if err := r.cm.SaveWith(func(c config.Config) config.Config {
// 		delete(c.TokensData, c.AutoConnectData.ID)
// 		c.AutoConnectData.ID = 0
// 		c.Mesh = false
// 		c.MeshPrivateKey = ""
// 		return c
// 	}); err != nil {
// 		return nil, err
// 	}

// 	r.publisher.Publish("user logged out")

// 	// RenewToken being empty means user logged in using Access Token
// 	if !in.GetPersistToken() && tokenData.RenewToken == "" {
// 		return &pb.Payload{Type: internal.CodeTokenInvalidated}, nil
// 	}

// 	return &pb.Payload{Type: internal.CodeSuccess}, nil
// }
