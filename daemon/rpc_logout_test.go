package daemon

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	daemonevents "github.com/NordSecurity/nordvpn-linux/daemon/events"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/nc"
	testcore "github.com/NordSecurity/nordvpn-linux/test/mock/core"
	"github.com/NordSecurity/nordvpn-linux/test/mock/networker"
	testnorduser "github.com/NordSecurity/nordvpn-linux/test/mock/norduser/service"
)

type mockNC struct {
	nc.NotificationClient
}

func (mockNC) Stop() error      { return nil }
func (mockNC) Revoke(bool) bool { return true }

type mockApi struct {
	core.CombinedAPI
}

func (mockApi) DeleteToken(token string) error { return nil }
func (mockApi) Logout(token string) error      { return nil }

func TestLogout_Token(t *testing.T) {
	rpc := RPC{
		ac:             &workingLoginChecker{},
		cm:             newMockConfigManager(),
		norduser:       &testnorduser.MockNorduserCombinedService{},
		netw:           &networker.Mock{},
		ncClient:       mockNC{},
		publisher:      &subs.Subject[string]{},
		credentialsAPI: &testcore.CredentialsAPIMock{},
		events:         &daemonevents.Events{User: &daemonevents.LoginEvents{Logout: &daemonevents.MockPublisherSubscriber[events.DataAuthorization]{}}},
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
				return c
			})
			assert.NoError(t, err)
			resp, err := rpc.Logout(context.Background(), &pb.LogoutRequest{PersistToken: test.persistToken})
			assert.NoError(t, err)
			assert.Equal(t, test.result, resp.Type)
		})
	}
}
