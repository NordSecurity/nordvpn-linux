package daemon

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/events"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	testauth "github.com/NordSecurity/nordvpn-linux/test/mock/auth"
	testcore "github.com/NordSecurity/nordvpn-linux/test/mock/core"
	"gotest.tools/v3/assert"
)

var user1 core.CurrentUserResponse = core.CurrentUserResponse{
	Username: "username1",
	Email:    "user1@mail.com",
}

var user2 core.CurrentUserResponse = core.CurrentUserResponse{
	Username: "username2",
	Email:    "user2mail.com",
}

func userResponseToAccountResponse(userResponse core.CurrentUserResponse) *pb.AccountResponse {
	return &pb.AccountResponse{
		Type:              internal.CodeSuccess,
		Username:          userResponse.Username,
		Email:             userResponse.Email,
		DedicatedIpStatus: internal.CodeNoService,
		MfaStatus:         pb.TriState_DISABLED,
	}
}

func TestAccountInfo(t *testing.T) {
	category.Set(t, category.Unit)

	dataManager := NewDataManager("", "", "", "", events.NewDataUpdateEvents())
	authCheeckerMock := testauth.AuthCheckerMock{LoggedIn: true}
	configManagerMock := mock.NewMockConfigManager()
	credentialsAPIMock := testcore.CredentialsAPIMock{}

	r := RPC{
		dm:             dataManager,
		ac:             &authCheeckerMock,
		cm:             configManagerMock,
		credentialsAPI: &credentialsAPIMock,
		events:         events.NewEventsEmpty(),
	}

	tests := []struct {
		name             string
		userResponse     core.CurrentUserResponse
		cachedResponse   *pb.AccountResponse
		expectedResponse *pb.AccountResponse
		cacheExpired     bool
		full             bool
	}{
		{
			name:             "full request",
			userResponse:     user2,
			cachedResponse:   userResponseToAccountResponse(user1),
			expectedResponse: userResponseToAccountResponse(user2),
			full:             true,
		},
		{
			name:             "limited request",
			userResponse:     user2,
			cachedResponse:   userResponseToAccountResponse(user1),
			expectedResponse: userResponseToAccountResponse(user1),
			cacheExpired:     false,
			full:             false,
		},
		{
			name:             "limited request cache expired",
			userResponse:     user2,
			cachedResponse:   userResponseToAccountResponse(user1),
			expectedResponse: userResponseToAccountResponse(user2),
			cacheExpired:     true,
			full:             false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dataManager.accountData.checkCacheValidityFunc = func(_ time.Time) bool {
				return test.cacheExpired
			}
			dataManager.SetAccountData(test.cachedResponse)
			credentialsAPIMock.CurrentUserResponse = test.userResponse
			resp, err := r.AccountInfo(context.Background(), &pb.AccountRequest{Full: test.full})
			assert.NilError(t, err)
			assert.Equal(t, test.expectedResponse.String(), resp.String())

			cachedResponse := r.dm.accountData.accountData
			assert.Equal(t, test.expectedResponse.String(), cachedResponse.String())
		})
	}
}

func TestAccountInfo_FullRequestUpdatesCache(t *testing.T) {
	category.Set(t, category.Unit)

	dataManager := NewDataManager("", "", "", "", events.NewDataUpdateEvents())
	dataManager.accountData.checkCacheValidityFunc = func(_ time.Time) bool {
		return false
	}

	cachedResponse := userResponseToAccountResponse(user1)
	dataManager.SetAccountData(cachedResponse)

	authCheeckerMock := testauth.AuthCheckerMock{LoggedIn: true}
	configManagerMock := mock.NewMockConfigManager()

	credentialsAPIMock := testcore.CredentialsAPIMock{}
	credentialsAPIMock.CurrentUserResponse = user2

	r := RPC{
		dm:             dataManager,
		ac:             &authCheeckerMock,
		cm:             configManagerMock,
		credentialsAPI: &credentialsAPIMock,
		events:         events.NewEventsEmpty(),
	}

	resp, _ := r.AccountInfo(context.Background(), &pb.AccountRequest{Full: false})
	assert.Equal(t, cachedResponse.String(), resp.String())

	// update the cache
	updatedResponse, _ := r.AccountInfo(context.Background(), &pb.AccountRequest{Full: true})

	resp, _ = r.AccountInfo(context.Background(), &pb.AccountRequest{Full: false})
	assert.Equal(t, updatedResponse.String(), resp.String())
}

func TestAccountInfo_FailedRequestDoesntUpdateTheCache(t *testing.T) {
	category.Set(t, category.Unit)

	dataManager := NewDataManager("", "", "", "", events.NewDataUpdateEvents())
	dataManager.accountData.checkCacheValidityFunc = func(t time.Time) bool { return false }
	authCheeckerMock := testauth.AuthCheckerMock{LoggedIn: true}
	configManagerMock := mock.NewMockConfigManager()
	credentialsAPIMock := testcore.CredentialsAPIMock{}

	r := RPC{
		dm:             dataManager,
		ac:             &authCheeckerMock,
		cm:             configManagerMock,
		credentialsAPI: &credentialsAPIMock,
		events:         events.NewEventsEmpty(),
	}

	tests := []struct {
		name                      string
		isVPNExpiredErr           error
		getDedicatedIPServicesErr error
		loadConfigErr             error
		currentUserErr            error
	}{
		{
			name:            "get vpn expired fail",
			isVPNExpiredErr: errors.New("get vpn expired error"),
		},
		{
			name:                      "get DIP services fail",
			getDedicatedIPServicesErr: errors.New("get DIP services error"),
		},
		{
			name:          "load config fail",
			loadConfigErr: errors.New("load config error"),
		},
		{
			name:           "get current user fail",
			currentUserErr: errors.New("get current user error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			authCheeckerMock.IsVPNExpiredErr = test.isVPNExpiredErr
			authCheeckerMock.GetDedicatedIPServicesErr = test.getDedicatedIPServicesErr

			configManagerMock.LoadErr = test.loadConfigErr

			credentialsAPIMock.CurrentUserErr = test.currentUserErr

			credentialsAPIMock.CurrentUserResponse = user2
			dataManager.SetAccountData(userResponseToAccountResponse(user1))

			r.AccountInfo(context.Background(), &pb.AccountRequest{Full: true})

			resp, _ := r.AccountInfo(context.Background(), &pb.AccountRequest{Full: false})
			assert.Equal(t, userResponseToAccountResponse(user1).String(), resp.String(),
				"Invalid data returned from the RPC(should be the cached data)")
		})
	}
}
