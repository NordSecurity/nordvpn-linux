package core

import (
	"errors"
	"log"
	"net/http"
	"runtime"

	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/google/uuid"
)

type ClientAPI interface {
	CredentialsAPI
	InsightsAPI
	ServersAPI
	CombinedAPI
	SubscriptionAPI
	mesh.Mapper
	mesh.Registry
	mesh.Inviter
}

const (
	logTag = "[api]"
)

type SmartClientAPI struct {
	wrapped  *SimpleClientAPI
	tokenMan TokenManager
}

// callerName gets the called name for parent's parent
func callerName(hist int) string {
	pc, _, _, ok := runtime.Caller(hist)
	if !ok {
		return ""
	}
	return runtime.FuncForPC(pc).Name() + ":"
}

func NewSmartClientAPI(
	client *SimpleClientAPI,
	loginTokenMan TokenManager,
) ClientAPI {
	return &SmartClientAPI{
		wrapped:  client,
		tokenMan: loginTokenMan,
	}
}

func callWithToken[T any](tokenMan TokenManager, call func(token string) (T, error)) (T, error) {
	var zero T
	tryCalling := func() (T, error) {
		token, err := tokenMan.Token()
		if err != nil {
			return zero, err
		}
		log.Println("[smart-api]", internal.DebugPrefix, "Trying to call Smart API for", callerName(3))
		return call(token)
	}

	res, err := tryCalling()
	if err == nil {
		return res, nil
	}

	if errors.Is(err, ErrUnauthorized) {
		log.Println("[smart-api]", internal.DebugPrefix, "Got 'ErrUnauthorized' for", callerName(3), "will try to renew token")
		if err := tokenMan.Renew(); err != nil {
			log.Println("[smart-api]", internal.DebugPrefix, "renewing login token", err)
			return zero, err
		}

		res, err = tryCalling()
	}

	if errors.Is(err, ErrUnauthorized) {
		tokenMan.Invalidate(ErrUnauthorized)
	}

	return res, err
}

func (s *SmartClientAPI) NotificationCredentials(appUserID string) (NotificationCredentialsResponse, error) {
	return callWithToken(s.tokenMan, func(token string) (NotificationCredentialsResponse, error) {
		return s.wrapped.NotificationCredentials(token, appUserID)
	})
}

func (s *SmartClientAPI) NotificationCredentialsRevoke(appUserID string, purgeSession bool) (NotificationCredentialsRevokeResponse, error) {
	return callWithToken(s.tokenMan, func(token string) (NotificationCredentialsRevokeResponse, error) {
		return s.wrapped.NotificationCredentialsRevoke(token, appUserID, purgeSession)
	})
}

func (s *SmartClientAPI) ServiceCredentials(token string) (*CredentialsResponse, error) {
	return s.wrapped.ServiceCredentials(token)
}

func (s *SmartClientAPI) TokenRenew(renewalToken string, idempotencyKey uuid.UUID) (*TokenRenewResponse, error) {
	return s.wrapped.TokenRenew(renewalToken, idempotencyKey)
}

func (s *SmartClientAPI) Services() (ServicesResponse, error) {
	return callWithToken(s.tokenMan, func(token string) (ServicesResponse, error) {
		return s.wrapped.Services(token)
	})
}

func (s *SmartClientAPI) CurrentUser() (*CurrentUserResponse, error) {
	return callWithToken(s.tokenMan, func(token string) (*CurrentUserResponse, error) {
		return s.wrapped.CurrentUser(token)
	})
}

func (s *SmartClientAPI) DeleteToken() error {
	_, err := callWithToken(s.tokenMan, func(token string) (struct{}, error) {
		return struct{}{}, s.wrapped.DeleteToken(token)
	})
	return err
}

func (s *SmartClientAPI) TrustedPassToken() (*TrustedPassTokenResponse, error) {
	return callWithToken(s.tokenMan, func(token string) (*TrustedPassTokenResponse, error) {
		return s.wrapped.TrustedPassToken(token)
	})
}

func (s *SmartClientAPI) MultifactorAuthStatus() (*MultifactorAuthStatusResponse, error) {
	return callWithToken(s.tokenMan, func(token string) (*MultifactorAuthStatusResponse, error) {
		return s.wrapped.MultifactorAuthStatus(token)
	})
}

func (s *SmartClientAPI) Logout() error {
	_, ret := callWithToken(s.tokenMan, func(token string) (struct{}, error) {
		return struct{}{}, s.wrapped.Logout(token)
	})
	return ret
}

func (s *SmartClientAPI) Insights() (*Insights, error) {
	return s.wrapped.Insights()
}

func (s *SmartClientAPI) Servers() (Servers, http.Header, error) {
	return s.wrapped.Servers()
}

func (s *SmartClientAPI) RecommendedServers(filter ServersFilter, longitude, latitude float64) (Servers, http.Header, error) {
	return s.wrapped.RecommendedServers(filter, longitude, latitude)
}

func (s *SmartClientAPI) Server(id int64) (*Server, error) {
	return s.wrapped.Server(id)
}

func (s *SmartClientAPI) ServersCountries() (Countries, http.Header, error) {
	return s.wrapped.ServersCountries()
}

func (s *SmartClientAPI) Base() string {
	return s.wrapped.Base()
}

func (s *SmartClientAPI) Plans() (*Plans, error) {
	return s.wrapped.Plans()
}

func (s *SmartClientAPI) CreateUser(email, password string) (*UserCreateResponse, error) {
	return s.wrapped.CreateUser(email, password)
}

func (s *SmartClientAPI) Orders() ([]Order, error) {
	return callWithToken(s.tokenMan, func(token string) ([]Order, error) {
		return s.wrapped.Orders(token)
	})
}

func (s *SmartClientAPI) Payments() ([]PaymentResponse, error) {
	return callWithToken(s.tokenMan, func(token string) ([]PaymentResponse, error) {
		return s.wrapped.Payments(token)
	})
}

// for mesh we dont do anything right now
func (s *SmartClientAPI) Register(token string, peer mesh.Machine) (*mesh.Machine, error) {
	// return callWithToken(s.tokenMan, func(token string) (*mesh.Machine, error) {
	return s.wrapped.Register(token, peer)
	// })
}

func (s *SmartClientAPI) Update(token string, id uuid.UUID, info mesh.MachineUpdateRequest) error {
	// _, err := callWithToken(s.tokenMan, func(token string) (struct{}, error) {
	// return struct{}{}, s.wrapped.Update(token, id, info)
	// })
	// return err
	return s.wrapped.Update(token, id, info)
}

func (s *SmartClientAPI) Configure(token string, id uuid.UUID, peerID uuid.UUID, peerUpdateInfo mesh.PeerUpdateRequest) error {
	// _, err := callWithToken(s.tokenMan, func(token string) (struct{}, error) {
	// return struct{}{}, s.wrapped.Configure(token, id, peerID, peerUpdateInfo)
	// })
	// return err
	return s.wrapped.Configure(token, id, peerID, peerUpdateInfo)
}

func (s *SmartClientAPI) Unregister(token string, self uuid.UUID) error {
	// _, err := callWithToken(s.tokenMan, func(token string) (struct{}, error) {
	// return struct{}{}, s.wrapped.Unregister(token, self)
	// })
	// return err
	return s.wrapped.Unregister(token, self)
}

func (s *SmartClientAPI) Map(token string, self uuid.UUID) (*mesh.MachineMap, error) {
	// return callWithToken(s.tokenMan, func(token string) (*mesh.MachineMap, error) {
	return s.wrapped.Map(token, self)
	// })
}

func (s *SmartClientAPI) Unpair(token string, self uuid.UUID, peer uuid.UUID) error {
	// _, err := callWithToken(s.tokenMan, func(token string) (struct{}, error) {
	// return struct{}{}, s.wrapped.Unpair(token, self, peer)
	// })
	// return err
	return s.wrapped.Unpair(token, self, peer)
}

func (s *SmartClientAPI) Invite(
	token string,
	self uuid.UUID,
	email string,
	doIAllowInbound,
	doIAllowRouting,
	doIAllowLocalNetwork,
	doIAllowFileshare bool,
) error {
	// _, err := callWithToken(s.tokenMan, func(token string) (struct{}, error) {
	// 	return struct{}{}, s.wrapped.Invite(token, self, email, doIAllowInbound, doIAllowRouting,
	// 		doIAllowLocalNetwork, doIAllowFileshare)
	// })
	// return err
	return s.wrapped.Invite(token, self, email, doIAllowInbound, doIAllowRouting,
		doIAllowLocalNetwork, doIAllowFileshare)
}

func (s *SmartClientAPI) Received(token string, self uuid.UUID) (mesh.Invitations, error) {
	// return callWithToken(s.tokenMan, func(token string) (mesh.Invitations, error) {
	return s.wrapped.Received(token, self)
	// })
}

func (s *SmartClientAPI) Sent(token string, self uuid.UUID) (mesh.Invitations, error) {
	// return callWithToken(s.tokenMan, func(token string) (mesh.Invitations, error) {
	return s.wrapped.Sent(token, self)
	// })
}

func (s *SmartClientAPI) Accept(
	token string,
	self uuid.UUID,
	invitation uuid.UUID,
	doIAllowInbound, doIAllowRouting, doIAllowLocalNetwork, doIAllowFileshare bool,
) error {
	// _, err := callWithToken(s.tokenMan, func(token string) (struct{}, error) {
	// 	return struct{}{}, s.wrapped.Accept(token, self, invitation, doIAllowInbound,
	// 		doIAllowRouting, doIAllowLocalNetwork, doIAllowFileshare)
	// })
	// return err
	return s.wrapped.Accept(token, self, invitation, doIAllowInbound,
		doIAllowRouting, doIAllowLocalNetwork, doIAllowFileshare)
}

func (s *SmartClientAPI) Reject(token string, self uuid.UUID, invitation uuid.UUID) error {
	// _, err := callWithToken(s.tokenMan, func(token string) (struct{}, error) {
	// return struct{}{}, s.wrapped.Reject(token, self, invitation)
	// })
	// return err
	return s.wrapped.Reject(token, self, invitation)
}

func (s *SmartClientAPI) Revoke(token string, self uuid.UUID, invitation uuid.UUID) error {
	// _, err := callWithToken(s.tokenMan, func(token string) (struct{}, error) {
	// return struct{}{}, s.wrapped.Revoke(token, self, invitation)
	// })
	// return err
	return s.wrapped.Revoke(token, self, invitation)
}

func (s *SmartClientAPI) NotifyNewTransfer(
	token string,
	self uuid.UUID,
	peer uuid.UUID,
	fileName string,
	fileCount int,
	transferID string,
) error {
	// _, err := callWithToken(s.tokenMan, func(token string) (struct{}, error) {
	// 	return struct{}{}, s.wrapped.NotifyNewTransfer(token, self, peer, fileName, fileCount,
	// 		transferID)
	// })
	// return err
	return s.wrapped.NotifyNewTransfer(token, self, peer, fileName, fileCount,
		transferID)
}

// TODO: this one is dangling without usage
//  func (*DefaultAPI) List(token string, self uuid.UUID) (mesh.MachinePeers, error)
