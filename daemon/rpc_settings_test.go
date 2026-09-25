package daemon

import (
	"context"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"gotest.tools/v3/assert"
)

func TestSettings_NoPeerContext(t *testing.T) {
	category.Set(t, category.Unit)

	r := testRPC()
	resp, err := r.Settings(context.Background(), &pb.Empty{})

	assert.NilError(t, err)
	assert.Equal(t, resp.Type, internal.CodeFailure)
}
