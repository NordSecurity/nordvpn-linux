package daemon

import (
	"context"
	"errors"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/internal/caching"
)

// ErrMissingExchangeToken is returned when login was successful but
// there is not enough data to request the token
var ErrMissingExchangeToken = errors.New("exchange token not provided")

type customCallbackType func() (*core.LoginResponse, *pb.LoginResponse, error)

var isTokenValid = regexp.MustCompile(`^[a-f0-9]*$`).MatchString

var lastLoginAttemptTime time.Time

// Login the user with given token
func (r *RPC) LoginWithToken(ctx context.Context, in *pb.LoginWithTokenRequest) (*pb.LoginResponse, error) {
	if !r.consentChecker.IsConsentFlowCompleted() {
		return &pb.LoginResponse{
			Type: internal.CodeConsentMissing,
		}, nil
	}

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
				ExpiresAt: core.ManualLoginTokenExpiryDate,
			}, nil, nil
		}
		return nil, &pb.LoginResponse{
			Type: internal.CodeTokenLoginFailure,
		}, nil
	})
}

// loginCommon common login
func (r *RPC) loginCommon(customCB customCallbackType) (payload *pb.LoginResponse, retErr error) {
	if r.ac.IsLoggedIn() {
		return nil, internal.ErrAlreadyLoggedIn
	}

	loginStartTime := time.Now()
	r.events.User.Login.Publish(events.DataAuthorization{
		DurationMs:   -1,
		EventTrigger: events.TriggerUser,
		EventStatus:  events.StatusAttempt,
		EventType:    events.LoginLogin,
	})

	defer func() {
		eventStatus := events.StatusSuccess
		if retErr != nil || payload != nil && payload.Type != internal.CodeSuccess {
			eventStatus = events.StatusFailure
		}
		r.events.User.Login.Publish(events.DataAuthorization{
			DurationMs:   max(int(time.Since(loginStartTime).Milliseconds()), 1),
			EventTrigger: events.TriggerUser,
			EventStatus:  eventStatus,
			EventType:    events.LoginLogin,
		})
	}()

	resp, pbresp, err := customCB()
	if err != nil || pbresp != nil {
		return pbresp, nil
	}

	credentials, err := r.credentialsAPI.ServiceCredentials(resp.Token)
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

	// get user's current mfa status (should be invoked after config with creds are saved)
	// (errors are checked and logged inside this fn, if any)
	r.ac.IsMFAEnabled()

	go StartNC("[login]", r.ncClient)
	r.publisher.Publish("user logged in")

	return &pb.LoginResponse{
		Type: internal.CodeSuccess,
	}, nil
}

// LoginOAuth2 is called when logging in with OAuth2.
func (r *RPC) LoginOAuth2(ctx context.Context, in *pb.LoginOAuth2Request) (*pb.LoginOAuth2Response, error) {
	if !r.consentChecker.IsConsentFlowCompleted() {
		return &pb.LoginOAuth2Response{
			Status: pb.LoginStatus_CONSENT_MISSING,
		}, nil
	}

	if r.ac.IsLoggedIn() {
		return &pb.LoginOAuth2Response{
			Status: pb.LoginStatus_ALREADY_LOGGED_IN,
		}, nil
	}

	lastLoginAttemptTime = time.Now()

	eventType := events.LoginLogin
	if in.GetType() == pb.LoginType_LoginType_SIGNUP {
		eventType = events.LoginSignUp
	}

	r.events.User.Login.Publish(events.DataAuthorization{
		DurationMs:   -1,
		EventTrigger: events.TriggerUser,
		EventStatus:  events.StatusAttempt,
		EventType:    eventType,
	})

	url, err := r.authentication.Login(in.GetType() == pb.LoginType_LoginType_LOGIN)
	if err != nil {
		if strings.Contains(err.Error(), "network is unreachable") ||
			strings.Contains(err.Error(), "Client.Timeout exceeded while awaiting headers") {
			return &pb.LoginOAuth2Response{
				Status: pb.LoginStatus_NO_NET,
			}, nil
		}

		return &pb.LoginOAuth2Response{
			Status: pb.LoginStatus_UNKNOWN_OAUTH2_ERROR,
		}, nil
	}

	return &pb.LoginOAuth2Response{
		Status: pb.LoginStatus_SUCCESS,
		Url:    url,
	}, nil
}

// LoginOAuth2Callback is called by the browser via cli during OAuth2 login.
func (r *RPC) LoginOAuth2Callback(ctx context.Context, in *pb.LoginOAuth2CallbackRequest) (payload *pb.LoginOAuth2CallbackResponse, retErr error) {
	if !r.consentChecker.IsConsentFlowCompleted() {
		return &pb.LoginOAuth2CallbackResponse{
			Status: pb.LoginStatus_CONSENT_MISSING,
		}, nil
	}
	if r.ac.IsLoggedIn() {
		return nil, internal.ErrAlreadyLoggedIn
	}

	loginType := events.LoginLogin
	if in.GetType() == pb.LoginType_LoginType_SIGNUP {
		loginType = events.LoginSignUp
	}

	defer func() {
		eventStatus := events.StatusSuccess
		if retErr != nil {
			eventStatus = events.StatusFailure
		}
		r.events.User.Login.Publish(events.DataAuthorization{
			DurationMs:   max(int(time.Since(lastLoginAttemptTime).Milliseconds()), 1),
			EventTrigger: events.TriggerUser,
			EventStatus:  eventStatus,
			EventType:    loginType,
		})
		lastLoginAttemptTime = time.Time{}
	}()

	if in.GetToken() == "" {
		r.publisher.Publish(ErrMissingExchangeToken.Error())
		return nil, ErrMissingExchangeToken
	}

	resp, err := r.authentication.Token(in.GetToken())
	if err != nil {
		r.publisher.Publish(err.Error())
		return nil, err
	}

	credentials, err := r.credentialsAPI.ServiceCredentials(resp.Token)
	if err != nil {
		r.publisher.Publish(err.Error())
		return nil, err
	}

	if err := r.cm.SaveWith(func(c config.Config) config.Config {
		c.TokensData[credentials.ID] = config.TokenData{
			Token:              resp.Token,
			RenewToken:         resp.RenewToken,
			TokenExpiry:        resp.ExpiresAt,
			NordLynxPrivateKey: credentials.NordlynxPrivateKey,
			OpenVPNUsername:    credentials.Username,
			OpenVPNPassword:    credentials.Password,
			IsOAuth:            true,
		}
		c.AutoConnectData.ID = credentials.ID
		return c
	}); err != nil {
		return nil, err
	}

	// get user's current mfa status (should be invoked after config with creds are saved)
	// (errors are checked and logged inside this fn, if any)
	r.ac.IsMFAEnabled()

	go StartNC("[login callback]", r.ncClient)
	return &pb.LoginOAuth2CallbackResponse{
		Status: pb.LoginStatus_SUCCESS,
	}, nil
}

var (
	isLoggedInCache     *caching.Cache[bool]
	isLoggedInCacheInit sync.Once
	isLoggedInCacheTTL  = time.Second * 9
)

func (r *RPC) IsLoggedIn(ctx context.Context, _ *pb.Empty) (*pb.IsLoggedInResponse, error) {
	if !r.consentChecker.IsConsentFlowCompleted() {
		return &pb.IsLoggedInResponse{Status: pb.LoginStatus_CONSENT_MISSING}, nil
	}

	// create cache on first call
	isLoggedInCacheInit.Do(func() {
		isLoggedInCache = caching.NewCacheWithTTL(
			isLoggedInCacheTTL,
			func() (bool, error) { return r.ac.IsLoggedIn(), nil },
		)
	})
	loggedIn, _ := isLoggedInCache.Get()
	return &pb.IsLoggedInResponse{IsLoggedIn: loggedIn}, nil
}
