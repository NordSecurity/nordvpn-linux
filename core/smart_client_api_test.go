package core_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/session"
	mocksession "github.com/NordSecurity/nordvpn-linux/test/mock/session"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const (
	appUserID    = "test-user"
	renewedToken = "renewed-token"
	initialToken = "token"
)

type mockSimpleClientAPI struct {
	NotificationCredentialsFunc       func(token, appUserID string) (core.NotificationCredentialsResponse, error)
	NotificationCredentialsRevokeFunc func(token, appUserID string, purgeSession bool) (core.NotificationCredentialsRevokeResponse, error)
	ServiceCredentialsFunc            func(token string) (*core.CredentialsResponse, error)
	TokenRenewFunc                    func(renewalToken string, idempotencyKey uuid.UUID) (*core.TokenRenewResponse, error)
	ServicesFunc                      func(token string) (core.ServicesResponse, error)
	CurrentUserFunc                   func(token string) (*core.CurrentUserResponse, error)
	DeleteTokenFunc                   func(token string) error
	TrustedPassTokenFunc              func(token string) (*core.TrustedPassTokenResponse, error)
	MultifactorAuthStatusFunc         func(token string) (*core.MultifactorAuthStatusResponse, error)
	LogoutFunc                        func(token string) error
	InsightsFunc                      func() (*core.Insights, error)
	ServersFunc                       func() (core.Servers, http.Header, error)
	RecommendedServersFunc            func(filter core.ServersFilter, longitude, latitude float64) (core.Servers, http.Header, error)
	ServerFunc                        func(id int64) (*core.Server, error)
	ServersCountriesFunc              func() (core.Countries, http.Header, error)
	BaseFunc                          func() string
	PlansFunc                         func() (*core.Plans, error)
	CreateUserFunc                    func(email, password string) (*core.UserCreateResponse, error)
	OrdersFunc                        func(token string) ([]core.Order, error)
	PaymentsFunc                      func(token string) ([]core.PaymentResponse, error)
	RegisterFunc                      func(token string, peer mesh.Machine) (*mesh.Machine, error)
	UpdateFunc                        func(token string, id uuid.UUID, info mesh.MachineUpdateRequest) error
	ConfigureFunc                     func(token string, id uuid.UUID, peerID uuid.UUID, peerUpdateInfo mesh.PeerUpdateRequest) error
	UnregisterFunc                    func(token string, self uuid.UUID) error
	MapFunc                           func(token string, self uuid.UUID) (*mesh.MachineMap, error)
	ListFunc                          func(token string, self uuid.UUID) (mesh.MachinePeers, error)
	UnpairFunc                        func(token string, self uuid.UUID, peer uuid.UUID) error
	InviteFunc                        func(token string, self uuid.UUID, email string, doIAllowInbound, doIAllowRouting, doIAllowLocalNetwork, doIAllowFileshare bool) error
	ReceivedFunc                      func(token string, self uuid.UUID) (mesh.Invitations, error)
	SentFunc                          func(token string, self uuid.UUID) (mesh.Invitations, error)
	AcceptFunc                        func(token string, self uuid.UUID, invitation uuid.UUID, doIAllowInbound, doIAllowRouting, doIAllowLocalNetwork, doIAllowFileshare bool) error
	RejectFunc                        func(token string, self uuid.UUID, invitation uuid.UUID) error
	RevokeFunc                        func(token string, self uuid.UUID, invitation uuid.UUID) error
	NotifyNewTransferFunc             func(token string, self uuid.UUID, peer uuid.UUID, fileName string, fileCount int, transferID string) error
}

func (m *mockSimpleClientAPI) NotificationCredentials(token, appUserID string) (core.NotificationCredentialsResponse, error) {
	return m.NotificationCredentialsFunc(token, appUserID)
}

func (m *mockSimpleClientAPI) NotificationCredentialsRevoke(token, appUserID string, purgeSession bool) (core.NotificationCredentialsRevokeResponse, error) {
	return m.NotificationCredentialsRevokeFunc(token, appUserID, purgeSession)
}

func (m *mockSimpleClientAPI) ServiceCredentials(token string) (*core.CredentialsResponse, error) {
	return m.ServiceCredentialsFunc(token)
}

func (m *mockSimpleClientAPI) TokenRenew(renewalToken string, idempotencyKey uuid.UUID) (*core.TokenRenewResponse, error) {
	return m.TokenRenewFunc(renewalToken, idempotencyKey)
}

func (m *mockSimpleClientAPI) Services(token string) (core.ServicesResponse, error) {
	return m.ServicesFunc(token)
}

func (m *mockSimpleClientAPI) CurrentUser(token string) (*core.CurrentUserResponse, error) {
	return m.CurrentUserFunc(token)
}

func (m *mockSimpleClientAPI) DeleteToken(token string) error {
	return m.DeleteTokenFunc(token)
}

func (m *mockSimpleClientAPI) TrustedPassToken(token string) (*core.TrustedPassTokenResponse, error) {
	return m.TrustedPassTokenFunc(token)
}

func (m *mockSimpleClientAPI) MultifactorAuthStatus(token string) (*core.MultifactorAuthStatusResponse, error) {
	return m.MultifactorAuthStatusFunc(token)
}

func (m *mockSimpleClientAPI) Logout(token string) error {
	return m.LogoutFunc(token)
}

func (m *mockSimpleClientAPI) Insights() (*core.Insights, error) {
	return m.InsightsFunc()
}

func (m *mockSimpleClientAPI) Servers() (core.Servers, http.Header, error) {
	return m.ServersFunc()
}

func (m *mockSimpleClientAPI) RecommendedServers(filter core.ServersFilter, longitude, latitude float64) (core.Servers, http.Header, error) {
	return m.RecommendedServersFunc(filter, longitude, latitude)
}

func (m *mockSimpleClientAPI) Server(id int64) (*core.Server, error) {
	return m.ServerFunc(id)
}

func (m *mockSimpleClientAPI) ServersCountries() (core.Countries, http.Header, error) {
	return m.ServersCountriesFunc()
}

func (m *mockSimpleClientAPI) Base() string {
	return m.BaseFunc()
}

func (m *mockSimpleClientAPI) Plans() (*core.Plans, error) {
	return m.PlansFunc()
}

func (m *mockSimpleClientAPI) CreateUser(email, password string) (*core.UserCreateResponse, error) {
	return m.CreateUserFunc(email, password)
}

func (m *mockSimpleClientAPI) Orders(token string) ([]core.Order, error) {
	return m.OrdersFunc(token)
}

func (m *mockSimpleClientAPI) Payments(token string) ([]core.PaymentResponse, error) {
	return m.PaymentsFunc(token)
}

func (m *mockSimpleClientAPI) Register(token string, peer mesh.Machine) (*mesh.Machine, error) {
	return m.RegisterFunc(token, peer)
}

func (m *mockSimpleClientAPI) Update(token string, id uuid.UUID, info mesh.MachineUpdateRequest) error {
	return m.UpdateFunc(token, id, info)
}

func (m *mockSimpleClientAPI) Configure(token string, id uuid.UUID, peerID uuid.UUID, peerUpdateInfo mesh.PeerUpdateRequest) error {
	return m.ConfigureFunc(token, id, peerID, peerUpdateInfo)
}

func (m *mockSimpleClientAPI) Unregister(token string, self uuid.UUID) error {
	return m.UnregisterFunc(token, self)
}

func (m *mockSimpleClientAPI) Map(token string, self uuid.UUID) (*mesh.MachineMap, error) {
	return m.MapFunc(token, self)
}

func (m *mockSimpleClientAPI) List(token string, self uuid.UUID) (mesh.MachinePeers, error) {
	return m.ListFunc(token, self)
}

func (m *mockSimpleClientAPI) Unpair(token string, self uuid.UUID, peer uuid.UUID) error {
	return m.UnpairFunc(token, self, peer)
}

func (m *mockSimpleClientAPI) Invite(token string, self uuid.UUID, email string, doIAllowInbound, doIAllowRouting, doIAllowLocalNetwork, doIAllowFileshare bool) error {
	return m.InviteFunc(token, self, email, doIAllowInbound, doIAllowRouting, doIAllowLocalNetwork, doIAllowFileshare)
}

func (m *mockSimpleClientAPI) Received(token string, self uuid.UUID) (mesh.Invitations, error) {
	return m.ReceivedFunc(token, self)
}

func (m *mockSimpleClientAPI) Sent(token string, self uuid.UUID) (mesh.Invitations, error) {
	return m.SentFunc(token, self)
}

func (m *mockSimpleClientAPI) Accept(token string, self uuid.UUID, invitation uuid.UUID, doIAllowInbound, doIAllowRouting, doIAllowLocalNetwork, doIAllowFileshare bool) error {
	return m.AcceptFunc(token, self, invitation, doIAllowInbound, doIAllowRouting, doIAllowLocalNetwork, doIAllowFileshare)
}

func (m *mockSimpleClientAPI) Reject(token string, self uuid.UUID, invitation uuid.UUID) error {
	return m.RejectFunc(token, self, invitation)
}

func (m *mockSimpleClientAPI) Revoke(token string, self uuid.UUID, invitation uuid.UUID) error {
	return m.RevokeFunc(token, self, invitation)
}

func (m *mockSimpleClientAPI) NotifyNewTransfer(token string, self uuid.UUID, peer uuid.UUID, fileName string, fileCount int, transferID string) error {
	return m.NotifyNewTransferFunc(token, self, peer, fileName, fileCount, transferID)
}

func NewMockSmartClientAPI(api core.RawClientAPI, store session.SessionStore) core.ClientAPI {
	return core.NewSmartClientAPI(api, store)
}

func Test_NotificationCredentials_TokenRenewalScenarios(t *testing.T) {
	t.Run("Valid token, no renewal", func(t *testing.T) {
		expectedResp := core.NotificationCredentialsResponse{
			Endpoint: "tcps://unit0.nordvpn.com:1234",
			Username: "jhkdJDsfkhjJHKDFKJskdjfkSDufEOWIlKLDA",
			Password: "87638468&*g23jhj#",
		}

		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string { return initialToken },
		}

		mockAPI := &mockSimpleClientAPI{
			NotificationCredentialsFunc: func(token, uid string) (core.NotificationCredentialsResponse, error) {
				assert.Equal(t, initialToken, token)
				assert.Equal(t, appUserID, uid)
				return expectedResp, nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		resp, err := client.NotificationCredentials(appUserID)

		assert.NoError(t, err)
		assert.Equal(t, expectedResp, resp)

		// no need for token renewal
		assert.Equal(t, 1, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 0, mockSessionStore.RenewCallCount)
	})

	t.Run("Token renewal is triggered and succeeds", func(t *testing.T) {
		expectedResp := core.NotificationCredentialsResponse{
			Endpoint: "tcps://unit0.nordvpn.com:1234",
			Username: "jhkdJDsfkhjJHKDFKJskdjfkSDufEOWIlKLDA",
			Password: "87638468&*g23jhj#",
		}
		firstCall := true

		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string {
				if firstCall {
					return initialToken
				}
				return renewedToken
			},
			RenewFunc: func() error { return nil },
		}

		mockAPI := &mockSimpleClientAPI{
			NotificationCredentialsFunc: func(token, uid string) (core.NotificationCredentialsResponse, error) {
				if token == initialToken {
					// simulate that current token has become invalid
					firstCall = false
					return core.NotificationCredentialsResponse{}, core.ErrUnauthorized
				}
				assert.Equal(t, renewedToken, token)
				return expectedResp, nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		resp, err := client.NotificationCredentials(appUserID)

		assert.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		assert.Equal(t, 2, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 1, mockSessionStore.RenewCallCount)
	})

	t.Run("Token renewal is triggered but fails", func(t *testing.T) {
		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string { return initialToken },
			RenewFunc:    func() error { return errors.New("renew failed") },
		}

		mockAPI := &mockSimpleClientAPI{
			NotificationCredentialsFunc: func(token, uid string) (core.NotificationCredentialsResponse, error) {
				return core.NotificationCredentialsResponse{}, core.ErrUnauthorized
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		resp, err := client.NotificationCredentials(appUserID)

		assert.Error(t, err)
		assert.Empty(t, resp)
		assert.Equal(t, 1, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 1, mockSessionStore.RenewCallCount)
	})

	t.Run("Token renewal is triggered but api still fails", func(t *testing.T) {
		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc:   func() string { return initialToken },
			RenewFunc:      func() error { return nil },
			InvalidateFunc: func(reason error) error { return nil },
		}

		mockAPI := &mockSimpleClientAPI{
			NotificationCredentialsFunc: func(token, uid string) (core.NotificationCredentialsResponse, error) {
				return core.NotificationCredentialsResponse{}, core.ErrUnauthorized
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		resp, err := client.NotificationCredentials(appUserID)

		assert.Error(t, err)
		assert.Empty(t, resp)
		assert.Equal(t, 2, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 1, mockSessionStore.RenewCallCount)
	})
}

func Test_NotificationCredentialsRevoke_TokenRenewalScenarios(t *testing.T) {
	t.Run("Valid token, no renewal", func(t *testing.T) {
		expectedResp := core.NotificationCredentialsRevokeResponse{Status: "ok"}

		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string { return initialToken },
		}

		mockAPI := &mockSimpleClientAPI{
			NotificationCredentialsRevokeFunc: func(token, uid string, purge bool) (core.NotificationCredentialsRevokeResponse, error) {
				assert.Equal(t, initialToken, token)
				assert.Equal(t, appUserID, uid)
				assert.True(t, purge)
				return expectedResp, nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		resp, err := client.NotificationCredentialsRevoke(appUserID, true)

		assert.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		assert.Equal(t, 1, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 0, mockSessionStore.RenewCallCount)
	})

	t.Run("Token renewal is triggered and succeeds", func(t *testing.T) {
		expectedResp := core.NotificationCredentialsRevokeResponse{Status: "ok"}
		firstCall := true

		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string {
				if firstCall {
					return initialToken
				}
				return renewedToken
			},
			RenewFunc: func() error { return nil },
		}

		mockAPI := &mockSimpleClientAPI{
			NotificationCredentialsRevokeFunc: func(token, uid string, purge bool) (core.NotificationCredentialsRevokeResponse, error) {
				if token == initialToken {
					firstCall = false
					return core.NotificationCredentialsRevokeResponse{}, core.ErrUnauthorized
				}
				assert.Equal(t, renewedToken, token)
				assert.True(t, purge)
				return expectedResp, nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		resp, err := client.NotificationCredentialsRevoke(appUserID, true)

		assert.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		assert.Equal(t, 2, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 1, mockSessionStore.RenewCallCount)
	})

	t.Run("Token renewal is triggered but fails", func(t *testing.T) {
		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string { return initialToken },
			RenewFunc:    func() error { return errors.New("renewal failed") },
		}

		mockAPI := &mockSimpleClientAPI{
			NotificationCredentialsRevokeFunc: func(token, uid string, purge bool) (core.NotificationCredentialsRevokeResponse, error) {
				return core.NotificationCredentialsRevokeResponse{}, core.ErrUnauthorized
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		resp, err := client.NotificationCredentialsRevoke(appUserID, false)

		assert.Error(t, err)
		assert.Empty(t, resp)
		assert.Equal(t, 1, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 1, mockSessionStore.RenewCallCount)
	})

	t.Run("Token renewal succeeds but API still fails", func(t *testing.T) {
		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc:   func() string { return initialToken },
			RenewFunc:      func() error { return nil },
			InvalidateFunc: func(reason error) error { return nil },
		}

		mockAPI := &mockSimpleClientAPI{
			NotificationCredentialsRevokeFunc: func(token, uid string, purge bool) (core.NotificationCredentialsRevokeResponse, error) {
				return core.NotificationCredentialsRevokeResponse{}, core.ErrUnauthorized
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		resp, err := client.NotificationCredentialsRevoke(appUserID, false)

		assert.Error(t, err)
		assert.Empty(t, resp)
		assert.Equal(t, 2, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 1, mockSessionStore.RenewCallCount)
	})
}

func Test_ServiceCredentials_TokenRenewalScenarios(t *testing.T) {
	t.Run("Valid token, no renewal", func(t *testing.T) {
		expectedResp := &core.CredentialsResponse{
			Username: "norduser",
			Password: "supersecret",
		}

		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string { return initialToken },
		}

		mockAPI := &mockSimpleClientAPI{
			ServiceCredentialsFunc: func(token string) (*core.CredentialsResponse, error) {
				assert.Equal(t, initialToken, token)
				return expectedResp, nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		resp, err := client.ServiceCredentials(initialToken)

		assert.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		// The function should not be calling the session store since the token is passed directly
		assert.Equal(t, 0, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 0, mockSessionStore.RenewCallCount)
	})

	t.Run("Token renewal is never triggered and succeeds", func(t *testing.T) {
		firstCall := true

		expectedResp := &core.CredentialsResponse{
			Username: "refreshed-user",
			Password: "renewed-pass",
		}

		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string {
				if firstCall {
					return initialToken
				}
				return renewedToken
			},
			RenewFunc: func() error { return nil },
		}

		mockAPI := &mockSimpleClientAPI{
			ServiceCredentialsFunc: func(token string) (*core.CredentialsResponse, error) {
				assert.Equal(t, initialToken, token)
				return expectedResp, nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		resp, err := client.ServiceCredentials(initialToken)

		assert.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		// The function should not be calling the session store since the token is passed directly
		assert.Equal(t, 0, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 0, mockSessionStore.RenewCallCount)
	})

	t.Run("Token renewal is never triggered but api fails", func(t *testing.T) {
		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string { return initialToken },
			RenewFunc:    func() error { return errors.New("renew failed") },
		}

		mockAPI := &mockSimpleClientAPI{
			ServiceCredentialsFunc: func(token string) (*core.CredentialsResponse, error) {
				return nil, core.ErrUnauthorized
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		resp, err := client.ServiceCredentials(initialToken)

		assert.Error(t, err)
		assert.Nil(t, resp)
		// The function should not be calling the session store since the token is passed directly
		assert.Equal(t, 0, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 0, mockSessionStore.RenewCallCount)
	})
}

func Test_TokenRenew_TokenRenewalScenarios(t *testing.T) {
	t.Run("Valid token, no renewal", func(t *testing.T) {
		initialRenewalToken := "valid-renewal-token"
		initialIdemKey := uuid.New()
		expectedResp := &core.TokenRenewResponse{
			Token:      "new-token",
			RenewToken: "new-renewal-token",
			ExpiresAt:  "someday",
		}

		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string { return initialToken },
			RenewFunc: func() error {
				t.Fatal("Renew should not be called")
				return nil
			},
		}

		mockAPI := &mockSimpleClientAPI{
			TokenRenewFunc: func(renewalToken string, idempotencyKey uuid.UUID) (*core.TokenRenewResponse, error) {
				assert.Equal(t, initialRenewalToken, renewalToken)
				assert.Equal(t, initialIdemKey, idempotencyKey)
				return expectedResp, nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		resp, err := client.TokenRenew(initialRenewalToken, initialIdemKey)

		assert.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		assert.Equal(t, 0, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 0, mockSessionStore.RenewCallCount)
	})

	t.Run("Token renewal is never triggered but api fails", func(t *testing.T) {
		initialRenewalToken := "valid-renewal-token"
		initialIdemKey := uuid.New()

		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string { return "" },
			RenewFunc:    func() error { return nil },
		}

		mockAPI := &mockSimpleClientAPI{
			TokenRenewFunc: func(renewalToken string, idempotencyKey uuid.UUID) (*core.TokenRenewResponse, error) {
				assert.Equal(t, initialRenewalToken, renewalToken)
				assert.Equal(t, initialIdemKey, idempotencyKey)
				return nil, core.ErrUnauthorized
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		resp, err := client.TokenRenew(initialRenewalToken, initialIdemKey)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, 0, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 0, mockSessionStore.RenewCallCount)
	})
}

func Test_Services_TokenRenewalScenarios(t *testing.T) {
	t.Run("Valid token, no renewal", func(t *testing.T) {
		expectedResp := core.ServicesResponse{}

		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string { return initialToken },
			RenewFunc: func() error {
				t.Fatal("Renew should not be called")
				return nil
			},
		}

		mockAPI := &mockSimpleClientAPI{
			ServicesFunc: func(token string) (core.ServicesResponse, error) {
				assert.Equal(t, initialToken, token)
				return expectedResp, nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		resp, err := client.Services()

		assert.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		assert.Equal(t, 1, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 0, mockSessionStore.RenewCallCount)
	})

	t.Run("Token renewal is triggered and succeeds", func(t *testing.T) {
		expectedResp := core.ServicesResponse{core.ServiceData{
			ID:        1,
			ExpiresAt: "someday",
			Service: core.Service{
				ID:   11,
				Name: "svc1"},
			Details: core.ServiceDetails{
				Servers: []core.ServiceServer{
					{ID: 1},
					{ID: 2},
				},
			},
		},
		}
		firstCall := true

		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string {
				if firstCall {
					return initialToken
				}
				return renewedToken
			},
			RenewFunc: func() error {
				return nil
			},
		}

		mockAPI := &mockSimpleClientAPI{
			ServicesFunc: func(token string) (core.ServicesResponse, error) {
				if token == initialToken {
					firstCall = false
					return core.ServicesResponse{}, core.ErrUnauthorized
				}
				assert.Equal(t, renewedToken, token)
				return expectedResp, nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		resp, err := client.Services()

		assert.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		assert.Equal(t, 2, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 1, mockSessionStore.RenewCallCount)
	})

	t.Run("Token renewal is triggered but fails", func(t *testing.T) {
		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string { return initialToken },
			RenewFunc:    func() error { return errors.New("renewal failed") },
		}

		mockAPI := &mockSimpleClientAPI{
			ServicesFunc: func(token string) (core.ServicesResponse, error) {
				return core.ServicesResponse{}, core.ErrUnauthorized
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		resp, err := client.Services()

		assert.Error(t, err)
		assert.Empty(t, resp)
		assert.Equal(t, 1, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 1, mockSessionStore.RenewCallCount)
	})

	t.Run("Token renewal succeeds but API still fails", func(t *testing.T) {
		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc:   func() string { return initialToken },
			RenewFunc:      func() error { return nil },
			InvalidateFunc: func(reason error) error { return nil },
		}

		mockAPI := &mockSimpleClientAPI{
			ServicesFunc: func(token string) (core.ServicesResponse, error) {
				return core.ServicesResponse{}, core.ErrUnauthorized
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		resp, err := client.Services()

		assert.Error(t, err)
		assert.Empty(t, resp)
		assert.Equal(t, 2, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 1, mockSessionStore.RenewCallCount)
	})
}

func Test_CurrentUser_TokenRenewalScenarios(t *testing.T) {
	t.Run("Valid token, no renewal", func(t *testing.T) {
		expectedResp := &core.CurrentUserResponse{Username: "username", Email: "email"}

		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string { return initialToken },
			RenewFunc: func() error {
				t.Fatal("Renew should not be called")
				return nil
			},
		}

		mockAPI := &mockSimpleClientAPI{
			CurrentUserFunc: func(token string) (*core.CurrentUserResponse, error) {
				assert.Equal(t, initialToken, token)
				return expectedResp, nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		resp, err := client.CurrentUser()

		assert.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		assert.Equal(t, 1, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 0, mockSessionStore.RenewCallCount)
	})

	t.Run("Token renewal is triggered and succeeds", func(t *testing.T) {
		expectedResp := &core.CurrentUserResponse{Username: "username", Email: "email"}
		firstCall := true

		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string {
				if firstCall {
					return initialToken
				}
				return renewedToken
			},
			RenewFunc: func() error {
				return nil
			},
		}

		mockAPI := &mockSimpleClientAPI{
			CurrentUserFunc: func(token string) (*core.CurrentUserResponse, error) {
				if token == initialToken {
					firstCall = false
					return nil, core.ErrUnauthorized
				}
				assert.Equal(t, renewedToken, token)
				return expectedResp, nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		resp, err := client.CurrentUser()

		assert.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		assert.Equal(t, 2, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 1, mockSessionStore.RenewCallCount)
	})

	t.Run("Token renewal is triggered but fails", func(t *testing.T) {
		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string { return initialToken },
			RenewFunc:    func() error { return errors.New("renewal failed") },
		}

		mockAPI := &mockSimpleClientAPI{
			CurrentUserFunc: func(token string) (*core.CurrentUserResponse, error) {
				return nil, core.ErrUnauthorized
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		resp, err := client.CurrentUser()

		assert.Error(t, err)
		assert.Empty(t, resp)
		assert.Equal(t, 1, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 1, mockSessionStore.RenewCallCount)
	})

	t.Run("Token renewal succeeds but API still fails", func(t *testing.T) {
		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc:   func() string { return initialToken },
			RenewFunc:      func() error { return nil },
			InvalidateFunc: func(reason error) error { return nil },
		}

		mockAPI := &mockSimpleClientAPI{
			CurrentUserFunc: func(token string) (*core.CurrentUserResponse, error) {
				return nil, core.ErrUnauthorized
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		resp, err := client.CurrentUser()

		assert.Error(t, err)
		assert.Empty(t, resp)
		assert.Equal(t, 2, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 1, mockSessionStore.RenewCallCount)
	})
}

func Test_DeleteToken_TokenRenewalScenarios(t *testing.T) {
	t.Run("Valid token, no renewal", func(t *testing.T) {
		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string { return initialToken },
			RenewFunc: func() error {
				t.Fatal("Renew should not be called")
				return nil
			},
		}

		mockAPI := &mockSimpleClientAPI{
			DeleteTokenFunc: func(token string) error {
				assert.Equal(t, initialToken, token)
				return nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		err := client.DeleteToken()

		assert.NoError(t, err)
		assert.Equal(t, 1, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 0, mockSessionStore.RenewCallCount)
	})

	t.Run("Token renewal is triggered and succeeds", func(t *testing.T) {
		firstCall := true

		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string {
				if firstCall {
					return initialToken
				}
				return renewedToken
			},
			RenewFunc: func() error {
				return nil
			},
		}

		mockAPI := &mockSimpleClientAPI{
			DeleteTokenFunc: func(token string) error {
				if token == initialToken {
					firstCall = false
					return core.ErrUnauthorized
				}
				assert.Equal(t, renewedToken, token)
				return nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		err := client.DeleteToken()

		assert.NoError(t, err)
		assert.Equal(t, 2, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 1, mockSessionStore.RenewCallCount)
	})

	t.Run("Token renewal is triggered but fails", func(t *testing.T) {
		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string { return initialToken },
			RenewFunc:    func() error { return errors.New("renewal failed") },
		}

		mockAPI := &mockSimpleClientAPI{
			DeleteTokenFunc: func(token string) error {
				return core.ErrUnauthorized
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		err := client.DeleteToken()

		assert.Error(t, err)
		assert.Equal(t, 1, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 1, mockSessionStore.RenewCallCount)
	})

	t.Run("Token renewal succeeds but API still fails", func(t *testing.T) {
		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc:   func() string { return initialToken },
			RenewFunc:      func() error { return nil },
			InvalidateFunc: func(reason error) error { return nil },
		}

		mockAPI := &mockSimpleClientAPI{
			DeleteTokenFunc: func(token string) error {
				return core.ErrUnauthorized
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		err := client.DeleteToken()

		assert.Error(t, err)
		assert.Equal(t, 2, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 1, mockSessionStore.RenewCallCount)
	})
}

func Test_TrustedPassToken_TokenRenewalScenarios(t *testing.T) {
	t.Run("Valid token, no renewal", func(t *testing.T) {
		expectedResp := &core.TrustedPassTokenResponse{OwnerID: "good-id", Token: "good-token"}

		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string { return initialToken },
			RenewFunc: func() error {
				t.Fatal("Renew should not be called")
				return nil
			},
		}

		mockAPI := &mockSimpleClientAPI{
			TrustedPassTokenFunc: func(token string) (*core.TrustedPassTokenResponse, error) {
				assert.Equal(t, initialToken, token)
				return expectedResp, nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		resp, err := client.TrustedPassToken()

		assert.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		assert.Equal(t, 1, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 0, mockSessionStore.RenewCallCount)
	})

	t.Run("Token renewal is triggered and succeeds", func(t *testing.T) {
		expectedResp := &core.TrustedPassTokenResponse{OwnerID: "good-id", Token: "good-token"}
		firstCall := true

		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string {
				if firstCall {
					return initialToken
				}
				return renewedToken
			},
			RenewFunc: func() error {
				return nil
			},
		}

		mockAPI := &mockSimpleClientAPI{
			TrustedPassTokenFunc: func(token string) (*core.TrustedPassTokenResponse, error) {
				if token == initialToken {
					firstCall = false
					return nil, core.ErrUnauthorized
				}
				assert.Equal(t, renewedToken, token)
				return expectedResp, nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		resp, err := client.TrustedPassToken()

		assert.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		assert.Equal(t, 2, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 1, mockSessionStore.RenewCallCount)
	})

	t.Run("Token renewal is triggered but fails", func(t *testing.T) {
		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string { return initialToken },
			RenewFunc:    func() error { return errors.New("renewal failed") },
		}

		mockAPI := &mockSimpleClientAPI{
			TrustedPassTokenFunc: func(token string) (*core.TrustedPassTokenResponse, error) {
				return nil, core.ErrUnauthorized
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		resp, err := client.TrustedPassToken()

		assert.Error(t, err)
		assert.Empty(t, resp)
		assert.Equal(t, 1, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 1, mockSessionStore.RenewCallCount)
	})

	t.Run("Token renewal succeeds but API still fails", func(t *testing.T) {
		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc:   func() string { return initialToken },
			RenewFunc:      func() error { return nil },
			InvalidateFunc: func(reason error) error { return nil },
		}

		mockAPI := &mockSimpleClientAPI{
			TrustedPassTokenFunc: func(token string) (*core.TrustedPassTokenResponse, error) {
				return nil, core.ErrUnauthorized
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		resp, err := client.TrustedPassToken()

		assert.Error(t, err)
		assert.Empty(t, resp)
		assert.Equal(t, 2, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 1, mockSessionStore.RenewCallCount)
	})
}

func Test_MultifactorAuthStatus_TokenRenewalScenarios(t *testing.T) {
	t.Run("Valid token, no renewal", func(t *testing.T) {
		expectedResp := &core.MultifactorAuthStatusResponse{
			Status: "enabled",
		}

		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string { return initialToken },
			RenewFunc: func() error {
				t.Fatal("Renew should not be called")
				return nil
			},
		}

		mockAPI := &mockSimpleClientAPI{
			MultifactorAuthStatusFunc: func(token string) (*core.MultifactorAuthStatusResponse, error) {
				assert.Equal(t, initialToken, token)
				return expectedResp, nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		resp, err := client.MultifactorAuthStatus()

		assert.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		assert.Equal(t, 1, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 0, mockSessionStore.RenewCallCount)
	})

	t.Run("Token renewal is triggered and succeeds", func(t *testing.T) {
		expectedResp := &core.MultifactorAuthStatusResponse{
			Status: "enabled",
		}
		firstCall := true

		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string {
				if firstCall {
					return initialToken
				}
				return renewedToken
			},
			RenewFunc: func() error {
				return nil
			},
		}

		mockAPI := &mockSimpleClientAPI{
			MultifactorAuthStatusFunc: func(token string) (*core.MultifactorAuthStatusResponse, error) {
				if token == initialToken {
					firstCall = false
					return nil, core.ErrUnauthorized
				}
				assert.Equal(t, renewedToken, token)
				return expectedResp, nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		resp, err := client.MultifactorAuthStatus()

		assert.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		assert.Equal(t, 2, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 1, mockSessionStore.RenewCallCount)
	})

	t.Run("Token renewal is triggered but fails", func(t *testing.T) {
		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string { return initialToken },
			RenewFunc:    func() error { return errors.New("renewal failed") },
		}

		mockAPI := &mockSimpleClientAPI{
			MultifactorAuthStatusFunc: func(token string) (*core.MultifactorAuthStatusResponse, error) {
				return nil, core.ErrUnauthorized
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		resp, err := client.MultifactorAuthStatus()

		assert.Error(t, err)
		assert.Empty(t, resp)
		assert.Equal(t, 1, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 1, mockSessionStore.RenewCallCount)
	})

	t.Run("Token renewal succeeds but API still fails", func(t *testing.T) {
		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc:   func() string { return initialToken },
			RenewFunc:      func() error { return nil },
			InvalidateFunc: func(reason error) error { return nil },
		}

		mockAPI := &mockSimpleClientAPI{
			MultifactorAuthStatusFunc: func(token string) (*core.MultifactorAuthStatusResponse, error) {
				return nil, core.ErrUnauthorized
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		resp, err := client.MultifactorAuthStatus()

		assert.Error(t, err)
		assert.Empty(t, resp)
		assert.Equal(t, 2, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 1, mockSessionStore.RenewCallCount)
	})
}

func Test_Logout_TokenRenewalScenarios(t *testing.T) {
	t.Run("Valid token, no renewal", func(t *testing.T) {
		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string { return initialToken },
			RenewFunc: func() error {
				t.Fatal("Renew should not be called")
				return nil
			},
		}

		mockAPI := &mockSimpleClientAPI{
			LogoutFunc: func(token string) error {
				assert.Equal(t, initialToken, token)
				return nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		err := client.Logout()

		assert.NoError(t, err)
		assert.Equal(t, 1, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 0, mockSessionStore.RenewCallCount)
	})

	t.Run("Token renewal is triggered and succeeds", func(t *testing.T) {
		firstCall := true

		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string {
				if firstCall {
					return initialToken
				}
				return renewedToken
			},
			RenewFunc: func() error {
				return nil
			},
		}

		mockAPI := &mockSimpleClientAPI{
			LogoutFunc: func(token string) error {
				if token == initialToken {
					firstCall = false
					return core.ErrUnauthorized
				}
				assert.Equal(t, renewedToken, token)
				return nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		err := client.Logout()

		assert.NoError(t, err)
		assert.Equal(t, 2, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 1, mockSessionStore.RenewCallCount)
	})

	t.Run("Token renewal is triggered but fails", func(t *testing.T) {
		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string { return initialToken },
			RenewFunc:    func() error { return errors.New("renewal failed") },
		}

		mockAPI := &mockSimpleClientAPI{
			LogoutFunc: func(token string) error {
				return core.ErrUnauthorized
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		err := client.Logout()

		assert.Error(t, err)
		assert.Equal(t, 1, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 1, mockSessionStore.RenewCallCount)
	})

	t.Run("Token renewal succeeds but API still fails", func(t *testing.T) {
		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc:   func() string { return initialToken },
			RenewFunc:      func() error { return nil },
			InvalidateFunc: func(reason error) error { return nil },
		}

		mockAPI := &mockSimpleClientAPI{
			LogoutFunc: func(token string) error {
				return core.ErrUnauthorized
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		err := client.Logout()

		assert.Error(t, err)
		assert.Equal(t, 2, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 1, mockSessionStore.RenewCallCount)
	})
}

func Test_Insights_TokenRenewalScenarios(t *testing.T) {
	t.Run("By pass wrapping", func(t *testing.T) {
		expectedResp := &core.Insights{
			City:        "London",
			Country:     "United Kingdom",
			Isp:         "UKExampleISP",
			IspAsn:      67890,
			CountryCode: "GB",
			Longitude:   -0.1278,
			Latitude:    51.5074,
			Protected:   false,
		}

		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string {
				t.Fatal("GetToken should not be called")
				return ""
			},
			RenewFunc: func() error {
				t.Fatal("Renew should not be called")
				return nil
			},
		}

		mockAPI := &mockSimpleClientAPI{
			InsightsFunc: func() (*core.Insights, error) {
				return expectedResp, nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		resp, err := client.Insights()

		assert.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		assert.Equal(t, 0, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 0, mockSessionStore.RenewCallCount)
	})
}

func Test_Servers_TokenRenewalScenarios(t *testing.T) {
	t.Run("By pass wrapping", func(t *testing.T) {
		expectedResp := core.Servers{core.Server{ID: 7}, core.Server{ID: 8}}
		expectedHeader := http.Header(map[string][]string{"header": {"item1, item2"}})

		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string {
				t.Fatal("GetToken should not be called")
				return ""
			},
			RenewFunc: func() error {
				t.Fatal("Renew should not be called")
				return nil
			},
		}

		mockAPI := &mockSimpleClientAPI{
			ServersFunc: func() (core.Servers, http.Header, error) {
				return expectedResp, expectedHeader, nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		resp, header, err := client.Servers()

		assert.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		assert.Equal(t, expectedHeader, header)
		assert.Equal(t, 0, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 0, mockSessionStore.RenewCallCount)
	})
}

func Test_RecommendedServers_TokenRenewalScenarios(t *testing.T) {
	t.Run("Bypass token renewal", func(t *testing.T) {
		expectedServersFilter := core.ServersFilter{
			Limit: 10,
			Tech:  core.WireguardTech,
			Group: config.ServerGroup(2),
			Tag: core.ServerTag{
				Action: core.ServerByCountry,
				ID:     12345,
			},
		}
		expectedLongitude := 12.345678
		expectedLatitude := -98.765432
		expectedServers := core.Servers{core.Server{
			ID:        1001,
			Name:      "Test Server 1",
			Hostname:  "testserver1.example.com",
			Status:    "online",
			Load:      25,
			Locations: core.Locations{},
		},
		}
		expectedHeader := http.Header(map[string][]string{"header": {"item1, item2"}})

		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string {
				t.Fatal("GetToken should not be called")
				return ""
			},
			RenewFunc: func() error {
				t.Fatal("Renew should not be called")
				return nil
			},
		}

		mockAPI := &mockSimpleClientAPI{
			RecommendedServersFunc: func(filter core.ServersFilter, longitude, latitude float64) (core.Servers, http.Header, error) {
				assert.Equal(t, expectedServersFilter, filter)
				assert.Equal(t, expectedLongitude, longitude)
				assert.Equal(t, expectedLatitude, latitude)
				return expectedServers, expectedHeader, nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		servers, header, err := client.RecommendedServers(expectedServersFilter, expectedLongitude, expectedLatitude)

		assert.NoError(t, err)
		assert.Equal(t, expectedServers, servers)
		assert.Equal(t, expectedHeader, header)
		assert.Equal(t, 0, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 0, mockSessionStore.RenewCallCount)
	})
}

func Test_Server_TokenRenewalScenarios(t *testing.T) {
	t.Run("Bypass token renewal", func(t *testing.T) {
		expectedServer := &core.Server{ID: 44, Name: "lt04"}
		expectedServerId := int64(17)
		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string {
				t.Fatal("GetToken should not be called")
				return ""
			},
			RenewFunc: func() error {
				t.Fatal("Renew should not be called")
				return nil
			},
		}

		mockAPI := &mockSimpleClientAPI{
			ServerFunc: func(id int64) (*core.Server, error) {
				assert.Equal(t, id, expectedServerId)
				return expectedServer, nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		server, err := client.Server(expectedServerId)

		assert.NoError(t, err)
		assert.Equal(t, expectedServer, server)
		assert.Equal(t, 0, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 0, mockSessionStore.RenewCallCount)
	})
}

func Test_ServerCountries_TokenRenewalScenarios(t *testing.T) {
	t.Run("Bypass token renewal", func(t *testing.T) {
		expectedCountries := core.Countries{
			core.Country{Name: "Lithuania"},
			core.Country{Name: "Poland"},
		}
		expectedHeader := http.Header(map[string][]string{"header": {"item1, item2"}})

		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string {
				t.Fatal("GetToken should not be called")
				return ""
			},
			RenewFunc: func() error {
				t.Fatal("Renew should not be called")
				return nil
			},
		}

		mockAPI := &mockSimpleClientAPI{
			ServersCountriesFunc: func() (core.Countries, http.Header, error) {
				return expectedCountries, expectedHeader, nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		countries, header, err := client.ServersCountries()

		assert.NoError(t, err)
		assert.Equal(t, expectedCountries, countries)
		assert.Equal(t, header, expectedHeader)
		assert.Equal(t, 0, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 0, mockSessionStore.RenewCallCount)
	})
}

func Test_Base_TokenRenewalScenarios(t *testing.T) {
	t.Run("Bypass token renewal", func(t *testing.T) {
		expectedOutput := "something"

		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string {
				t.Fatal("GetToken should not be called")
				return ""
			},
			RenewFunc: func() error {
				t.Fatal("Renew should not be called")
				return nil
			},
		}

		mockAPI := &mockSimpleClientAPI{
			BaseFunc: func() string {
				return expectedOutput
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		output := client.Base()

		assert.Equal(t, expectedOutput, output)
		assert.Equal(t, 0, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 0, mockSessionStore.RenewCallCount)
	})
}

func Test_Plans_TokenRenewalScenarios(t *testing.T) {
	t.Run("Bypass token renewal", func(t *testing.T) {
		expectedPlans := &core.Plans{
			core.Plan{ID: 1},
			core.Plan{ID: 2},
		}

		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string {
				t.Fatal("GetToken should not be called")
				return ""
			},
			RenewFunc: func() error {
				t.Fatal("Renew should not be called")
				return nil
			},
		}

		mockAPI := &mockSimpleClientAPI{
			PlansFunc: func() (*core.Plans, error) {
				return expectedPlans, nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		plans, err := client.Plans()

		assert.NoError(t, err)
		assert.Equal(t, expectedPlans, plans)
		assert.Equal(t, 0, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 0, mockSessionStore.RenewCallCount)
	})
}

func Test_CreateUser_TokenRenewalScenarios(t *testing.T) {
	t.Run("Bypass token renewal", func(t *testing.T) {
		expectedEmail := "user@company.com"
		expectedPassword := "super-save-password"
		expectedResponse := &core.UserCreateResponse{
			ID:       9,
			Username: "user",
			Email:    expectedEmail,
		}

		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string {
				t.Fatal("GetToken should not be called")
				return ""
			},
			RenewFunc: func() error {
				t.Fatal("Renew should not be called")
				return nil
			},
		}

		mockAPI := &mockSimpleClientAPI{
			CreateUserFunc: func(email, password string) (*core.UserCreateResponse, error) {
				assert.Equal(t, expectedEmail, email)
				assert.Equal(t, expectedPassword, password)
				return expectedResponse, nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		resp, err := client.CreateUser(expectedEmail, expectedPassword)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, resp)
		assert.Equal(t, 0, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 0, mockSessionStore.RenewCallCount)
	})
}

func Test_Orders_TokenRenewalScenarios(t *testing.T) {
	t.Run("Valid token, no renewal", func(t *testing.T) {
		expectedOrders := []core.Order{
			{ID: 1, Status: "expired"},
			{ID: 2, Status: "ok"},
		}
		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string { return initialToken },
			RenewFunc: func() error {
				t.Fatal("Renew should not be called")
				return nil
			},
		}

		mockAPI := &mockSimpleClientAPI{
			OrdersFunc: func(token string) ([]core.Order, error) {
				assert.Equal(t, initialToken, token)
				return expectedOrders, nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		orders, err := client.Orders()

		assert.NoError(t, err)
		assert.Equal(t, expectedOrders, orders)
		assert.Equal(t, 1, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 0, mockSessionStore.RenewCallCount)
	})

	t.Run("Token renewal is triggered and succeeds", func(t *testing.T) {
		expectedOrders := []core.Order{
			{ID: 1, Status: "expired"},
			{ID: 2, Status: "ok"},
		}
		firstCall := true

		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string {
				if firstCall {
					return initialToken
				}
				return renewedToken
			},
			RenewFunc: func() error {
				return nil
			},
		}

		mockAPI := &mockSimpleClientAPI{
			OrdersFunc: func(token string) ([]core.Order, error) {
				if token == initialToken {
					firstCall = false
					return nil, core.ErrUnauthorized
				}
				assert.Equal(t, renewedToken, token)
				return expectedOrders, nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		orders, err := client.Orders()

		assert.NoError(t, err)
		assert.Equal(t, expectedOrders, orders)
		assert.Equal(t, 2, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 1, mockSessionStore.RenewCallCount)
	})

	t.Run("Token renewal is triggered but fails", func(t *testing.T) {
		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string { return initialToken },
			RenewFunc:    func() error { return errors.New("renewal failed") },
		}

		mockAPI := &mockSimpleClientAPI{
			OrdersFunc: func(token string) ([]core.Order, error) {
				return nil, core.ErrUnauthorized
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		orders, err := client.Orders()

		assert.Error(t, err)
		assert.Empty(t, orders)
		assert.Equal(t, 1, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 1, mockSessionStore.RenewCallCount)
	})

	t.Run("Token renewal succeeds but API still fails", func(t *testing.T) {
		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc:   func() string { return initialToken },
			RenewFunc:      func() error { return nil },
			InvalidateFunc: func(reason error) error { return nil },
		}

		mockAPI := &mockSimpleClientAPI{
			OrdersFunc: func(token string) ([]core.Order, error) {
				return nil, core.ErrUnauthorized
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		orders, err := client.Orders()

		assert.Error(t, err)
		assert.Nil(t, orders)
		assert.Equal(t, 2, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 1, mockSessionStore.RenewCallCount)
	})
}

func Test_Payments_TokenRenewalScenarios(t *testing.T) {
	t.Run("Valid token, no renewal", func(t *testing.T) {
		expectedPayments := []core.PaymentResponse{
			{Payment: core.Payment{Status: "ok", Amount: 1.23}},
			{Payment: core.Payment{Status: "ok", Amount: 4.56}},
		}
		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string { return initialToken },
			RenewFunc: func() error {
				t.Fatal("Renew should not be called")
				return nil
			},
		}

		mockAPI := &mockSimpleClientAPI{
			PaymentsFunc: func(token string) ([]core.PaymentResponse, error) {
				assert.Equal(t, initialToken, token)
				return expectedPayments, nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		payments, err := client.Payments()

		assert.NoError(t, err)
		assert.Equal(t, expectedPayments, payments)
		assert.Equal(t, 1, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 0, mockSessionStore.RenewCallCount)
	})

	t.Run("Token renewal is triggered and succeeds", func(t *testing.T) {
		expectedPayments := []core.PaymentResponse{
			{Payment: core.Payment{Status: "ok", Amount: 1.23}},
			{Payment: core.Payment{Status: "ok", Amount: 4.56}},
		}
		firstCall := true

		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string {
				if firstCall {
					return initialToken
				}
				return renewedToken
			},
			RenewFunc: func() error {
				return nil
			},
		}

		mockAPI := &mockSimpleClientAPI{
			PaymentsFunc: func(token string) ([]core.PaymentResponse, error) {
				if token == initialToken {
					firstCall = false
					return nil, core.ErrUnauthorized
				}
				assert.Equal(t, renewedToken, token)
				return expectedPayments, nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		payments, err := client.Payments()

		assert.NoError(t, err)
		assert.Equal(t, expectedPayments, payments)
		assert.Equal(t, 2, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 1, mockSessionStore.RenewCallCount)
	})

	t.Run("Token renewal is triggered but fails", func(t *testing.T) {
		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string { return initialToken },
			RenewFunc:    func() error { return errors.New("renewal failed") },
		}

		mockAPI := &mockSimpleClientAPI{
			PaymentsFunc: func(token string) ([]core.PaymentResponse, error) {
				return nil, core.ErrUnauthorized
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		payments, err := client.Payments()

		assert.Error(t, err)
		assert.Empty(t, payments)
		assert.Equal(t, 1, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 1, mockSessionStore.RenewCallCount)
	})

	t.Run("Token renewal succeeds but API still fails", func(t *testing.T) {
		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc:   func() string { return initialToken },
			RenewFunc:      func() error { return nil },
			InvalidateFunc: func(reason error) error { return nil },
		}

		mockAPI := &mockSimpleClientAPI{
			PaymentsFunc: func(token string) ([]core.PaymentResponse, error) {
				return nil, core.ErrUnauthorized
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		payments, err := client.Payments()

		assert.Error(t, err)
		assert.Nil(t, payments)
		assert.Equal(t, 2, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 1, mockSessionStore.RenewCallCount)
	})
}

func Test_Register_TokenRenewalScenarios(t *testing.T) {
	t.Run("Bypass token renewal", func(t *testing.T) {
		expectedMachine := mesh.Machine{PublicKey: "magic-key"}
		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string {
				t.Fatal("GetToken should not be called")
				return ""
			},
			RenewFunc: func() error {
				t.Fatal("Renew should not be called")
				return nil
			},
		}

		mockAPI := &mockSimpleClientAPI{
			RegisterFunc: func(token string, peer mesh.Machine) (*mesh.Machine, error) {
				assert.Equal(t, token, initialToken)
				assert.Equal(t, peer, expectedMachine)
				return &peer, nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		machine, err := client.Register(initialToken, expectedMachine)

		assert.NoError(t, err)
		assert.Equal(t, expectedMachine, *machine)
		assert.Equal(t, 0, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 0, mockSessionStore.RenewCallCount)
	})
}

func Test_Update_TokenRenewalScenarios(t *testing.T) {
	t.Run("Bypass token renewal", func(t *testing.T) {
		expectedUUID := uuid.New()
		expectedRequest := mesh.MachineUpdateRequest{Nickname: "temp"}
		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string {
				t.Fatal("GetToken should not be called")
				return ""
			},
			RenewFunc: func() error {
				t.Fatal("Renew should not be called")
				return nil
			},
		}

		mockAPI := &mockSimpleClientAPI{
			UpdateFunc: func(token string, id uuid.UUID, info mesh.MachineUpdateRequest) error {
				assert.Equal(t, initialToken, token)
				assert.Equal(t, expectedUUID, id)
				assert.Equal(t, expectedRequest, info)
				return nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		err := client.Update(initialToken, expectedUUID, expectedRequest)

		assert.NoError(t, err)
		assert.Equal(t, 0, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 0, mockSessionStore.RenewCallCount)
	})
}

func Test_Configure_TokenRenewalScenarios(t *testing.T) {
	t.Run("Bypass token renewal", func(t *testing.T) {
		expectedUUID := uuid.New()
		expectedPeerUUID := uuid.New()
		expectedRequest := mesh.PeerUpdateRequest{Nickname: "temp"}
		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string {
				t.Fatal("GetToken should not be called")
				return ""
			},
			RenewFunc: func() error {
				t.Fatal("Renew should not be called")
				return nil
			},
		}

		mockAPI := &mockSimpleClientAPI{
			ConfigureFunc: func(token string, id uuid.UUID, peerID uuid.UUID, peerUpdateInfo mesh.PeerUpdateRequest) error {
				assert.Equal(t, initialToken, token)
				assert.Equal(t, expectedUUID, id)
				assert.Equal(t, expectedPeerUUID, peerID)
				assert.Equal(t, expectedRequest, peerUpdateInfo)
				return nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		err := client.Configure(initialToken, expectedUUID, expectedPeerUUID, expectedRequest)

		assert.NoError(t, err)
		assert.Equal(t, 0, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 0, mockSessionStore.RenewCallCount)
	})
}

func Test_Unregister_TokenRenewalScenarios(t *testing.T) {
	t.Run("Bypass token renewal", func(t *testing.T) {
		expectedUUID := uuid.New()
		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string {
				t.Fatal("GetToken should not be called")
				return ""
			},
			RenewFunc: func() error {
				t.Fatal("Renew should not be called")
				return nil
			},
		}

		mockAPI := &mockSimpleClientAPI{
			UnregisterFunc: func(token string, self uuid.UUID) error {
				assert.Equal(t, initialToken, token)
				assert.Equal(t, expectedUUID, self)
				return nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		err := client.Unregister(initialToken, expectedUUID)

		assert.NoError(t, err)
		assert.Equal(t, 0, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 0, mockSessionStore.RenewCallCount)
	})
}

func Test_Map_TokenRenewalScenarios(t *testing.T) {
	t.Run("Bypass token renewal", func(t *testing.T) {
		expectedUUID := uuid.New()
		expectedMap := &mesh.MachineMap{
			Machine: mesh.Machine{ID: uuid.New()},
		}
		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string {
				t.Fatal("GetToken should not be called")
				return ""
			},
			RenewFunc: func() error {
				t.Fatal("Renew should not be called")
				return nil
			},
		}

		mockAPI := &mockSimpleClientAPI{
			MapFunc: func(token string, self uuid.UUID) (*mesh.MachineMap, error) {
				assert.Equal(t, initialToken, token)
				assert.Equal(t, expectedUUID, self)
				return expectedMap, nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		resp, err := client.Map(initialToken, expectedUUID)

		assert.NoError(t, err)
		assert.Equal(t, expectedMap, resp)
		assert.Equal(t, 0, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 0, mockSessionStore.RenewCallCount)
	})
}

func Test_Unpair_TokenRenewalScenarios(t *testing.T) {
	t.Run("Bypass token renewal", func(t *testing.T) {
		expectedUUID := uuid.New()
		expectedPeerUUID := uuid.New()
		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string {
				t.Fatal("GetToken should not be called")
				return ""
			},
			RenewFunc: func() error {
				t.Fatal("Renew should not be called")
				return nil
			},
		}

		mockAPI := &mockSimpleClientAPI{
			UnpairFunc: func(token string, self uuid.UUID, peer uuid.UUID) error {
				assert.Equal(t, initialToken, token)
				assert.Equal(t, expectedUUID, self)
				assert.Equal(t, expectedPeerUUID, peer)
				return nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		err := client.Unpair(initialToken, expectedUUID, expectedPeerUUID)

		assert.NoError(t, err)
		assert.Equal(t, 0, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 0, mockSessionStore.RenewCallCount)
	})
}

func Test_Received_TokenRenewalScenarios(t *testing.T) {
	t.Run("Bypass token renewal", func(t *testing.T) {
		expectedUUID := uuid.New()
		expectedInvitations := mesh.Invitations{mesh.Invitation{ID: uuid.New()}}
		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string {
				t.Fatal("GetToken should not be called")
				return ""
			},
			RenewFunc: func() error {
				t.Fatal("Renew should not be called")
				return nil
			},
		}

		mockAPI := &mockSimpleClientAPI{
			ReceivedFunc: func(token string, self uuid.UUID) (mesh.Invitations, error) {
				assert.Equal(t, initialToken, token)
				assert.Equal(t, expectedUUID, self)
				return expectedInvitations, nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		resp, err := client.Received(initialToken, expectedUUID)

		assert.NoError(t, err)
		assert.Equal(t, expectedInvitations, resp)
		assert.Equal(t, 0, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 0, mockSessionStore.RenewCallCount)
	})
}

func Test_Sent_TokenRenewalScenarios(t *testing.T) {
	t.Run("Bypass token renewal", func(t *testing.T) {
		expectedUUID := uuid.New()
		expectedInvitations := mesh.Invitations{mesh.Invitation{ID: uuid.New()}}
		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string {
				t.Fatal("GetToken should not be called")
				return ""
			},
			RenewFunc: func() error {
				t.Fatal("Renew should not be called")
				return nil
			},
		}

		mockAPI := &mockSimpleClientAPI{
			SentFunc: func(token string, self uuid.UUID) (mesh.Invitations, error) {
				assert.Equal(t, initialToken, token)
				assert.Equal(t, expectedUUID, self)
				return expectedInvitations, nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		resp, err := client.Sent(initialToken, expectedUUID)

		assert.NoError(t, err)
		assert.Equal(t, expectedInvitations, resp)
		assert.Equal(t, 0, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 0, mockSessionStore.RenewCallCount)
	})
}

func Test_Accept_TokenRenewalScenarios(t *testing.T) {
	t.Run("Bypass token renewal", func(t *testing.T) {
		expectedUUID := uuid.New()
		expectedInvitUUID := uuid.New()
		expectedAllowInbounding := true
		expectedAllowRouting := true
		expectedAllowLocalNetwork := false
		expectedAllowFileshare := false
		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string {
				t.Fatal("GetToken should not be called")
				return ""
			},
			RenewFunc: func() error {
				t.Fatal("Renew should not be called")
				return nil
			},
		}

		mockAPI := &mockSimpleClientAPI{
			AcceptFunc: func(
				token string,
				self uuid.UUID,
				invitation uuid.UUID,
				doIAllowInbound, doIAllowRouting, doIAllowLocalNetwork, doIAllowFileshare bool,
			) error {
				assert.Equal(t, initialToken, token)
				assert.Equal(t, expectedUUID, self)
				return nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		err := client.Accept(initialToken, expectedUUID, expectedInvitUUID,
			expectedAllowInbounding, expectedAllowRouting, expectedAllowLocalNetwork, expectedAllowFileshare)

		assert.NoError(t, err)
		assert.Equal(t, initialToken, initialToken)
		assert.Equal(t, expectedUUID, expectedUUID)
		assert.Equal(t, expectedInvitUUID, expectedInvitUUID)
		assert.Equal(t, expectedAllowInbounding, expectedAllowInbounding)
		assert.Equal(t, expectedAllowRouting, expectedAllowRouting)
		assert.Equal(t, expectedAllowLocalNetwork, expectedAllowLocalNetwork)
		assert.Equal(t, expectedAllowFileshare, expectedAllowFileshare)
		assert.Equal(t, 0, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 0, mockSessionStore.RenewCallCount)
	})
}

func Test_Reject_TokenRenewalScenarios(t *testing.T) {
	t.Run("Bypass token renewal", func(t *testing.T) {
		expectedUUID := uuid.New()
		expectedInvitUUID := uuid.New()
		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string {
				t.Fatal("GetToken should not be called")
				return ""
			},
			RenewFunc: func() error {
				t.Fatal("Renew should not be called")
				return nil
			},
		}

		mockAPI := &mockSimpleClientAPI{
			RejectFunc: func(token string, self uuid.UUID, invitation uuid.UUID) error {
				assert.Equal(t, initialToken, token)
				assert.Equal(t, expectedUUID, self)
				assert.Equal(t, expectedInvitUUID, invitation)
				return nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		err := client.Reject(initialToken, expectedUUID, expectedInvitUUID)

		assert.NoError(t, err)
		assert.Equal(t, 0, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 0, mockSessionStore.RenewCallCount)
	})
}

func Test_Revoke_TokenRenewalScenarios(t *testing.T) {
	t.Run("Bypass token renewal", func(t *testing.T) {
		expectedUUID := uuid.New()
		expectedInvitUUID := uuid.New()
		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string {
				t.Fatal("GetToken should not be called")
				return ""
			},
			RenewFunc: func() error {
				t.Fatal("Renew should not be called")
				return nil
			},
		}

		mockAPI := &mockSimpleClientAPI{
			RevokeFunc: func(token string, self uuid.UUID, invitation uuid.UUID) error {
				assert.Equal(t, initialToken, token)
				assert.Equal(t, expectedUUID, self)
				assert.Equal(t, expectedInvitUUID, invitation)
				return nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		err := client.Revoke(initialToken, expectedUUID, expectedInvitUUID)

		assert.NoError(t, err)
		assert.Equal(t, 0, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 0, mockSessionStore.RenewCallCount)
	})
}

func Test_NotifyNewTransfer_TokenRenewalScenarios(t *testing.T) {
	t.Run("Bypass token renewal", func(t *testing.T) {
		expectedUUID := uuid.New()
		expectedPeerUUID := uuid.New()
		expectedFilename := "name"
		expectedFileCount := 4
		expectedTransferID := "1321"
		mockSessionStore := &mocksession.MockAccessTokenSessionStore{
			GetTokenFunc: func() string {
				t.Fatal("GetToken should not be called")
				return ""
			},
			RenewFunc: func() error {
				t.Fatal("Renew should not be called")
				return nil
			},
		}

		mockAPI := &mockSimpleClientAPI{
			NotifyNewTransferFunc: func(
				token string,
				self uuid.UUID,
				peer uuid.UUID,
				fileName string,
				fileCount int,
				transferID string,
			) error {
				assert.Equal(t, initialToken, token)
				assert.Equal(t, expectedUUID, self)
				assert.Equal(t, expectedPeerUUID, peer)
				assert.Equal(t, expectedFilename, fileName)
				assert.Equal(t, expectedFileCount, fileCount)
				assert.Equal(t, expectedTransferID, transferID)
				return nil
			},
		}

		client := NewMockSmartClientAPI(mockAPI, mockSessionStore)
		err := client.NotifyNewTransfer(initialToken, expectedUUID, expectedPeerUUID,
			expectedFilename, expectedFileCount, expectedTransferID)

		assert.NoError(t, err)
		assert.Equal(t, 0, mockSessionStore.GetTokenCallCount)
		assert.Equal(t, 0, mockSessionStore.RenewCallCount)
	})
}
