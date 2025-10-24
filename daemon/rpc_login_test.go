package daemon

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	daemonevents "github.com/NordSecurity/nordvpn-linux/daemon/events"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
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
					consentChecker: &mock.AnalyticsConsentCheckerMock{ConsentCompleted: true},
					ac:             &testauth.AuthCheckerMock{LoggedIn: false},
					cm:             mock.NewMockConfigManager(),
					credentialsAPI: &testcore.CredentialsAPIMock{},
					events:         &daemonevents.Events{User: &daemonevents.LoginEvents{Login: em}},
					authentication: &testcore.AuthenticationAPImock{TokenValue: "token"},
					publisher:      &subs.Subject[string]{},
					ncClient:       &mock.NotificationClientMock{},
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
					consentChecker: &mock.AnalyticsConsentCheckerMock{ConsentCompleted: true},
					ac:             &testauth.AuthCheckerMock{LoggedIn: false},
					cm:             mock.NewMockConfigManager(),
					credentialsAPI: &testcore.CredentialsAPIMock{},
					events:         &daemonevents.Events{User: &daemonevents.LoginEvents{Login: em}},
					authentication: &testcore.AuthenticationAPImock{TokenValue: "token"},
					publisher:      &subs.Subject[string]{},
					ncClient:       &mock.NotificationClientMock{},
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
