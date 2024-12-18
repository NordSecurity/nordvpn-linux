package daemon

import (
	"context"
	"fmt"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"github.com/stretchr/testify/assert"
)

func TestSetTechnology_NordWhisper(t *testing.T) {
	category.Set(t, category.Unit)

	remoteConfigGetter := mock.NewRemoteConfigMock()
	r := RPC{
		remoteConfigGetter: remoteConfigGetter,
	}

	tests := []struct {
		name                  string
		nordWhisperEnabledErr error
	}{
		{
			name: "NordWhisper disabled",
		},
		{
			name:                  "failed to get NordWhisper status",
			nordWhisperEnabledErr: fmt.Errorf("failed to get NordWhisper status"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := r.SetTechnology(context.Background(),
				&pb.SetTechnologyRequest{Technology: config.Technology_NORDWHISPER})
			assert.Nil(t, err, "Unexpected error returned by IsNordWhisperEnabled rpc.")
			assert.Equal(t, resp.Type, internal.CodeFeatureHidden, "Unexpected response type received.")
		})
	}
}
