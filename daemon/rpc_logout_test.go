package daemon

import (
	"context"
	"net/http"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/response"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/fileshare/service"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/nc"
	"github.com/stretchr/testify/assert"
)

type mockNC struct {
	nc.NotificationClient
}

func (mockNC) Stop() error { return nil }

func TestLogout_Token(t *testing.T) {
	rpc := RPC{
		ac:        &workingLoginChecker{},
		cm:        newMockConfigManager(),
		fileshare: service.MockFileshare{},
		netw:      workingNetworker{},
		ncClient:  mockNC{},
		publisher: &subs.Subject[string]{},
		api:       core.NewDefaultAPI("", "", http.DefaultClient, response.MockValidator{}),
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
					tokenData.RenewToken = "1234"
				} else {
					tokenData.RenewToken = ""
				}
				c.TokensData[c.AutoConnectData.ID] = tokenData
				return c
			})
			assert.NoError(t, err)
			rpc.Logout(context.Background(), &pb.LogoutRequest{PersistToken: test.persistToken})
		})
	}

}
