package core

import (
	"errors"
	"net/http"

	"github.com/NordSecurity/nordvpn-linux/core/mesh"
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

type smartClientAPI struct {
	wrapped  RawClientAPI
	tokenMan TokenManager
}

// NewSmartClientAPI creates new client instance of smart-API
func NewSmartClientAPI(client RawClientAPI, loginTokenMan TokenManager) ClientAPI {
	return &smartClientAPI{wrapped: client, tokenMan: loginTokenMan}
}

func callWithToken[T any](tokenMan TokenManager, call func(token string) (T, error)) (T, error) {
	callAPIWithToken := func() (T, error) {
		token, err := tokenMan.Token()
		if err != nil {
			var zero T
			return zero, err
		}
		return call(token)
	}

	res, err := callAPIWithToken()
	if err == nil {
		return res, nil
	}

	if errors.Is(err, ErrUnauthorized) {
		if err := tokenMan.Renew(); err != nil {
			var zero T
			return zero, err
		}

		res, err = callAPIWithToken()
	}

	if errors.Is(err, ErrUnauthorized) {
		tokenMan.Invalidate(ErrUnauthorized)
	}

	return res, err
}

func (s *smartClientAPI) NotificationCredentials(appUserID string) (NotificationCredentialsResponse, error) {
	return callWithToken(s.tokenMan, func(token string) (NotificationCredentialsResponse, error) {
		return s.wrapped.NotificationCredentials(token, appUserID)
	})
}

func (s *smartClientAPI) NotificationCredentialsRevoke(appUserID string, purgeSession bool) (NotificationCredentialsRevokeResponse, error) {
	return callWithToken(s.tokenMan, func(token string) (NotificationCredentialsRevokeResponse, error) {
		return s.wrapped.NotificationCredentialsRevoke(token, appUserID, purgeSession)
	})
}

func (s *smartClientAPI) ServiceCredentials(token string) (*CredentialsResponse, error) {
	return s.wrapped.ServiceCredentials(token)
}

func (s *smartClientAPI) TokenRenew(renewalToken string, idempotencyKey uuid.UUID) (*TokenRenewResponse, error) {
	return s.wrapped.TokenRenew(renewalToken, idempotencyKey)
}

func (s *smartClientAPI) Services() (ServicesResponse, error) {
	return callWithToken(s.tokenMan, func(token string) (ServicesResponse, error) {
		return s.wrapped.Services(token)
	})
}

func (s *smartClientAPI) CurrentUser() (*CurrentUserResponse, error) {
	return callWithToken(s.tokenMan, func(token string) (*CurrentUserResponse, error) {
		return s.wrapped.CurrentUser(token)
	})
}

func (s *smartClientAPI) DeleteToken() error {
	_, err := callWithToken(s.tokenMan, func(token string) (struct{}, error) {
		return struct{}{}, s.wrapped.DeleteToken(token)
	})
	return err
}

func (s *smartClientAPI) TrustedPassToken() (*TrustedPassTokenResponse, error) {
	return callWithToken(s.tokenMan, func(token string) (*TrustedPassTokenResponse, error) {
		return s.wrapped.TrustedPassToken(token)
	})
}

func (s *smartClientAPI) MultifactorAuthStatus() (*MultifactorAuthStatusResponse, error) {
	return callWithToken(s.tokenMan, func(token string) (*MultifactorAuthStatusResponse, error) {
		return s.wrapped.MultifactorAuthStatus(token)
	})
}

func (s *smartClientAPI) Logout() error {
	_, ret := callWithToken(s.tokenMan, func(token string) (struct{}, error) {
		return struct{}{}, s.wrapped.Logout(token)
	})
	return ret
}

func (s *smartClientAPI) Insights() (*Insights, error) {
	return s.wrapped.Insights()
}

func (s *smartClientAPI) Servers() (Servers, http.Header, error) {
	return s.wrapped.Servers()
}

func (s *smartClientAPI) RecommendedServers(filter ServersFilter, longitude, latitude float64) (Servers, http.Header, error) {
	return s.wrapped.RecommendedServers(filter, longitude, latitude)
}

func (s *smartClientAPI) Server(id int64) (*Server, error) {
	return s.wrapped.Server(id)
}

func (s *smartClientAPI) ServersCountries() (Countries, http.Header, error) {
	return s.wrapped.ServersCountries()
}

func (s *smartClientAPI) Base() string {
	return s.wrapped.Base()
}

func (s *smartClientAPI) Plans() (*Plans, error) {
	return s.wrapped.Plans()
}

func (s *smartClientAPI) CreateUser(email, password string) (*UserCreateResponse, error) {
	return s.wrapped.CreateUser(email, password)
}

func (s *smartClientAPI) Orders() ([]Order, error) {
	return callWithToken(s.tokenMan, func(token string) ([]Order, error) {
		return s.wrapped.Orders(token)
	})
}

func (s *smartClientAPI) Payments() ([]PaymentResponse, error) {
	return callWithToken(s.tokenMan, func(token string) ([]PaymentResponse, error) {
		return s.wrapped.Payments(token)
	})
}

// for mesh we dont do anything right now
func (s *smartClientAPI) Register(token string, peer mesh.Machine) (*mesh.Machine, error) {
	return s.wrapped.Register(token, peer)
}

func (s *smartClientAPI) Update(token string, id uuid.UUID, info mesh.MachineUpdateRequest) error {
	return s.wrapped.Update(token, id, info)
}

func (s *smartClientAPI) Configure(token string, id uuid.UUID, peerID uuid.UUID, peerUpdateInfo mesh.PeerUpdateRequest) error {
	return s.wrapped.Configure(token, id, peerID, peerUpdateInfo)
}

func (s *smartClientAPI) Unregister(token string, self uuid.UUID) error {
	return s.wrapped.Unregister(token, self)
}

func (s *smartClientAPI) Map(token string, self uuid.UUID) (*mesh.MachineMap, error) {
	return s.wrapped.Map(token, self)
}

func (s *smartClientAPI) Unpair(token string, self uuid.UUID, peer uuid.UUID) error {
	return s.wrapped.Unpair(token, self, peer)
}

func (s *smartClientAPI) Invite(
	token string,
	self uuid.UUID,
	email string,
	doIAllowInbound,
	doIAllowRouting,
	doIAllowLocalNetwork,
	doIAllowFileshare bool,
) error {
	return s.wrapped.Invite(token, self, email, doIAllowInbound, doIAllowRouting,
		doIAllowLocalNetwork, doIAllowFileshare)
}

func (s *smartClientAPI) Received(token string, self uuid.UUID) (mesh.Invitations, error) {
	return s.wrapped.Received(token, self)
}

func (s *smartClientAPI) Sent(token string, self uuid.UUID) (mesh.Invitations, error) {
	return s.wrapped.Sent(token, self)
}

func (s *smartClientAPI) Accept(
	token string,
	self uuid.UUID,
	invitation uuid.UUID,
	doIAllowInbound, doIAllowRouting, doIAllowLocalNetwork, doIAllowFileshare bool,
) error {
	return s.wrapped.Accept(token, self, invitation, doIAllowInbound,
		doIAllowRouting, doIAllowLocalNetwork, doIAllowFileshare)
}

func (s *smartClientAPI) Reject(token string, self uuid.UUID, invitation uuid.UUID) error {
	return s.wrapped.Reject(token, self, invitation)
}

func (s *smartClientAPI) Revoke(token string, self uuid.UUID, invitation uuid.UUID) error {
	return s.wrapped.Revoke(token, self, invitation)
}

func (s *smartClientAPI) NotifyNewTransfer(
	token string,
	self uuid.UUID,
	peer uuid.UUID,
	fileName string,
	fileCount int,
	transferID string,
) error {
	return s.wrapped.NotifyNewTransfer(token, self, peer, fileName, fileCount,
		transferID)
}
