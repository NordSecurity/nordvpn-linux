package daemon

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/NordSecurity/nordvpn-linux/core"
	daemonevents "github.com/NordSecurity/nordvpn-linux/daemon/events"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	testauth "github.com/NordSecurity/nordvpn-linux/test/mock/auth"
	testcore "github.com/NordSecurity/nordvpn-linux/test/mock/core"
)

func TestLoginEvents(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name            string
		setup           func(*daemonevents.MockPublisherSubscriber[events.DataAuthorization]) *RPC
		loginType1      pb.LoginType
		expectedStatus1 pb.LoginStatus
		expectedEvent1  events.DataAuthorization
		expectedError1  bool
		loginType2      pb.LoginType
		expectedStatus2 pb.LoginStatus
		expectedEvent2  events.DataAuthorization
		expectedError2  bool
	}{
		{
			name: "login attempt then login success",
			setup: func(em *daemonevents.MockPublisherSubscriber[events.DataAuthorization]) *RPC {
				return &RPC{
					consentChecker:   &mock.AnalyticsConsentCheckerMock{ConsentCompleted: true},
					ac:               &testauth.AuthCheckerMock{LoggedIn: false},
					cm:               mock.NewMockConfigManager(),
					credentialsAPI:   &testcore.CredentialsAPIMock{},
					events:           &daemonevents.Events{User: &daemonevents.LoginEvents{Login: em}},
					authentication:   &testcore.AuthenticationAPImock{TokenValue: "token"},
					publisher:        &subs.Subject[string]{},
					ncClient:         &mock.NotificationClientMock{},
					initialLoginType: NewAtomicLoginType(),
				}
			},
			loginType1:      pb.LoginType_LoginType_LOGIN,
			expectedStatus1: pb.LoginStatus_SUCCESS,
			expectedEvent1: events.DataAuthorization{
				EventType:                  events.LoginLogin,
				EventTrigger:               events.TriggerUser,
				EventStatus:                events.StatusAttempt,
				IsAlteredFlowOnNordAccount: false,
			},
			expectedError1:  false,
			loginType2:      pb.LoginType_LoginType_LOGIN,
			expectedStatus2: pb.LoginStatus_SUCCESS,
			expectedEvent2: events.DataAuthorization{
				EventType:                  events.LoginLogin,
				EventTrigger:               events.TriggerUser,
				EventStatus:                events.StatusSuccess,
				IsAlteredFlowOnNordAccount: false,
			},
			expectedError2: false,
		},
		{
			name: "signup attempt then login success",
			setup: func(em *daemonevents.MockPublisherSubscriber[events.DataAuthorization]) *RPC {
				return &RPC{
					consentChecker:   &mock.AnalyticsConsentCheckerMock{ConsentCompleted: true},
					ac:               &testauth.AuthCheckerMock{LoggedIn: false},
					cm:               mock.NewMockConfigManager(),
					credentialsAPI:   &testcore.CredentialsAPIMock{},
					events:           &daemonevents.Events{User: &daemonevents.LoginEvents{Login: em}},
					authentication:   &testcore.AuthenticationAPImock{TokenValue: "token"},
					publisher:        &subs.Subject[string]{},
					ncClient:         &mock.NotificationClientMock{},
					initialLoginType: NewAtomicLoginType(),
				}
			},
			loginType1:      pb.LoginType_LoginType_SIGNUP,
			expectedStatus1: pb.LoginStatus_SUCCESS,
			expectedEvent1: events.DataAuthorization{
				EventType:                  events.LoginSignUp,
				EventTrigger:               events.TriggerUser,
				EventStatus:                events.StatusAttempt,
				IsAlteredFlowOnNordAccount: false,
			},
			expectedError1:  false,
			loginType2:      pb.LoginType_LoginType_LOGIN,
			expectedStatus2: pb.LoginStatus_SUCCESS,
			expectedEvent2: events.DataAuthorization{
				EventType:                  events.LoginLogin,
				EventTrigger:               events.TriggerUser,
				EventStatus:                events.StatusSuccess,
				IsAlteredFlowOnNordAccount: true,
			},
			expectedError2: false,
		},
		{
			name: "login attempt fail - consent not completed",
			setup: func(em *daemonevents.MockPublisherSubscriber[events.DataAuthorization]) *RPC {
				return &RPC{
					consentChecker: &mock.AnalyticsConsentCheckerMock{ConsentCompleted: false},
				}
			},
			loginType1:      pb.LoginType_LoginType_LOGIN,
			expectedStatus1: pb.LoginStatus_CONSENT_MISSING,
			expectedError1:  false,
			loginType2:      pb.LoginType_LoginType_LOGIN,
			expectedStatus2: pb.LoginStatus_CONSENT_MISSING,
			expectedError2:  false,
		},
		{
			name: "signup attempt fail - already logged in",
			setup: func(em *daemonevents.MockPublisherSubscriber[events.DataAuthorization]) *RPC {
				return &RPC{
					consentChecker: &mock.AnalyticsConsentCheckerMock{ConsentCompleted: true},
					ac:             &testauth.AuthCheckerMock{LoggedIn: true},
				}
			},
			loginType1:      pb.LoginType_LoginType_SIGNUP,
			expectedStatus1: pb.LoginStatus_ALREADY_LOGGED_IN,
			expectedError1:  false,
			loginType2:      pb.LoginType_LoginType_LOGIN,
			expectedError2:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventsMock := &daemonevents.MockPublisherSubscriber[events.DataAuthorization]{}
			r := tt.setup(eventsMock)
			resp1, err := r.LoginOAuth2(context.Background(), &pb.LoginOAuth2Request{Type: tt.loginType1})
			if tt.expectedError1 {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, tt.expectedStatus1 == resp1.Status)
				if resp1.Status == pb.LoginStatus_SUCCESS {
					assert.True(t, tt.expectedEvent1.EventType == eventsMock.Event.EventType)
					assert.True(t, tt.expectedEvent1.EventStatus == eventsMock.Event.EventStatus)
					assert.True(t, tt.expectedEvent1.EventTrigger == eventsMock.Event.EventTrigger)
					assert.True(t, tt.expectedEvent1.IsAlteredFlowOnNordAccount == eventsMock.Event.IsAlteredFlowOnNordAccount)
				}
			}
			resp2, err := r.LoginOAuth2Callback(context.Background(), &pb.LoginOAuth2CallbackRequest{
				Token: "exchange-token",
				Type:  tt.loginType2,
			})
			if tt.expectedError2 {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, tt.expectedStatus2 == resp2.Status)
				if resp2.Status == pb.LoginStatus_SUCCESS {
					assert.True(t, tt.expectedEvent2.EventType == eventsMock.Event.EventType)
					assert.True(t, tt.expectedEvent2.EventStatus == eventsMock.Event.EventStatus)
					assert.True(t, tt.expectedEvent2.EventTrigger == eventsMock.Event.EventTrigger)
					assert.True(t, tt.expectedEvent2.IsAlteredFlowOnNordAccount == eventsMock.Event.IsAlteredFlowOnNordAccount)
				}
			}
		})
	}
}

func TestLoginOAuth2_AnalyticsEvents(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name                string
		setup               func(*daemonevents.MockPublisherSubscriber[events.DataAuthorization]) *RPC
		loginType           pb.LoginType
		prevLoginType       pb.LoginType
		expectedStatus      pb.LoginStatus
		expectedEventsCount int
		validateEvents      func(*testing.T, []*events.DataAuthorization)
	}{
		{
			name: "network unreachable error",
			setup: func(em *daemonevents.MockPublisherSubscriber[events.DataAuthorization]) *RPC {
				return &RPC{
					consentChecker: &mock.AnalyticsConsentCheckerMock{ConsentCompleted: true},
					ac:             &testauth.AuthCheckerMock{LoggedIn: false},
					events:         &daemonevents.Events{User: &daemonevents.LoginEvents{Login: em}},
					authentication: &testcore.AuthenticationAPImock{
						LoginError: errors.New("network is unreachable"),
					},
					initialLoginType: NewAtomicLoginType(),
				}
			},
			loginType:           pb.LoginType_LoginType_LOGIN,
			prevLoginType:       pb.LoginType_LoginType_UNKNOWN,
			expectedStatus:      pb.LoginStatus_NO_NET,
			expectedEventsCount: 1,
			validateEvents: func(t *testing.T, evts []*events.DataAuthorization) {
				assert.Equal(t, events.StatusFailure, evts[0].EventStatus)
				assert.Equal(t, events.TriggerUser, evts[0].EventTrigger)
				assert.Equal(t, events.LoginLogin, evts[0].EventType)
				assert.Equal(t, events.ReasonLoginURLRetrieveFailed, evts[0].Reason)
				assert.Equal(t, -1, evts[0].DurationMs)
			},
		},
		{
			name: "client timeout error",
			setup: func(em *daemonevents.MockPublisherSubscriber[events.DataAuthorization]) *RPC {
				return &RPC{
					consentChecker: &mock.AnalyticsConsentCheckerMock{ConsentCompleted: true},
					ac:             &testauth.AuthCheckerMock{LoggedIn: false},
					events:         &daemonevents.Events{User: &daemonevents.LoginEvents{Login: em}},
					authentication: &testcore.AuthenticationAPImock{
						LoginError: errors.New("Client.Timeout exceeded while awaiting headers"),
					},
					initialLoginType: NewAtomicLoginType(),
				}
			},
			loginType:           pb.LoginType_LoginType_SIGNUP,
			prevLoginType:       pb.LoginType_LoginType_UNKNOWN,
			expectedStatus:      pb.LoginStatus_NO_NET,
			expectedEventsCount: 1,
			validateEvents: func(t *testing.T, evts []*events.DataAuthorization) {
				assert.Equal(t, events.StatusFailure, evts[0].EventStatus)
				assert.Equal(t, events.TriggerUser, evts[0].EventTrigger)
				assert.Equal(t, events.LoginSignUp, evts[0].EventType)
				assert.Equal(t, events.ReasonLoginURLRetrieveFailed, evts[0].Reason)
			},
		},
		{
			name: "unknown OAuth2 error",
			setup: func(em *daemonevents.MockPublisherSubscriber[events.DataAuthorization]) *RPC {
				return &RPC{
					consentChecker: &mock.AnalyticsConsentCheckerMock{ConsentCompleted: true},
					ac:             &testauth.AuthCheckerMock{LoggedIn: false},
					events:         &daemonevents.Events{User: &daemonevents.LoginEvents{Login: em}},
					authentication: &testcore.AuthenticationAPImock{
						LoginError: errors.New("some other error"),
					},
					initialLoginType: NewAtomicLoginType(),
				}
			},
			loginType:           pb.LoginType_LoginType_LOGIN,
			prevLoginType:       pb.LoginType_LoginType_UNKNOWN,
			expectedStatus:      pb.LoginStatus_UNKNOWN_OAUTH2_ERROR,
			expectedEventsCount: 1,
			validateEvents: func(t *testing.T, evts []*events.DataAuthorization) {
				assert.Equal(t, events.StatusFailure, evts[0].EventStatus)
				assert.Equal(t, events.ReasonLoginURLRetrieveFailed, evts[0].Reason)
			},
		},
		{
			name: "unfinished previous login - login after login",
			setup: func(em *daemonevents.MockPublisherSubscriber[events.DataAuthorization]) *RPC {
				return &RPC{
					consentChecker:   &mock.AnalyticsConsentCheckerMock{ConsentCompleted: true},
					ac:               &testauth.AuthCheckerMock{LoggedIn: false},
					events:           &daemonevents.Events{User: &daemonevents.LoginEvents{Login: em}},
					authentication:   &testcore.AuthenticationAPImock{},
					initialLoginType: NewAtomicLoginType(),
				}
			},
			loginType:           pb.LoginType_LoginType_LOGIN,
			prevLoginType:       pb.LoginType_LoginType_LOGIN,
			expectedStatus:      pb.LoginStatus_SUCCESS,
			expectedEventsCount: 2,
			validateEvents: func(t *testing.T, evts []*events.DataAuthorization) {
				// First event: unfinished previous login
				assert.Equal(t, events.StatusFailure, evts[0].EventStatus)
				assert.Equal(t, events.TriggerUser, evts[0].EventTrigger)
				assert.Equal(t, events.LoginLogin, evts[0].EventType)
				assert.Equal(t, events.ReasonUnfinishedPrevLogin, evts[0].Reason)
				assert.Equal(t, -1, evts[0].DurationMs)

				// Second event: current login attempt
				assert.Equal(t, events.StatusAttempt, evts[1].EventStatus)
				assert.Equal(t, events.TriggerUser, evts[1].EventTrigger)
				assert.Equal(t, events.LoginLogin, evts[1].EventType)
				assert.Equal(t, events.ReasonNotSpecified, evts[1].Reason)
				assert.Equal(t, -1, evts[1].DurationMs)
			},
		},
		{
			name: "unfinished previous signup - login after signup",
			setup: func(em *daemonevents.MockPublisherSubscriber[events.DataAuthorization]) *RPC {
				return &RPC{
					consentChecker:   &mock.AnalyticsConsentCheckerMock{ConsentCompleted: true},
					ac:               &testauth.AuthCheckerMock{LoggedIn: false},
					events:           &daemonevents.Events{User: &daemonevents.LoginEvents{Login: em}},
					authentication:   &testcore.AuthenticationAPImock{},
					initialLoginType: NewAtomicLoginType(),
				}
			},
			loginType:           pb.LoginType_LoginType_LOGIN,
			prevLoginType:       pb.LoginType_LoginType_SIGNUP,
			expectedStatus:      pb.LoginStatus_SUCCESS,
			expectedEventsCount: 2,
			validateEvents: func(t *testing.T, evts []*events.DataAuthorization) {
				// First event: unfinished previous signup
				assert.Equal(t, events.StatusFailure, evts[0].EventStatus)
				assert.Equal(t, events.TriggerUser, evts[0].EventTrigger)
				assert.Equal(t, events.LoginLogin, evts[0].EventType)
				assert.Equal(t, events.ReasonUnfinishedPrevLogin, evts[0].Reason)

				// Second event: current login attempt
				assert.Equal(t, events.StatusAttempt, evts[1].EventStatus)
				assert.Equal(t, events.LoginLogin, evts[1].EventType)
			},
		},
		{
			name: "successful login without previous unfinished login",
			setup: func(em *daemonevents.MockPublisherSubscriber[events.DataAuthorization]) *RPC {
				return &RPC{
					consentChecker:   &mock.AnalyticsConsentCheckerMock{ConsentCompleted: true},
					ac:               &testauth.AuthCheckerMock{LoggedIn: false},
					events:           &daemonevents.Events{User: &daemonevents.LoginEvents{Login: em}},
					authentication:   &testcore.AuthenticationAPImock{},
					initialLoginType: NewAtomicLoginType(),
				}
			},
			loginType:           pb.LoginType_LoginType_LOGIN,
			prevLoginType:       pb.LoginType_LoginType_UNKNOWN,
			expectedStatus:      pb.LoginStatus_SUCCESS,
			expectedEventsCount: 1,
			validateEvents: func(t *testing.T, evts []*events.DataAuthorization) {
				assert.Equal(t, events.StatusAttempt, evts[0].EventStatus)
				assert.Equal(t, events.TriggerUser, evts[0].EventTrigger)
				assert.Equal(t, events.LoginLogin, evts[0].EventType)
				assert.Equal(t, events.ReasonNotSpecified, evts[0].Reason)
				assert.Equal(t, -1, evts[0].DurationMs)
			},
		},
		{
			name: "successful signup without previous unfinished login",
			setup: func(em *daemonevents.MockPublisherSubscriber[events.DataAuthorization]) *RPC {
				return &RPC{
					consentChecker:   &mock.AnalyticsConsentCheckerMock{ConsentCompleted: true},
					ac:               &testauth.AuthCheckerMock{LoggedIn: false},
					events:           &daemonevents.Events{User: &daemonevents.LoginEvents{Login: em}},
					authentication:   &testcore.AuthenticationAPImock{},
					initialLoginType: NewAtomicLoginType(),
				}
			},
			loginType:           pb.LoginType_LoginType_SIGNUP,
			prevLoginType:       pb.LoginType_LoginType_UNKNOWN,
			expectedStatus:      pb.LoginStatus_SUCCESS,
			expectedEventsCount: 1,
			validateEvents: func(t *testing.T, evts []*events.DataAuthorization) {
				assert.Equal(t, events.StatusAttempt, evts[0].EventStatus)
				assert.Equal(t, events.TriggerUser, evts[0].EventTrigger)
				assert.Equal(t, events.LoginSignUp, evts[0].EventType)
				assert.Equal(t, events.ReasonNotSpecified, evts[0].Reason)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var publishedEvents []*events.DataAuthorization
			eventsMock := &daemonevents.MockPublisherSubscriber[events.DataAuthorization]{
				Handler: func(event events.DataAuthorization) error {
					eventCopy := event
					publishedEvents = append(publishedEvents, &eventCopy)
					return nil
				},
			}

			r := tt.setup(eventsMock)
			r.initialLoginType.set(tt.prevLoginType)

			resp, err := r.LoginOAuth2(context.Background(), &pb.LoginOAuth2Request{Type: tt.loginType})

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.Status)

			assert.Equal(t, tt.expectedEventsCount, len(publishedEvents),
				"Expected %d events to be published, but got %d",
				tt.expectedEventsCount, len(publishedEvents))

			if tt.validateEvents != nil && len(publishedEvents) > 0 {
				tt.validateEvents(t, publishedEvents)
			}
		})
	}
}

func TestLoginOAuth2Callback_AnalyticsEvents(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name                string
		setup               func(*daemonevents.MockPublisherSubscriber[events.DataAuthorization]) *RPC
		token               string
		loginType           pb.LoginType
		prevLoginType       pb.LoginType
		expectedStatus      pb.LoginStatus
		expectedError       error
		expectedEventsCount int
		validateEvents      func(*testing.T, []*events.DataAuthorization)
	}{
		{
			name: "missing exchange token",
			setup: func(em *daemonevents.MockPublisherSubscriber[events.DataAuthorization]) *RPC {
				return &RPC{
					consentChecker:   &mock.AnalyticsConsentCheckerMock{ConsentCompleted: true},
					ac:               &testauth.AuthCheckerMock{LoggedIn: false},
					events:           &daemonevents.Events{User: &daemonevents.LoginEvents{Login: em}},
					publisher:        &subs.Subject[string]{},
					initialLoginType: NewAtomicLoginType(),
				}
			},
			token:               "",
			loginType:           pb.LoginType_LoginType_LOGIN,
			prevLoginType:       pb.LoginType_LoginType_LOGIN,
			expectedError:       errors.New("The exchange token is missing. Please try logging in again. If the issue persists, contact our customer support."),
			expectedEventsCount: 1,
			validateEvents: func(t *testing.T, evts []*events.DataAuthorization) {
				assert.Equal(t, events.StatusFailure, evts[0].EventStatus)
				assert.Equal(t, events.TriggerUser, evts[0].EventTrigger)
				assert.Equal(t, events.LoginLogin, evts[0].EventType)
				assert.Equal(t, events.ReasonLoginExchangeTokenMissing, evts[0].Reason)
				assert.False(t, evts[0].IsAlteredFlowOnNordAccount)
			},
		},
		{
			name: "exchange token failed",
			setup: func(em *daemonevents.MockPublisherSubscriber[events.DataAuthorization]) *RPC {
				return &RPC{
					consentChecker:   &mock.AnalyticsConsentCheckerMock{ConsentCompleted: true},
					ac:               &testauth.AuthCheckerMock{LoggedIn: false},
					events:           &daemonevents.Events{User: &daemonevents.LoginEvents{Login: em}},
					authentication:   &testcore.AuthenticationAPImock{TokenError: errors.New("token exchange failed")},
					publisher:        &subs.Subject[string]{},
					initialLoginType: NewAtomicLoginType(),
				}
			},
			token:               "exchange-token",
			loginType:           pb.LoginType_LoginType_LOGIN,
			prevLoginType:       pb.LoginType_LoginType_LOGIN,
			expectedError:       errors.New("token exchange failed"),
			expectedEventsCount: 1,
			validateEvents: func(t *testing.T, evts []*events.DataAuthorization) {
				assert.Equal(t, events.StatusFailure, evts[0].EventStatus)
				assert.Equal(t, events.ReasonLoginExchangeTokenFailed, evts[0].Reason)
			},
		},
		{
			name: "get user info failed",
			setup: func(em *daemonevents.MockPublisherSubscriber[events.DataAuthorization]) *RPC {
				return &RPC{
					consentChecker:   &mock.AnalyticsConsentCheckerMock{ConsentCompleted: true},
					ac:               &testauth.AuthCheckerMock{LoggedIn: false},
					events:           &daemonevents.Events{User: &daemonevents.LoginEvents{Login: em}},
					authentication:   &testcore.AuthenticationAPImock{},
					credentialsAPI:   &testcore.CredentialsAPIMock{ServiceCredentialsErr: errors.New("failed to get credentials")},
					publisher:        &subs.Subject[string]{},
					initialLoginType: NewAtomicLoginType(),
				}
			},
			token:               "exchange-token",
			loginType:           pb.LoginType_LoginType_LOGIN,
			prevLoginType:       pb.LoginType_LoginType_LOGIN,
			expectedError:       errors.New("failed to get credentials"),
			expectedEventsCount: 1,
			validateEvents: func(t *testing.T, evts []*events.DataAuthorization) {
				assert.Equal(t, events.StatusFailure, evts[0].EventStatus)
				assert.Equal(t, events.ReasonLoginGetUserInfoFailed, evts[0].Reason)
			},
		},
		{
			name: "successful login callback - unaltered flow",
			setup: func(em *daemonevents.MockPublisherSubscriber[events.DataAuthorization]) *RPC {
				return &RPC{
					consentChecker:   &mock.AnalyticsConsentCheckerMock{ConsentCompleted: true},
					ac:               &testauth.AuthCheckerMock{LoggedIn: false},
					cm:               mock.NewMockConfigManager(),
					events:           &daemonevents.Events{User: &daemonevents.LoginEvents{Login: em}},
					authentication:   &testcore.AuthenticationAPImock{},
					credentialsAPI:   &testcore.CredentialsAPIMock{},
					initialLoginType: NewAtomicLoginType(),
					ncClient:         &mock.NotificationClientMock{},
				}
			},
			token:               "exchange-token",
			loginType:           pb.LoginType_LoginType_LOGIN,
			prevLoginType:       pb.LoginType_LoginType_LOGIN,
			expectedStatus:      pb.LoginStatus_SUCCESS,
			expectedEventsCount: 1,
			validateEvents: func(t *testing.T, evts []*events.DataAuthorization) {
				assert.Equal(t, events.StatusSuccess, evts[0].EventStatus)
				assert.Equal(t, events.TriggerUser, evts[0].EventTrigger)
				assert.Equal(t, events.LoginLogin, evts[0].EventType)
				assert.Equal(t, events.ReasonNotSpecified, evts[0].Reason)
				assert.False(t, evts[0].IsAlteredFlowOnNordAccount)
				assert.Greater(t, evts[0].DurationMs, 0)
			},
		},
		{
			name: "successful signup callback - altered flow",
			setup: func(em *daemonevents.MockPublisherSubscriber[events.DataAuthorization]) *RPC {
				return &RPC{
					consentChecker:   &mock.AnalyticsConsentCheckerMock{ConsentCompleted: true},
					ac:               &testauth.AuthCheckerMock{LoggedIn: false},
					cm:               mock.NewMockConfigManager(),
					events:           &daemonevents.Events{User: &daemonevents.LoginEvents{Login: em}},
					authentication:   &testcore.AuthenticationAPImock{},
					credentialsAPI:   &testcore.CredentialsAPIMock{},
					initialLoginType: NewAtomicLoginType(),
					ncClient:         &mock.NotificationClientMock{},
				}
			},
			token:               "exchange-token",
			loginType:           pb.LoginType_LoginType_LOGIN,
			prevLoginType:       pb.LoginType_LoginType_SIGNUP,
			expectedStatus:      pb.LoginStatus_SUCCESS,
			expectedEventsCount: 1,
			validateEvents: func(t *testing.T, evts []*events.DataAuthorization) {
				assert.Equal(t, events.StatusSuccess, evts[0].EventStatus)
				assert.Equal(t, events.TriggerUser, evts[0].EventTrigger)
				assert.Equal(t, events.LoginLogin, evts[0].EventType)
				assert.Equal(t, events.ReasonNotSpecified, evts[0].Reason)
				assert.True(t, evts[0].IsAlteredFlowOnNordAccount)
			},
		},
		{
			name: "config save error",
			setup: func(em *daemonevents.MockPublisherSubscriber[events.DataAuthorization]) *RPC {
				cm := mock.NewMockConfigManager()
				cm.SaveErr = errors.New("config save error")
				return &RPC{
					consentChecker:   &mock.AnalyticsConsentCheckerMock{ConsentCompleted: true},
					ac:               &testauth.AuthCheckerMock{LoggedIn: false},
					cm:               cm,
					events:           &daemonevents.Events{User: &daemonevents.LoginEvents{Login: em}},
					authentication:   &testcore.AuthenticationAPImock{},
					credentialsAPI:   &testcore.CredentialsAPIMock{},
					initialLoginType: NewAtomicLoginType(),
					ncClient:         &mock.NotificationClientMock{},
				}
			},
			token:               "exchange-token",
			loginType:           pb.LoginType_LoginType_LOGIN,
			prevLoginType:       pb.LoginType_LoginType_LOGIN,
			expectedError:       errors.New("config save error"),
			expectedEventsCount: 1,
			validateEvents: func(t *testing.T, evts []*events.DataAuthorization) {
				assert.Equal(t, events.StatusFailure, evts[0].EventStatus)
				assert.Equal(t, events.ReasonNotSpecified, evts[0].Reason)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var publishedEvents []*events.DataAuthorization
			eventsMock := &daemonevents.MockPublisherSubscriber[events.DataAuthorization]{
				Handler: func(event events.DataAuthorization) error {
					eventCopy := event
					publishedEvents = append(publishedEvents, &eventCopy)
					return nil
				},
			}

			r := tt.setup(eventsMock)
			r.initialLoginType.set(tt.prevLoginType)

			resp, err := r.LoginOAuth2Callback(context.Background(), &pb.LoginOAuth2CallbackRequest{
				Token: tt.token,
				Type:  tt.loginType,
			})

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, resp.Status)
			}

			assert.Equal(t, tt.expectedEventsCount, len(publishedEvents),
				"Expected %d events to be published, but got %d",
				tt.expectedEventsCount, len(publishedEvents))

			if tt.validateEvents != nil && len(publishedEvents) > 0 {
				tt.validateEvents(t, publishedEvents)
			}

			// Verify that initialLoginType is reset at the end
			assert.Equal(t, pb.LoginType_LoginType_UNKNOWN, r.initialLoginType.get())
		})
	}
}

func TestLoginWithTokenInputValidation(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name            string
		setup           func() *RPC
		token           string
		expectedPayload *pb.LoginResponse
		expectedError   error
	}{
		{
			name: "empty token",
			setup: func() *RPC {
				return &RPC{
					consentChecker: &mock.AnalyticsConsentCheckerMock{ConsentCompleted: true},
				}
			},
			token:           "",
			expectedPayload: &pb.LoginResponse{Type: internal.CodeTokenLoginFailure},
			expectedError:   nil,
		},
		{
			name: "invalid token format - contains non-hex characters",
			setup: func() *RPC {
				return &RPC{
					consentChecker: &mock.AnalyticsConsentCheckerMock{ConsentCompleted: true},
				}
			},
			token:           "invalid-token-with-dashes",
			expectedPayload: &pb.LoginResponse{Type: internal.CodeTokenInvalid},
			expectedError:   nil,
		},
		{
			name: "invalid token format - uppercase letters",
			setup: func() *RPC {
				return &RPC{
					consentChecker: &mock.AnalyticsConsentCheckerMock{ConsentCompleted: true},
				}
			},
			token:           "ABCDEF123456",
			expectedPayload: &pb.LoginResponse{Type: internal.CodeTokenInvalid},
			expectedError:   nil,
		},
		{
			name: "valid token format - successful login",
			setup: func() *RPC {
				eventsMock := &daemonevents.MockPublisherSubscriber[events.DataAuthorization]{}
				return &RPC{
					consentChecker:   &mock.AnalyticsConsentCheckerMock{ConsentCompleted: true},
					ac:               &testauth.AuthCheckerMock{LoggedIn: false},
					cm:               mock.NewMockConfigManager(),
					credentialsAPI:   &testcore.CredentialsAPIMock{},
					events:           &daemonevents.Events{User: &daemonevents.LoginEvents{Login: eventsMock}},
					publisher:        &subs.Subject[string]{},
					ncClient:         &mock.NotificationClientMock{},
					initialLoginType: NewAtomicLoginType(),
				}
			},
			token:           "abcdef123456",
			expectedPayload: &pb.LoginResponse{Type: internal.CodeSuccess},
			expectedError:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.setup()

			resp, err := r.LoginWithToken(context.Background(), &pb.LoginWithTokenRequest{
				Token: tt.token,
			})

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectedPayload != nil {
				assert.NotNil(t, resp)
				assert.Equal(t, tt.expectedPayload.Type, resp.Type)
			} else {
				assert.Nil(t, resp)
			}
		})
	}
}

func TestLoginWithToken(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name                   string
		setup                  func(*daemonevents.MockPublisherSubscriber[events.DataAuthorization]) *RPC
		token                  string
		expectedPayload        *pb.LoginResponse
		expectedError          error
		expectedPublishedCount int
		validateEvents         func(*testing.T, []*events.DataAuthorization)
	}{
		{
			name: "already logged in",
			setup: func(em *daemonevents.MockPublisherSubscriber[events.DataAuthorization]) *RPC {
				return &RPC{
					ac:               &testauth.AuthCheckerMock{LoggedIn: true},
					events:           &daemonevents.Events{User: &daemonevents.LoginEvents{Login: em}},
					initialLoginType: NewAtomicLoginType(),
				}
			},
			token:                  "test-token",
			expectedPayload:        nil,
			expectedError:          internal.ErrAlreadyLoggedIn,
			expectedPublishedCount: 0,
		},
		{
			name: "successful login",
			setup: func(em *daemonevents.MockPublisherSubscriber[events.DataAuthorization]) *RPC {
				return &RPC{
					ac:               &testauth.AuthCheckerMock{LoggedIn: false},
					cm:               mock.NewMockConfigManager(),
					credentialsAPI:   &testcore.CredentialsAPIMock{},
					events:           &daemonevents.Events{User: &daemonevents.LoginEvents{Login: em}},
					publisher:        &subs.Subject[string]{},
					ncClient:         &mock.NotificationClientMock{},
					initialLoginType: NewAtomicLoginType(),
				}
			},
			token:                  "test-token",
			expectedPayload:        &pb.LoginResponse{Type: internal.CodeSuccess},
			expectedError:          nil,
			expectedPublishedCount: 2,
			validateEvents: func(t *testing.T, evts []*events.DataAuthorization) {
				// First event should be attempt
				assert.Equal(t, events.StatusAttempt, evts[0].EventStatus)
				assert.Equal(t, events.TriggerUser, evts[0].EventTrigger)
				assert.Equal(t, events.LoginLogin, evts[0].EventType)
				assert.Equal(t, -1, evts[0].DurationMs)
				assert.Equal(t, events.ReasonNotSpecified, evts[0].Reason)

				// Second event should be success
				assert.Equal(t, events.StatusSuccess, evts[1].EventStatus)
				assert.Equal(t, events.TriggerUser, evts[1].EventTrigger)
				assert.Equal(t, events.LoginLogin, evts[1].EventType)
				assert.GreaterOrEqual(t, evts[1].DurationMs, 0)
				assert.Equal(t, events.ReasonNotSpecified, evts[1].Reason)
			},
		},
		{
			name: "credentials API returns server internal error",
			setup: func(em *daemonevents.MockPublisherSubscriber[events.DataAuthorization]) *RPC {
				return &RPC{
					ac:               &testauth.AuthCheckerMock{LoggedIn: false},
					credentialsAPI:   &testcore.CredentialsAPIMock{ServiceCredentialsErr: core.ErrServerInternal},
					events:           &daemonevents.Events{User: &daemonevents.LoginEvents{Login: em}},
					initialLoginType: NewAtomicLoginType(),
				}
			},
			token:                  "test-token",
			expectedPayload:        &pb.LoginResponse{Type: internal.CodeInternalError},
			expectedError:          nil,
			expectedPublishedCount: 2,
			validateEvents: func(t *testing.T, evts []*events.DataAuthorization) {
				assert.Equal(t, events.StatusFailure, evts[1].EventStatus)
				assert.Equal(t, events.ReasonLoginGetUserInfoFailed, evts[1].Reason)
			},
		},
		{
			name: "credentials API returns unauthorized error",
			setup: func(em *daemonevents.MockPublisherSubscriber[events.DataAuthorization]) *RPC {
				return &RPC{
					ac:               &testauth.AuthCheckerMock{LoggedIn: false},
					credentialsAPI:   &testcore.CredentialsAPIMock{ServiceCredentialsErr: core.ErrUnauthorized},
					events:           &daemonevents.Events{User: &daemonevents.LoginEvents{Login: em}},
					initialLoginType: NewAtomicLoginType(),
				}
			},
			token:                  "test-token",
			expectedPayload:        &pb.LoginResponse{Type: internal.CodeTokenInvalid},
			expectedError:          nil,
			expectedPublishedCount: 2,
			validateEvents: func(t *testing.T, evts []*events.DataAuthorization) {
				assert.Equal(t, events.StatusFailure, evts[1].EventStatus)
				assert.Equal(t, events.ReasonLoginGetUserInfoFailed, evts[1].Reason)
			},
		},
		{
			name: "credentials API returns generic error",
			setup: func(em *daemonevents.MockPublisherSubscriber[events.DataAuthorization]) *RPC {
				return &RPC{
					ac:               &testauth.AuthCheckerMock{LoggedIn: false},
					credentialsAPI:   &testcore.CredentialsAPIMock{ServiceCredentialsErr: errors.New("generic error")},
					events:           &daemonevents.Events{User: &daemonevents.LoginEvents{Login: em}},
					initialLoginType: NewAtomicLoginType(),
				}
			},
			token:                  "test-token",
			expectedPayload:        &pb.LoginResponse{Type: internal.CodeGatewayError},
			expectedError:          nil,
			expectedPublishedCount: 2,
			validateEvents: func(t *testing.T, evts []*events.DataAuthorization) {
				assert.Equal(t, events.StatusFailure, evts[1].EventStatus)
				assert.Equal(t, events.ReasonLoginGetUserInfoFailed, evts[1].Reason)
			},
		},
		{
			name: "config load error",
			setup: func(em *daemonevents.MockPublisherSubscriber[events.DataAuthorization]) *RPC {
				cm := mock.NewMockConfigManager()
				cm.LoadErr = errors.New("config load error")
				return &RPC{
					ac:               &testauth.AuthCheckerMock{LoggedIn: false},
					cm:               cm,
					credentialsAPI:   &testcore.CredentialsAPIMock{},
					events:           &daemonevents.Events{User: &daemonevents.LoginEvents{Login: em}},
					initialLoginType: NewAtomicLoginType(),
				}
			},
			token:                  "test-token",
			expectedPayload:        &pb.LoginResponse{Type: internal.CodeConfigError},
			expectedError:          nil,
			expectedPublishedCount: 2,
			validateEvents: func(t *testing.T, evts []*events.DataAuthorization) {
				assert.Equal(t, events.StatusFailure, evts[1].EventStatus)
				assert.Equal(t, events.ReasonNotSpecified, evts[1].Reason)
			},
		},
		{
			name: "config save error",
			setup: func(em *daemonevents.MockPublisherSubscriber[events.DataAuthorization]) *RPC {
				cm := mock.NewMockConfigManager()
				cm.SaveErr = errors.New("config save error")
				return &RPC{
					ac:               &testauth.AuthCheckerMock{LoggedIn: false},
					cm:               cm,
					credentialsAPI:   &testcore.CredentialsAPIMock{},
					events:           &daemonevents.Events{User: &daemonevents.LoginEvents{Login: em}},
					initialLoginType: NewAtomicLoginType(),
				}
			},
			token:                  "test-token",
			expectedPayload:        &pb.LoginResponse{Type: internal.CodeConfigError},
			expectedError:          nil,
			expectedPublishedCount: 2,
			validateEvents: func(t *testing.T, evts []*events.DataAuthorization) {
				assert.Equal(t, events.StatusFailure, evts[1].EventStatus)
				assert.Equal(t, events.ReasonNotSpecified, evts[1].Reason)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var publishedEvents []*events.DataAuthorization
			eventsMock := &daemonevents.MockPublisherSubscriber[events.DataAuthorization]{}
			eventsMock.Handler = func(event events.DataAuthorization) error {
				eventCopy := event
				publishedEvents = append(publishedEvents, &eventCopy)
				return nil
			}

			r := tt.setup(eventsMock)

			payload, err := r.loginWithToken(tt.token)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectedPayload != nil {
				assert.Equal(t, tt.expectedPayload.Type, payload.Type)
			} else {
				assert.Nil(t, payload)
			}

			assert.Equal(t, tt.expectedPublishedCount, len(publishedEvents),
				"Expected %d events to be published, but got %d",
				tt.expectedPublishedCount, len(publishedEvents))

			if tt.validateEvents != nil && len(publishedEvents) > 0 {
				tt.validateEvents(t, publishedEvents)
			}

			if tt.expectedPublishedCount > 0 {
				assert.True(t, eventsMock.EventPublished, "Events should have been published")
			}
		})
	}
}

// TestLoginWithToken_AlteredFlowFlag tests that loginWithToken does not incorrectly
// report IsAlteredFlowOnNordAccount=true when there's stale state from a previous
// failed OAuth2 flow. Token login is a standalone flow and should never report altered flow.
func TestLoginWithToken_AlteredFlowFlag(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name                       string
		prevLoginType              pb.LoginType
		expectedAlteredFlowOnEvent bool
	}{
		{
			name:                       "no previous login - altered flow should be false",
			prevLoginType:              pb.LoginType_LoginType_UNKNOWN,
			expectedAlteredFlowOnEvent: false,
		},
		{
			name:                       "previous LOGIN attempt - altered flow should be false",
			prevLoginType:              pb.LoginType_LoginType_LOGIN,
			expectedAlteredFlowOnEvent: false,
		},
		{
			name:                       "previous SIGNUP attempt - altered flow should be false (bug: was true)",
			prevLoginType:              pb.LoginType_LoginType_SIGNUP,
			expectedAlteredFlowOnEvent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var publishedEvents []*events.DataAuthorization
			eventsMock := &daemonevents.MockPublisherSubscriber[events.DataAuthorization]{
				Handler: func(event events.DataAuthorization) error {
					eventCopy := event
					publishedEvents = append(publishedEvents, &eventCopy)
					return nil
				},
			}

			r := &RPC{
				ac:               &testauth.AuthCheckerMock{LoggedIn: false},
				cm:               mock.NewMockConfigManager(),
				credentialsAPI:   &testcore.CredentialsAPIMock{},
				events:           &daemonevents.Events{User: &daemonevents.LoginEvents{Login: eventsMock}},
				publisher:        &subs.Subject[string]{},
				ncClient:         &mock.NotificationClientMock{},
				initialLoginType: NewAtomicLoginType(),
			}

			// Simulate stale state from a previous failed OAuth2 flow
			if tt.prevLoginType != pb.LoginType_LoginType_UNKNOWN {
				r.initialLoginType.set(tt.prevLoginType)
			}

			payload, err := r.loginWithToken("test-token")

			assert.NoError(t, err)
			assert.Equal(t, internal.CodeSuccess, payload.Type)

			// Find the success event (last event) and check altered flow flag
			assert.GreaterOrEqual(t, len(publishedEvents), 1, "Expected at least 1 event")

			// Check all events have correct altered flow flag
			for i, evt := range publishedEvents {
				assert.Equal(t, tt.expectedAlteredFlowOnEvent, evt.IsAlteredFlowOnNordAccount,
					"Event %d should have IsAlteredFlowOnNordAccount=%v, got %v",
					i, tt.expectedAlteredFlowOnEvent, evt.IsAlteredFlowOnNordAccount)
			}

			// Verify initialLoginType is reset after login
			assert.Equal(t, pb.LoginType_LoginType_UNKNOWN, r.initialLoginType.get(),
				"initialLoginType should be reset after login")
		})
	}
}

// TestOAuth2LoginCallbackFailure_ThenRetryLogin_AlteredFlowShouldBeFalse tests the scenario:
// 1. User starts OAuth2 LOGIN flow -> initialLoginType set to LOGIN
// 2. OAuth2 callback fails with ReasonLoginExchangeTokenMissing or ReasonLoginExchangeTokenFailed
// 3. initialLoginType is reset after callback (even on failure)
// 4. User retries OAuth2 LOGIN -> should NOT report altered flow
func TestOAuth2LoginCallbackFailure_ThenRetryLogin_AlteredFlowShouldBeFalse(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name           string
		callbackToken  string
		tokenError     error
		expectedReason events.ReasonCode
	}{
		{
			name:           "callback fails with missing exchange token",
			callbackToken:  "", // empty token causes ReasonLoginExchangeTokenMissing
			tokenError:     nil,
			expectedReason: events.ReasonLoginExchangeTokenMissing,
		},
		{
			name:           "callback fails with exchange token error",
			callbackToken:  "some-token",
			tokenError:     errors.New("token exchange failed"),
			expectedReason: events.ReasonLoginExchangeTokenFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var publishedEvents []*events.DataAuthorization
			eventsMock := &daemonevents.MockPublisherSubscriber[events.DataAuthorization]{
				Handler: func(event events.DataAuthorization) error {
					eventCopy := event
					publishedEvents = append(publishedEvents, &eventCopy)
					return nil
				},
			}

			r := &RPC{
				consentChecker: &mock.AnalyticsConsentCheckerMock{ConsentCompleted: true},
				ac:             &testauth.AuthCheckerMock{LoggedIn: false},
				cm:             mock.NewMockConfigManager(),
				credentialsAPI: &testcore.CredentialsAPIMock{},
				events:         &daemonevents.Events{User: &daemonevents.LoginEvents{Login: eventsMock}},
				authentication: &testcore.AuthenticationAPImock{
					TokenError: tt.tokenError,
				},
				publisher:        &subs.Subject[string]{},
				ncClient:         &mock.NotificationClientMock{},
				initialLoginType: NewAtomicLoginType(),
			}

			// Step 1: User initiates OAuth2 LOGIN flow
			resp, err := r.LoginOAuth2(context.Background(), &pb.LoginOAuth2Request{
				Type: pb.LoginType_LoginType_LOGIN,
			})
			assert.NoError(t, err)
			assert.Equal(t, pb.LoginStatus_SUCCESS, resp.Status)
			assert.Equal(t, pb.LoginType_LoginType_LOGIN, r.initialLoginType.get())

			// Step 2: OAuth2 callback fails
			_, err = r.LoginOAuth2Callback(context.Background(), &pb.LoginOAuth2CallbackRequest{
				Token: tt.callbackToken,
				Type:  pb.LoginType_LoginType_LOGIN,
			})
			assert.Error(t, err) // callback should fail

			// Verify the failure event has correct reason and NO altered flow
			// (LOGIN started, LOGIN callback - same type, not altered)
			var failureEvent *events.DataAuthorization
			for _, evt := range publishedEvents {
				if evt.EventStatus == events.StatusFailure && evt.Reason == tt.expectedReason {
					failureEvent = evt
					break
				}
			}
			assert.NotNil(t, failureEvent, "Expected failure event with reason %v", tt.expectedReason)
			assert.False(t, failureEvent.IsAlteredFlowOnNordAccount,
				"Callback failure event should have IsAlteredFlowOnNordAccount=false when LOGIN->LOGIN")

			// Step 3: Verify initialLoginType is reset after callback (even on failure)
			assert.Equal(t, pb.LoginType_LoginType_UNKNOWN, r.initialLoginType.get(),
				"initialLoginType should be reset after callback, even on failure")

			// Clear events
			publishedEvents = nil

			// Step 4: User retries OAuth2 LOGIN
			resp2, err := r.LoginOAuth2(context.Background(), &pb.LoginOAuth2Request{
				Type: pb.LoginType_LoginType_LOGIN,
			})
			assert.NoError(t, err)
			assert.Equal(t, pb.LoginStatus_SUCCESS, resp2.Status)

			// Verify retry events do NOT report altered flow
			for i, evt := range publishedEvents {
				assert.False(t, evt.IsAlteredFlowOnNordAccount,
					"Retry event %d should have IsAlteredFlowOnNordAccount=false", i)
			}
		})
	}
}

// TestOAuth2LoginFailure_InitialLoginTypeNotReset tests that when OAuth2 login flow
// fails (e.g., user abandons OAuth in browser), the initialLoginType remains set,
// and a subsequent loginWithToken should NOT report IsAlteredFlowOnNordAccount=true.
// This test reproduces the bug scenario:
// 1. User starts OAuth2 SIGNUP flow -> initialLoginType set to SIGNUP
// 2. OAuth2 callback never happens (user closes browser) or fails
// 3. User tries token login -> should NOT report altered flow
func TestOAuth2LoginFailure_ThenTokenLogin_AlteredFlowShouldBeFalse(t *testing.T) {
	category.Set(t, category.Unit)

	var publishedEvents []*events.DataAuthorization
	eventsMock := &daemonevents.MockPublisherSubscriber[events.DataAuthorization]{
		Handler: func(event events.DataAuthorization) error {
			eventCopy := event
			publishedEvents = append(publishedEvents, &eventCopy)
			return nil
		},
	}

	r := &RPC{
		consentChecker:   &mock.AnalyticsConsentCheckerMock{ConsentCompleted: true},
		ac:               &testauth.AuthCheckerMock{LoggedIn: false},
		cm:               mock.NewMockConfigManager(),
		credentialsAPI:   &testcore.CredentialsAPIMock{},
		events:           &daemonevents.Events{User: &daemonevents.LoginEvents{Login: eventsMock}},
		authentication:   &testcore.AuthenticationAPImock{}, // successful URL retrieval
		publisher:        &subs.Subject[string]{},
		ncClient:         &mock.NotificationClientMock{},
		initialLoginType: NewAtomicLoginType(),
	}

	// Step 1: User initiates OAuth2 SIGNUP flow
	resp, err := r.LoginOAuth2(context.Background(), &pb.LoginOAuth2Request{
		Type: pb.LoginType_LoginType_SIGNUP,
	})
	assert.NoError(t, err)
	assert.Equal(t, pb.LoginStatus_SUCCESS, resp.Status)

	// Verify initialLoginType is now set to SIGNUP
	assert.Equal(t, pb.LoginType_LoginType_SIGNUP, r.initialLoginType.get(),
		"initialLoginType should be SIGNUP after OAuth2 flow started")

	// Clear events from OAuth2 setup
	publishedEvents = nil

	// Step 2: User abandons OAuth (callback never happens)
	// initialLoginType remains SIGNUP - this is the stale state

	// Step 3: User tries token login instead
	payload, err := r.loginWithToken("test-token")
	assert.NoError(t, err)
	assert.Equal(t, internal.CodeSuccess, payload.Type)

	// Verify: All events from token login should have IsAlteredFlowOnNordAccount=false
	// because token login is a standalone flow, not a continuation of OAuth2
	assert.GreaterOrEqual(t, len(publishedEvents), 1, "Expected at least 1 event from token login")

	for i, evt := range publishedEvents {
		assert.False(t, evt.IsAlteredFlowOnNordAccount,
			"Event %d from token login should have IsAlteredFlowOnNordAccount=false, got true. "+
				"Token login is a standalone flow and should not inherit altered state from abandoned OAuth2 flow.",
			i)
	}

	// Verify initialLoginType is reset after token login
	assert.Equal(t, pb.LoginType_LoginType_UNKNOWN, r.initialLoginType.get(),
		"initialLoginType should be reset after token login")
}

// TestOAuth2Callback_AlteredFlowWhenInitialLoginTypeUnknown tests the scenario
// where LoginOAuth2Callback is called without a preceding LoginOAuth2 call
// (or after a previous flow fully completed and reset the state).
// In this case initialLoginType is UNKNOWN, and isAltered(LOGIN) or
// isAltered(SIGNUP) both return true because UNKNOWN != LOGIN/SIGNUP.
// The expected correct behavior is IsAlteredFlowOnNordAccount=false,
// since the flow was never started — there's nothing to alter.
func TestOAuth2Callback_AlteredFlowWhenInitialLoginTypeUnknown(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name      string
		loginType pb.LoginType
	}{
		{
			name:      "callback with LOGIN type without prior OAuth2 start",
			loginType: pb.LoginType_LoginType_LOGIN,
		},
		{
			name:      "callback with SIGNUP type without prior OAuth2 start",
			loginType: pb.LoginType_LoginType_SIGNUP,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var publishedEvents []*events.DataAuthorization
			eventsMock := &daemonevents.MockPublisherSubscriber[events.DataAuthorization]{
				Handler: func(event events.DataAuthorization) error {
					eventCopy := event
					publishedEvents = append(publishedEvents, &eventCopy)
					return nil
				},
			}

			r := &RPC{
				consentChecker:   &mock.AnalyticsConsentCheckerMock{ConsentCompleted: true},
				ac:               &testauth.AuthCheckerMock{LoggedIn: false},
				cm:               mock.NewMockConfigManager(),
				credentialsAPI:   &testcore.CredentialsAPIMock{},
				events:           &daemonevents.Events{User: &daemonevents.LoginEvents{Login: eventsMock}},
				authentication:   &testcore.AuthenticationAPImock{TokenValue: "token"},
				publisher:        &subs.Subject[string]{},
				ncClient:         &mock.NotificationClientMock{},
				initialLoginType: NewAtomicLoginType(),
			}

			// initialLoginType is UNKNOWN (default) — no LoginOAuth2 was called

			resp, err := r.LoginOAuth2Callback(context.Background(), &pb.LoginOAuth2CallbackRequest{
				Token: "exchange-token",
				Type:  tt.loginType,
			})

			assert.NoError(t, err)
			assert.Equal(t, pb.LoginStatus_SUCCESS, resp.Status)
			assert.GreaterOrEqual(t, len(publishedEvents), 1)

			for i, evt := range publishedEvents {
				assert.False(t, evt.IsAlteredFlowOnNordAccount,
					"Event %d: IsAlteredFlowOnNordAccount should be false "+
						"when no prior OAuth2 flow was started (initialLoginType=UNKNOWN), got true", i)
			}
		})
	}
}

// TestOAuth2Callback_AlteredFlowOnMissingExchangeToken tests the scenario where
// LoginOAuth2Callback receives an empty exchange token. The error event should
// have IsAlteredFlowOnNordAccount=false when the callback type matches the
// initially started flow type, and also false when no flow was started at all.
func TestOAuth2Callback_AlteredFlowOnMissingExchangeToken(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name                       string
		prevLoginType              pb.LoginType
		callbackType               pb.LoginType
		expectedAlteredFlowOnEvent bool
	}{
		{
			name:                       "no prior flow started, LOGIN callback with empty token",
			prevLoginType:              pb.LoginType_LoginType_UNKNOWN,
			callbackType:               pb.LoginType_LoginType_LOGIN,
			expectedAlteredFlowOnEvent: false,
		},
		{
			name:                       "LOGIN started, LOGIN callback with empty token",
			prevLoginType:              pb.LoginType_LoginType_LOGIN,
			callbackType:               pb.LoginType_LoginType_LOGIN,
			expectedAlteredFlowOnEvent: false,
		},
		{
			name:                       "SIGNUP started, LOGIN callback with empty token",
			prevLoginType:              pb.LoginType_LoginType_SIGNUP,
			callbackType:               pb.LoginType_LoginType_LOGIN,
			expectedAlteredFlowOnEvent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var publishedEvents []*events.DataAuthorization
			eventsMock := &daemonevents.MockPublisherSubscriber[events.DataAuthorization]{
				Handler: func(event events.DataAuthorization) error {
					eventCopy := event
					publishedEvents = append(publishedEvents, &eventCopy)
					return nil
				},
			}

			r := &RPC{
				consentChecker:   &mock.AnalyticsConsentCheckerMock{ConsentCompleted: true},
				ac:               &testauth.AuthCheckerMock{LoggedIn: false},
				events:           &daemonevents.Events{User: &daemonevents.LoginEvents{Login: eventsMock}},
				publisher:        &subs.Subject[string]{},
				initialLoginType: NewAtomicLoginType(),
			}

			if tt.prevLoginType != pb.LoginType_LoginType_UNKNOWN {
				r.initialLoginType.set(tt.prevLoginType)
			}

			_, err := r.LoginOAuth2Callback(context.Background(), &pb.LoginOAuth2CallbackRequest{
				Token: "", // empty token triggers ReasonLoginExchangeTokenMissing
				Type:  tt.callbackType,
			})
			assert.Error(t, err)

			assert.GreaterOrEqual(t, len(publishedEvents), 1)

			// Find the event with ReasonLoginExchangeTokenMissing
			var failureEvent *events.DataAuthorization
			for _, evt := range publishedEvents {
				if evt.Reason == events.ReasonLoginExchangeTokenMissing {
					failureEvent = evt
					break
				}
			}
			assert.NotNil(t, failureEvent, "Expected event with ReasonLoginExchangeTokenMissing")
			assert.Equal(t, events.StatusFailure, failureEvent.EventStatus)
			assert.Equal(t, tt.expectedAlteredFlowOnEvent, failureEvent.IsAlteredFlowOnNordAccount,
				"IsAlteredFlowOnNordAccount mismatch for prevLoginType=%v, callbackType=%v",
				tt.prevLoginType, tt.callbackType)
		})
	}
}

// TestOAuth2URLRetrieveFailed_StaleLoginType tests:
// 1. User starts OAuth2 LOGIN flow -> initialLoginType set to LOGIN
// 2. Callback never arrives (user abandons)
// 3. User starts new OAuth2 SIGNUP flow, but URL retrieval fails
// 4. initialLoginType remains LOGIN (stale from step 1) because set() is
//    only called after successful URL retrieval
// 5. User retries OAuth2 SIGNUP, URL retrieval succeeds
// 6. LoginOAuth2Callback completes as SIGNUP and correctly reports
//    IsAlteredFlowOnNordAccount=true (LOGIN was started, SIGNUP completed)
func TestOAuth2URLRetrieveFailed_StaleLoginType(t *testing.T) {
	category.Set(t, category.Unit)

	var publishedEvents []*events.DataAuthorization
	eventsMock := &daemonevents.MockPublisherSubscriber[events.DataAuthorization]{
		Handler: func(event events.DataAuthorization) error {
			eventCopy := event
			publishedEvents = append(publishedEvents, &eventCopy)
			return nil
		},
	}

	r := &RPC{
		consentChecker:   &mock.AnalyticsConsentCheckerMock{ConsentCompleted: true},
		ac:               &testauth.AuthCheckerMock{LoggedIn: false},
		cm:               mock.NewMockConfigManager(),
		credentialsAPI:   &testcore.CredentialsAPIMock{},
		events:           &daemonevents.Events{User: &daemonevents.LoginEvents{Login: eventsMock}},
		authentication:   &testcore.AuthenticationAPImock{}, // succeeds
		publisher:        &subs.Subject[string]{},
		ncClient:         &mock.NotificationClientMock{},
		initialLoginType: NewAtomicLoginType(),
	}

	// Step 1: User starts OAuth2 LOGIN flow successfully
	resp, err := r.LoginOAuth2(context.Background(), &pb.LoginOAuth2Request{
		Type: pb.LoginType_LoginType_LOGIN,
	})
	assert.NoError(t, err)
	assert.Equal(t, pb.LoginStatus_SUCCESS, resp.Status)
	assert.Equal(t, pb.LoginType_LoginType_LOGIN, r.initialLoginType.get())

	// Step 2: Callback never arrives, user abandons the flow.
	// initialLoginType remains LOGIN.

	// Step 3: User starts new OAuth2 SIGNUP, but URL retrieval fails.
	publishedEvents = nil
	r.authentication = &testcore.AuthenticationAPImock{
		LoginError: errors.New("network is unreachable"),
	}

	resp, err = r.LoginOAuth2(context.Background(), &pb.LoginOAuth2Request{
		Type: pb.LoginType_LoginType_SIGNUP,
	})
	assert.NoError(t, err)
	assert.Equal(t, pb.LoginStatus_NO_NET, resp.Status)

	// initialLoginType remains LOGIN (stale) because set() was not reached.
	assert.Equal(t, pb.LoginType_LoginType_LOGIN, r.initialLoginType.get(),
		"initialLoginType should still be LOGIN because URL retrieval failed before set() was called")

	// The ReasonUnfinishedPrevLogin event always reports IsAlteredFlowOnNordAccount=false
	// because LoginOAuth2 does not track flow alteration at URL-retrieval time.
	var unfinishedEvent *events.DataAuthorization
	for _, evt := range publishedEvents {
		if evt.Reason == events.ReasonUnfinishedPrevLogin {
			unfinishedEvent = evt
			break
		}
	}
	assert.NotNil(t, unfinishedEvent, "Expected an unfinished previous login event")
	assert.False(t, unfinishedEvent.IsAlteredFlowOnNordAccount,
		"ReasonUnfinishedPrevLogin event always reports IsAlteredFlowOnNordAccount=false")

	// Step 4: User retries SIGNUP, URL retrieval succeeds.
	publishedEvents = nil
	r.authentication = &testcore.AuthenticationAPImock{} // succeeds

	resp, err = r.LoginOAuth2(context.Background(), &pb.LoginOAuth2Request{
		Type: pb.LoginType_LoginType_SIGNUP,
	})
	assert.NoError(t, err)
	assert.Equal(t, pb.LoginStatus_SUCCESS, resp.Status)
	assert.Equal(t, pb.LoginType_LoginType_SIGNUP, r.initialLoginType.get())

	// Step 5: SIGNUP callback arrives. initialLoginType is SIGNUP, callback type is SIGNUP:
	// isAltered returns false — the final completed flow matches what was set.
	publishedEvents = nil
	cbResp, err := r.LoginOAuth2Callback(context.Background(), &pb.LoginOAuth2CallbackRequest{
		Token: "exchange-token",
		Type:  pb.LoginType_LoginType_SIGNUP,
	})
	assert.NoError(t, err)
	assert.Equal(t, pb.LoginStatus_SUCCESS, cbResp.Status)

	assert.GreaterOrEqual(t, len(publishedEvents), 1)
	var callbackEvent *events.DataAuthorization
	for _, evt := range publishedEvents {
		if evt.EventStatus == events.StatusSuccess {
			callbackEvent = evt
			break
		}
	}
	assert.NotNil(t, callbackEvent, "Expected a success callback event")
	assert.False(t, callbackEvent.IsAlteredFlowOnNordAccount,
		"Callback event should not report altered flow when SIGNUP started and SIGNUP completed")

	// Verify initialLoginType is reset after successful callback
	assert.Equal(t, pb.LoginType_LoginType_UNKNOWN, r.initialLoginType.get())
}
