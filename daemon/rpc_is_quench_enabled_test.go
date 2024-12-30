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

func TestIsNordWhisperEnabled(t *testing.T) {
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
			resp, err := r.IsNordWhisperEnabled(context.Background(), &pb.Empty{})
			assert.Nil(t, err, "Unexpected error returned by IsNordWhisperEnabled rpc.")
			assert.Equal(t, resp.Enabled, false, "Unexpected response type received.")
		})
	}
}
