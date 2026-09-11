package daemon

import (
	"os"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"github.com/NordSecurity/nordvpn-linux/test/mock/fs"
	"gotest.tools/v3/assert"
)

// regionalGroupEurope is the deprecated EUROPE group ID.
const regionalGroupEurope config.ServerGroup = 19

func TestMigrateDeprecatedRegionalAutoconnect_PreservesCountryAndClearsGroup(t *testing.T) {
	category.Set(t, category.Unit)

	cm := mock.NewMockConfigManager()
	cm.Cfg.AutoConnect = true
	cm.Cfg.AutoConnectData.ServerTag = "germany"
	cm.Cfg.AutoConnectData.Country = "de"
	cm.Cfg.AutoConnectData.City = ""
	cm.Cfg.AutoConnectData.Group = regionalGroupEurope

	assert.NilError(t, MigrateDeprecatedRegionalAutoconnect(cm))

	assert.Equal(t, cm.SaveCallCount, 1)
	assert.Equal(t, cm.Cfg.AutoConnectData.Group, config.ServerGroup_UNDEFINED)
	assert.Equal(t, cm.Cfg.AutoConnectData.ServerTag, "germany")
	assert.Equal(t, cm.Cfg.AutoConnectData.Country, "de")
	assert.Equal(t, cm.Cfg.AutoConnectData.City, "")
}

func TestMigrateDeprecatedRegionalAutoconnect_PreservesCityAndClearsGroup(t *testing.T) {
	category.Set(t, category.Unit)

	cm := mock.NewMockConfigManager()
	cm.Cfg.AutoConnect = true
	cm.Cfg.AutoConnectData.ServerTag = "berlin"
	cm.Cfg.AutoConnectData.Country = ""
	cm.Cfg.AutoConnectData.City = "berlin"
	cm.Cfg.AutoConnectData.Group = regionalGroupEurope

	assert.NilError(t, MigrateDeprecatedRegionalAutoconnect(cm))

	assert.Equal(t, cm.SaveCallCount, 1)
	assert.Equal(t, cm.Cfg.AutoConnectData.Group, config.ServerGroup_UNDEFINED)
	assert.Equal(t, cm.Cfg.AutoConnectData.ServerTag, "berlin")
	assert.Equal(t, cm.Cfg.AutoConnectData.Country, "")
	assert.Equal(t, cm.Cfg.AutoConnectData.City, "berlin")
}

func TestMigrateDeprecatedRegionalAutoconnect_OnlyRegionalFallsBackToQuickConnect(t *testing.T) {
	category.Set(t, category.Unit)

	cm := mock.NewMockConfigManager()
	cm.Cfg.AutoConnect = true
	cm.Cfg.AutoConnectData.ServerTag = "europe"
	cm.Cfg.AutoConnectData.Country = ""
	cm.Cfg.AutoConnectData.City = ""
	cm.Cfg.AutoConnectData.Group = regionalGroupEurope

	assert.NilError(t, MigrateDeprecatedRegionalAutoconnect(cm))

	assert.Equal(t, cm.SaveCallCount, 1)
	assert.Equal(t, cm.Cfg.AutoConnectData.Group, config.ServerGroup_UNDEFINED)
	assert.Equal(t, cm.Cfg.AutoConnectData.ServerTag, "")
}

func TestMigrateDeprecatedRegionalAutoconnect_NonRegionalGroup_NoSave(t *testing.T) {
	category.Set(t, category.Unit)

	cm := mock.NewMockConfigManager()
	cm.Cfg.AutoConnect = true
	cm.Cfg.AutoConnectData.ServerTag = "us"
	cm.Cfg.AutoConnectData.Country = "us"
	cm.Cfg.AutoConnectData.City = ""
	cm.Cfg.AutoConnectData.Group = config.ServerGroup_DOUBLE_VPN

	assert.NilError(t, MigrateDeprecatedRegionalAutoconnect(cm))

	assert.Equal(t, cm.SaveCallCount, 0)
	assert.Equal(t, cm.Cfg.AutoConnectData.Group, config.ServerGroup_DOUBLE_VPN)
	assert.Equal(t, cm.Cfg.AutoConnectData.ServerTag, "us")
	assert.Equal(t, cm.Cfg.AutoConnectData.Country, "us")
}

func TestMigrateDeprecatedRegionalAutoconnect_Idempotent(t *testing.T) {
	category.Set(t, category.Unit)

	cm := mock.NewMockConfigManager()
	cm.Cfg.AutoConnect = true
	cm.Cfg.AutoConnectData.ServerTag = "europe"
	cm.Cfg.AutoConnectData.Group = regionalGroupEurope

	assert.NilError(t, MigrateDeprecatedRegionalAutoconnect(cm))
	assert.Equal(t, cm.SaveCallCount, 1)

	assert.NilError(t, MigrateDeprecatedRegionalAutoconnect(cm))
	assert.Equal(t, cm.SaveCallCount, 1)
}

func TestMigrateLegacyAllowlist_MovesWhitelistIntoAllowlist(t *testing.T) {
	category.Set(t, category.Unit)

	legacy := config.NewAllowlist([]int64{51820}, []int64{22}, []string{"192.168.1.0/24"})
	cfg := config.Config{}
	cfg.AutoConnectData.LegacyAllowlist = &legacy

	cfg = MigrateLegacyAllowlist(cfg)

	assert.DeepEqual(t, cfg.AutoConnectData.Allowlist, legacy)
	assert.Assert(t, cfg.AutoConnectData.LegacyAllowlist == nil)
}

func TestMigrateLegacyAllowlist_NoLegacyKey_Unchanged(t *testing.T) {
	category.Set(t, category.Unit)

	current := config.NewAllowlist(nil, []int64{443}, nil)
	cfg := config.Config{}
	cfg.AutoConnectData.Allowlist = current

	cfg = MigrateLegacyAllowlist(cfg)

	assert.DeepEqual(t, cfg.AutoConnectData.Allowlist, current)
	assert.Assert(t, cfg.AutoConnectData.LegacyAllowlist == nil)
}

func TestMigrateLegacyAllowlist_RewritesConfigFile(t *testing.T) {
	category.Set(t, category.Unit)

	location := "/location"
	filesystem := fs.NewSystemFileHandleMock(t)
	cm := config.NewFilesystemConfigManager(
		location, "/vault", "",
		config.NewMachineID(os.ReadFile, os.Hostname),
		&filesystem,
		nil)
	// versions pre-6.0.0 stored the allowlist under the "whitelist" key
	oldFile := `.DAT{"auto_connect_data":{"whitelist":{"ports":{"tcp":[22,443],"udp":[51820]},"subnets":["192.168.1.0/24"]}}}`
	filesystem.WriteFile(location, []byte(oldFile), internal.PermUserRW)

	// migration on daemon start
	assert.NilError(t, cm.SaveWith(MigrateLegacyAllowlist))

	var cfg config.Config
	assert.NilError(t, cm.Load(&cfg))
	expected := config.NewAllowlist([]int64{51820}, []int64{22, 443}, []string{"192.168.1.0/24"})
	assert.DeepEqual(t, cfg.AutoConnectData.Allowlist, expected)

	saved, _ := filesystem.ReadFile(location)
	assert.Assert(t, strings.Contains(string(saved), `"allowlist":`))
	assert.Assert(t, !strings.Contains(string(saved), `"whitelist"`))
}
