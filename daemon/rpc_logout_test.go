package daemon

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/NordSecurity/nordvpn-linux/config"
	daemonevents "github.com/NordSecurity/nordvpn-linux/daemon/events"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/recents"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	mockconfig "github.com/NordSecurity/nordvpn-linux/test/mock/config"
	testcore "github.com/NordSecurity/nordvpn-linux/test/mock/core"

	"github.com/NordSecurity/nordvpn-linux/test/mock/networker"
	testnorduser "github.com/NordSecurity/nordvpn-linux/test/mock/norduser/service"
)

func TestLogout_Token(t *testing.T) {
	cfgManagerMock := newMockConfigManager()
	fs := mockconfig.NewFilesystemMock(t)

	rpc := RPC{
		ac:             &workingLoginChecker{},
		cm:             cfgManagerMock,
		norduser:       &testnorduser.MockNorduserCombinedService{},
		netw:           &networker.Mock{},
		ncClient:       &mock.NotificationClientMock{},
		publisher:      &subs.Subject[string]{},
		credentialsAPI: &testcore.CredentialsAPIMock{},
		events: &daemonevents.Events{
			User:    &daemonevents.LoginEvents{Logout: &daemonevents.MockPublisherSubscriber[events.DataAuthorization]{}},
			Service: &daemonevents.ServiceEvents{Disconnect: &daemonevents.MockPublisherSubscriber[events.DataDisconnect]{}},
		},
		recentVPNConnStore: recents.NewRecentConnectionsStore("/test/path", &fs),
	}

	tests := []struct {
		name              string
		persistToken      bool
		loggedInWithToken bool
		result            int64
	}{
		{
			persistToken:      true,
			loggedInWithToken: true,
			result:            internal.CodeSuccess,
		},
		{
			persistToken:      true,
			loggedInWithToken: false,
			result:            internal.CodeSuccess,
		},
		{
			persistToken:      false,
			loggedInWithToken: true,
			result:            internal.CodeTokenInvalidated,
		},
		{
			persistToken:      false,
			loggedInWithToken: false,
			result:            internal.CodeSuccess,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := rpc.cm.SaveWith(func(c config.Config) config.Config {
				tokenData := c.TokensData[c.AutoConnectData.ID]
				if test.loggedInWithToken {
					tokenData.RenewToken = ""
				} else {
					tokenData.RenewToken = "1234"
				}
				tokenData.Token = "1234"
				c.TokensData[c.AutoConnectData.ID] = tokenData
				c.MeshPrivateKey = "key"
				return c
			})
			assert.NoError(t, err)
			resp, err := rpc.Logout(context.Background(), &pb.LogoutRequest{PersistToken: test.persistToken})
			assert.NoError(t, err)
			assert.Equal(t, test.result, resp.Type)
			assert.Equal(t, "", cfgManagerMock.c.MeshPrivateKey, "Mesh private key not removed after logout.")
		})
	}
}
