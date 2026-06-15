package daemon

import (
	"context"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendUIEvent(t *testing.T) {
	category.Set(t, category.Unit)

	r := &RPC{}
	resp, err := r.SendUIEvent(context.Background(), &pb.Empty{})
	require.NoError(t, err)
	assert.Equal(t, internal.CodeSuccess, resp.Type)
}
