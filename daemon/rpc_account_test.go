package daemon

import (
	"context"
	"errors"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/events"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	testauth "github.com/NordSecurity/nordvpn-linux/test/mock/auth"
	testcore "github.com/NordSecurity/nordvpn-linux/test/mock/core"
	"github.com/stretchr/testify/assert"
)

var user1 core.CurrentUserResponse = core.CurrentUserResponse{
	Username: "username1",
	Email:    "user1@mail.com",
}

var user2 core.CurrentUserResponse = core.CurrentUserResponse{
	Username: "username2",
	Email:    "user2@mail.com",
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
	authCheckerMock := testauth.AuthCheckerMock{LoggedIn: true}
	configManagerMock := mock.NewMockConfigManager()
	credentialsAPIMock := testcore.CredentialsAPIMock{}

	r := RPC{
		dm:             dataManager,
		ac:             &authCheckerMock,
		cm:             configManagerMock,
		credentialsAPI: &credentialsAPIMock,
		events:         events.NewEventsEmpty(),
	}

	tests := []struct {
		name              string
		apiResponse       core.CurrentUserResponse
		cachedResponse    *pb.AccountResponse
		expectedResponse  *pb.AccountResponse
		cacheExpired      bool
		respectDataExpiry bool //former "Full" flag
	}{
		{
			name:              "Request only valid cache data. Cache is valid.",
			apiResponse:       user2,
			cachedResponse:    userResponseToAccountResponse(user1),
			expectedResponse:  userResponseToAccountResponse(user1),
			cacheExpired:      false,
			respectDataExpiry: true,
		},
		{
			name:              "Request only valid cache data. Cache is expired.",
			apiResponse:       user2,
			cachedResponse:    userResponseToAccountResponse(user1),
			expectedResponse:  userResponseToAccountResponse(user2),
			cacheExpired:      true,
			respectDataExpiry: true,
		},
		{
			name:              "Request any validity cache data (a.k.a give me what you got). Cache is valid.",
			apiResponse:       user2,
			cachedResponse:    userResponseToAccountResponse(user1),
			expectedResponse:  userResponseToAccountResponse(user1),
			cacheExpired:      false,
			respectDataExpiry: false,
		},
		{
			name:              "Request any validity cache data (a.k.a give me what you got). Cache is expired.",
			apiResponse:       user2,
			cachedResponse:    userResponseToAccountResponse(user1),
			expectedResponse:  userResponseToAccountResponse(user1),
			cacheExpired:      true,
			respectDataExpiry: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dataManager.SetAccountData(test.cachedResponse)
			if test.cacheExpired {
				dataManager.accountData.unset()
			}
			credentialsAPIMock.CurrentUserResponse = test.apiResponse
			resp, err := r.AccountInfo(context.Background(), &pb.AccountRequest{Full: test.respectDataExpiry})
			assert.NoError(t, err)
			assert.Equal(t, test.expectedResponse.String(), resp.String())
		})
	}
}

func TestAccountInfo_FullRequestUpdatesCache(t *testing.T) {
	category.Set(t, category.Unit)

	dataManager := NewDataManager("", "", "", "", events.NewDataUpdateEvents())

	cachedResponse := userResponseToAccountResponse(user1)
	dataManager.SetAccountData(cachedResponse)

	authCheckerMock := testauth.AuthCheckerMock{LoggedIn: true}
	configManagerMock := mock.NewMockConfigManager()

	credentialsAPIMock := testcore.CredentialsAPIMock{}
	credentialsAPIMock.CurrentUserResponse = user2

	r := RPC{
		dm:             dataManager,
		ac:             &authCheckerMock,
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
	authCheckerMock := testauth.AuthCheckerMock{LoggedIn: true}
	configManagerMock := mock.NewMockConfigManager()
	credentialsAPIMock := testcore.CredentialsAPIMock{}

	r := RPC{
		dm:             dataManager,
		ac:             &authCheckerMock,
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
			authCheckerMock.IsVPNExpiredErr = test.isVPNExpiredErr
			authCheckerMock.GetDedicatedIPServicesErr = test.getDedicatedIPServicesErr
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
