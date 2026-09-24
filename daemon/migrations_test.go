package daemon

import (
	"os"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
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

func TestMigrateDeprecatedAutoconnectToSpecificServer(t *testing.T) {
	category.Set(t, category.Unit)

	testServerTag := "ts123"
	testCountryName := "Test Country"
	testCityName := "Test City"

	testServerLocation := core.Locations{
		core.Location{
			Country: core.Country{
				Name: testCountryName,
				City: core.City{
					Name: testCityName,
				},
			},
		},
	}

	testServer := core.Server{
		Hostname:  testServerTag + ".nordvpn.com",
		Locations: testServerLocation,
	}

	testServersList := core.Servers{testServer}

	tests := []struct {
		name                string
		serversList         core.Servers
		currentServerTag    string
		expectedTag         string
		expectedCountryName string
		expectedCityName    string
	}{
		{
			name:                "server set to specific, fallback to country",
			currentServerTag:    testServerTag,
			serversList:         testServersList,
			expectedTag:         strings.ToLower(testCityName),
			expectedCountryName: testCountryName,
			expectedCityName:    testCityName,
		},
		{
			name:             "server set to country, no fallback",
			currentServerTag: strings.ToLower(testCountryName),
			serversList:      testServersList,
			expectedTag:      strings.ToLower(testCountryName),
		},
		{
			name:             "server set to city, no fallback",
			currentServerTag: strings.ToLower(testCityName),
			serversList:      testServersList,
			expectedTag:      strings.ToLower(testCityName),
		},
		{
			name:             "server tag empty, no fallback",
			currentServerTag: "",
			serversList:      testServersList,
		},
		{
			name:             "server set to specific, server not found, fallback to fastest",
			currentServerTag: testServerTag,
			serversList:      core.Servers{},
			expectedTag:      "ts",
		},
		{
			name:             "server set to specific, server tag invalid format, fallback to fastest",
			currentServerTag: "tttt111",
			serversList:      core.Servers{},
			expectedTag:      "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dataManager := DataManager{}
			dataManager.serversData = ServersData{
				Servers: test.serversList,
			}

			cfg := config.Config{
				AutoConnectData: config.AutoConnectData{
					ServerTag: test.currentServerTag,
				},
			}

			configManager := newMockConfigManager()
			configManager.c = cfg

			MigrateDeprecatedAutoconnectToSpecificServer(configManager, &dataManager)

			cfgAfterFallback := configManager.c

			assert.Equal(t, test.expectedTag, cfgAfterFallback.AutoConnectData.ServerTag,
				"Invalid server tag saved in local config after fallback.")
			assert.Equal(t, test.expectedCountryName, cfgAfterFallback.AutoConnectData.Country,
				"Invalid country saved in local config after fallback.")
			assert.Equal(t, test.expectedCityName, cfgAfterFallback.AutoConnectData.City,
				"Invalid city saved in local config after fallback.")
		})
	}
}
