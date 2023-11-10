package daemon

import (
	"context"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/fileshare/daemon"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/nc"
	"github.com/NordSecurity/nordvpn-linux/test/mock/networker"
	"github.com/stretchr/testify/assert"
)

type mockNC struct {
	nc.NotificationClient
}

func (mockNC) Stop() error { return nil }

type mockApi struct {
	core.CombinedAPI
}

func (mockApi) DeleteToken(token string) error { return nil }
func (mockApi) Logout(token string) error      { return nil }

func TestLogout_Token(t *testing.T) {
	rpc := RPC{
		ac:        &workingLoginChecker{},
		cm:        newMockConfigManager(),
		fileshare: daemon.NoopFileshare{},
		netw:      &networker.Mock{},
		ncClient:  mockNC{},
		publisher: &subs.Subject[string]{},
		api:       mockApi{},
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
