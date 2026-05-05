package daemon

import (
	"context"
	"sync"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"gotest.tools/v3/assert"
)

// resetSettingsMigrationOnces resets the package-level sync
func resetSettingsMigrationOnces() {
	adjustAutoconnectCfgOnce = sync.Once{}
	migrateRegionalAutoconnectOnce = sync.Once{}
}

// regionalGroupEurope is the deprecated EUROPE group ID
const regionalGroupEurope config.ServerGroup = 19

func TestSettings_NoPeerContext(t *testing.T) {
	category.Set(t, category.Unit)

	r := testRPC()
	resp, err := r.Settings(context.Background(), &pb.Empty{})

	assert.NilError(t, err)
	assert.Equal(t, resp.Type, internal.CodeFailure)
}

func TestSettings_AutoconnectMigrationRunsOnlyOnce(t *testing.T) {
	category.Set(t, category.Unit)
	resetSettingsMigrationOnces()

	cm := mock.NewMockConfigManager()

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

	// migration was not executed again
	assert.Equal(t, cm.SaveCallCount, 1)
}

func TestSettings_RegionalMigration_PreservesCountryAndClearsGroup(t *testing.T) {
	category.Set(t, category.Unit)
	resetSettingsMigrationOnces()

	cm := mock.NewMockConfigManager()
	cm.Cfg.AutoConnect = true
	cm.Cfg.AutoConnectData.ServerTag = "germany"
	cm.Cfg.AutoConnectData.Country = "de"
	cm.Cfg.AutoConnectData.City = ""
	cm.Cfg.AutoConnectData.Group = regionalGroupEurope

	r := testRPC()
	r.cm = cm

	_, err := r.Settings(peerCtx(trayTestUID), &pb.Empty{})
	assert.NilError(t, err)

	assert.Equal(t, cm.SaveCallCount, 1)
	assert.Equal(t, cm.Cfg.AutoConnectData.Group, config.ServerGroup_UNDEFINED)
	assert.Equal(t, cm.Cfg.AutoConnectData.ServerTag, "germany")
	assert.Equal(t, cm.Cfg.AutoConnectData.Country, "de")
	assert.Equal(t, cm.Cfg.AutoConnectData.City, "")
}

func TestSettings_RegionalMigration_PreservesCityAndClearsGroup(t *testing.T) {
	category.Set(t, category.Unit)
	resetSettingsMigrationOnces()

	cm := mock.NewMockConfigManager()
	cm.Cfg.AutoConnect = true
	cm.Cfg.AutoConnectData.ServerTag = "berlin"
	cm.Cfg.AutoConnectData.Country = ""
	cm.Cfg.AutoConnectData.City = "berlin"
	cm.Cfg.AutoConnectData.Group = regionalGroupEurope

	r := testRPC()
	r.cm = cm

	_, err := r.Settings(peerCtx(trayTestUID), &pb.Empty{})
	assert.NilError(t, err)

	assert.Equal(t, cm.SaveCallCount, 1)
	assert.Equal(t, cm.Cfg.AutoConnectData.Group, config.ServerGroup_UNDEFINED)
	assert.Equal(t, cm.Cfg.AutoConnectData.ServerTag, "berlin")
	assert.Equal(t, cm.Cfg.AutoConnectData.Country, "")
	assert.Equal(t, cm.Cfg.AutoConnectData.City, "berlin")
}

func TestSettings_RegionalMigration_OnlyRegionalFallsBackToQuickConnect(t *testing.T) {
	category.Set(t, category.Unit)
	resetSettingsMigrationOnces()

	cm := mock.NewMockConfigManager()
	cm.Cfg.AutoConnect = true
	cm.Cfg.AutoConnectData.ServerTag = "europe"
	cm.Cfg.AutoConnectData.Country = ""
	cm.Cfg.AutoConnectData.City = ""
	cm.Cfg.AutoConnectData.Group = regionalGroupEurope

	r := testRPC()
	r.cm = cm

	_, err := r.Settings(peerCtx(trayTestUID), &pb.Empty{})
	assert.NilError(t, err)

	assert.Equal(t, cm.SaveCallCount, 1)
	assert.Equal(t, cm.Cfg.AutoConnectData.Group, config.ServerGroup_UNDEFINED)
	assert.Equal(t, cm.Cfg.AutoConnectData.ServerTag, "")
}

func TestSettings_RegionalMigration_NonRegionalGroup_NoSave(t *testing.T) {
	category.Set(t, category.Unit)
	resetSettingsMigrationOnces()

	cm := mock.NewMockConfigManager()
	cm.Cfg.AutoConnect = true
	cm.Cfg.AutoConnectData.ServerTag = "us"
	cm.Cfg.AutoConnectData.Country = "us"
	cm.Cfg.AutoConnectData.City = ""
	cm.Cfg.AutoConnectData.Group = config.ServerGroup_DOUBLE_VPN

	r := testRPC()
	r.cm = cm

	_, err := r.Settings(peerCtx(trayTestUID), &pb.Empty{})
	assert.NilError(t, err)

	assert.Equal(t, cm.SaveCallCount, 0)
	assert.Equal(t, cm.Cfg.AutoConnectData.Group, config.ServerGroup_DOUBLE_VPN)
	assert.Equal(t, cm.Cfg.AutoConnectData.ServerTag, "us")
	assert.Equal(t, cm.Cfg.AutoConnectData.Country, "us")
}
