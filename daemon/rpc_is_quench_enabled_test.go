package daemon

import (
	"context"
	"fmt"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"github.com/stretchr/testify/assert"
)

func TestIsQuenchEnabled(t *testing.T) {
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
			resp, err := r.IsQuenchEnabled(context.Background(), &pb.Empty{})
			assert.Nil(t, err, "Unexpected error returned by IsQuenchEnabled rpc.")
			assert.Equal(t, resp.Enabled, false, "Unexpected response type received.")
		})
	}
}
