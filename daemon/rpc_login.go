package daemon

import (
	"context"
	"errors"
	"log"
	"regexp"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

var (
	// ErrMissingExchangeToken is returned when login was successful but
	// there is not enough data to request the token
	ErrMissingExchangeToken = errors.New("exchange token not provided")
)

type customCallbackType func() (*core.LoginResponse, *pb.LoginResponse, error)

// Login the user
func (r *RPC) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	// login common with custom logic
	return r.loginCommon(func() (*core.LoginResponse, *pb.LoginResponse, error) {
		resp, err := r.api.Login(in.GetUsername(), in.GetPassword())
		if err != nil {
			log.Println(internal.ErrorPrefix, "logging in:", err)
			switch {
			case errors.Is(err, core.ErrBadRequest):
				return nil, &pb.LoginResponse{
					Type: internal.CodeBadRequest,
				}, nil
			case errors.Is(err, core.ErrUnauthorized):
				return nil, &pb.LoginResponse{
					Type: internal.CodeUnauthorized,
				}, nil
			default:
				return nil, &pb.LoginResponse{
					Type: internal.CodeGatewayError,
				}, nil
			}
		}
		return resp, nil, nil
	})
}

var isTokenValid = regexp.MustCompile(`^[a-f0-9]*$`).MatchString

// Login the user with given token
func (r *RPC) LoginWithToken(ctx context.Context, in *pb.LoginWithTokenRequest) (*pb.LoginResponse, error) {
	if !isTokenValid(in.GetToken()) {
		return &pb.LoginResponse{
			Type: internal.CodeTokenInvalid,
		}, nil
	}
	// login common with custom logic
	return r.loginCommon(func() (*core.LoginResponse, *pb.LoginResponse, error) {
		if in.GetToken() != "" {
			return &core.LoginResponse{
				Token:     in.GetToken(),
				ExpiresAt: time.Now().AddDate(0, 1, 0).Format(internal.ServerDateFormat),
			}, nil, nil
		}
		return nil, &pb.LoginResponse{
			Type: internal.CodeTokenLoginFailure,
		}, nil
	})
}

// loginCommon common login
func (r *RPC) loginCommon(customCB customCallbackType) (*pb.LoginResponse, error) {
	if r.ac.IsLoggedIn() {
		return nil, internal.ErrAlreadyLoggedIn
	}

	resp, pbresp, err := customCB()
	if err != nil || pbresp != nil {
		return pbresp, nil
	}

	credentials, err := r.api.ServiceCredentials(resp.Token)
	if err != nil {
		log.Println(internal.ErrorPrefix, "retrieving credentials:", err)
		if errors.Is(err, core.ErrServerInternal) {
			return &pb.LoginResponse{
				Type: internal.CodeInternalError,
			}, nil
		}
		if errors.Is(err, core.ErrUnauthorized) {
			return &pb.LoginResponse{
				Type: internal.CodeTokenInvalid,
			}, nil
		}
		return &pb.LoginResponse{
			Type: internal.CodeGatewayError,
		}, nil
	}

	var cfg config.Config
	if err := r.cm.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.LoginResponse{
			Type: internal.CodeConfigError,
		}, nil
	}

	if err := r.cm.SaveWith(func(c config.Config) config.Config {
		c.TokensData[credentials.ID] = config.TokenData{
			Token:              resp.Token,
			RenewToken:         resp.RenewToken,
			TokenExpiry:        resp.ExpiresAt,
			NordLynxPrivateKey: credentials.NordlynxPrivateKey,
			OpenVPNUsername:    credentials.Username,
			OpenVPNPassword:    credentials.Password,
		}
		c.AutoConnectData.ID = credentials.ID
		return c
	}); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return &pb.LoginResponse{
			Type: internal.CodeConfigError,
		}, nil
	}

	go StartNotificationCenter(r.api, r.ncClient, r.cm)

	r.events.Service.Login.Publish(nil)
	r.publisher.Publish("user logged in")

	return &pb.LoginResponse{
		Type: internal.CodeSuccess,
	}, nil
}

// LoginOAuth2 is called when logging in with OAuth2.
func (r *RPC) LoginOAuth2(in *pb.Empty, srv pb.Daemon_LoginOAuth2Server) error {
	if r.ac.IsLoggedIn() {
		return internal.ErrAlreadyLoggedIn
	}

	url, err := r.authentication.Login()
	if err != nil {
		return err
	}

	return srv.Send(&pb.String{Data: url})
}

// LoginOAuth2Callback is called by the browser via cli during OAuth2 login.
func (r *RPC) LoginOAuth2Callback(ctx context.Context, in *pb.String) (*pb.Empty, error) {
	if r.ac.IsLoggedIn() {
		return &pb.Empty{}, internal.ErrAlreadyLoggedIn
	}

	if in.GetData() == "" {
		r.publisher.Publish(ErrMissingExchangeToken.Error())
		return &pb.Empty{}, ErrMissingExchangeToken
	}

	resp, err := r.authentication.Token(in.GetData())
	if err != nil {
		r.publisher.Publish(err.Error())
		return &pb.Empty{}, err
	}

	credentials, err := r.api.ServiceCredentials(resp.Token)
	if err != nil {
		r.publisher.Publish(err.Error())
		return &pb.Empty{}, err
	}

	if err := r.cm.SaveWith(func(c config.Config) config.Config {
		c.TokensData[credentials.ID] = config.TokenData{
			Token:              resp.Token,
			RenewToken:         resp.RenewToken,
			TokenExpiry:        resp.ExpiresAt,
			NordLynxPrivateKey: credentials.NordlynxPrivateKey,
			OpenVPNUsername:    credentials.Username,
			OpenVPNPassword:    credentials.Password,
		}
		c.AutoConnectData.ID = credentials.ID
		return c
	}); err != nil {
		return &pb.Empty{}, err
	}

	go StartNotificationCenter(r.api, r.ncClient, r.cm)
	r.events.Service.Login.Publish(nil)
	return &pb.Empty{}, nil
}

func (r *RPC) IsLoggedIn(ctx context.Context, _ *pb.Empty) (*pb.Bool, error) {
	return &pb.Bool{Value: r.ac.IsLoggedIn()}, nil
}
