package daemon

import (
	"context"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"gotest.tools/v3/assert"
)

func TestSettings_NoPeerContext(t *testing.T) {
	category.Set(t, category.Unit)

	r := testRPC()
	resp, err := r.Settings(context.Background(), &pb.Empty{})

	assert.NilError(t, err)
	assert.Equal(t, resp.Type, internal.CodeFailure)
}

func TestSettings_AutoconnectMigrationRunsOnlyOnce(t *testing.T) {
	category.Set(t, category.Unit)

	cm := mock.NewMockConfigManager()

	// conditions for triggering migration of autoconnect
	cm.Cfg.AutoConnect = true
	cm.Cfg.AutoConnectData.ServerTag = "not-empty"
	cm.Cfg.AutoConnectData.Country = ""
	cm.Cfg.AutoConnectData.City = ""
	cm.Cfg.AutoConnectData.Group = config.ServerGroup_UNDEFINED

	r := testRPC()
	r.cm = cm
	assert.Equal(t, cm.SaveCallCount, 0)

	ctx := peerCtx(trayTestUID)
	_, err := r.Settings(ctx, &pb.Empty{})
	assert.NilError(t, err, "first Settings() call returned error: %v", err)

	// migration was performed
	assert.Equal(t, cm.SaveCallCount, 1)

	_, err = r.Settings(ctx, &pb.Empty{})
	assert.NilError(t, err, "second Settings() call returned error: %v", err)

	// still 1 - migration was not executed again
	assert.Equal(t, cm.SaveCallCount, 1)
}
