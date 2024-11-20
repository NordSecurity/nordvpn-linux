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

func TestSetTechnology_Quench(t *testing.T) {
	category.Set(t, category.Unit)

	remoteConfigGetter := mock.NewRemoteConfigMock()
	r := RPC{
		remoteConfigGetter: remoteConfigGetter,
	}

	tests := []struct {
		name             string
		quenchEnabledErr error
	}{
		{
			name: "quench disabled",
		},
		{
			name:             "failed to get quench status",
			quenchEnabledErr: fmt.Errorf("failed to get quench status"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := r.SetTechnology(context.Background(),
				&pb.SetTechnologyRequest{Technology: config.Technology_QUENCH})
			assert.Nil(t, err, "Unexpected error returned by IsQuenchEnabled rpc.")
			assert.Equal(t, resp.Type, internal.CodeFeatureHidden, "Unexpected response type received.")
		})
	}
}
