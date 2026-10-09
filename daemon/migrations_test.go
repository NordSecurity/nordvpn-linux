package daemon

import (
	"os"
	"path/filepath"
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

// p2pGroup is the deprecated P2P group ID.
const p2pGroup config.ServerGroup = 15

func TestMigrateDeprecatedRegionalAutoconnect_PreservesCountryAndClearsGroup(t *testing.T) {
	category.Set(t, category.Unit)

	cm := mock.NewMockConfigManager()
	cm.Cfg.AutoConnect = true
	cm.Cfg.AutoConnectData.ServerTag = "germany"
	cm.Cfg.AutoConnectData.Country = "de"
	cm.Cfg.AutoConnectData.City = ""
	cm.Cfg.AutoConnectData.Group = regionalGroupEurope

	assert.NilError(t, MigrateDeprecatedGroupsAutoconnect(cm))

	assert.Equal(t, cm.SaveCallCount, 1)
	assert.Equal(t, cm.Cfg.AutoConnectData.Group, config.ServerGroup_UNDEFINED)
	assert.Equal(t, cm.Cfg.AutoConnectData.ServerTag, "germany")
	assert.Equal(t, cm.Cfg.AutoConnectData.Country, "de")
	assert.Equal(t, cm.Cfg.AutoConnectData.City, "")
}

func TestMigrateDeprecatedP2PAutoconnect_PreservesCountryAndClearsGroup(t *testing.T) {
	category.Set(t, category.Unit)

	cm := mock.NewMockConfigManager()
	cm.Cfg.AutoConnect = true
	cm.Cfg.AutoConnectData.ServerTag = "germany"
	cm.Cfg.AutoConnectData.Country = "de"
	cm.Cfg.AutoConnectData.City = ""
	cm.Cfg.AutoConnectData.Group = p2pGroup

	assert.NilError(t, MigrateDeprecatedGroupsAutoconnect(cm))

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

	assert.NilError(t, MigrateDeprecatedGroupsAutoconnect(cm))

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

	assert.NilError(t, MigrateDeprecatedGroupsAutoconnect(cm))

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

	assert.NilError(t, MigrateDeprecatedGroupsAutoconnect(cm))

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

	assert.NilError(t, MigrateDeprecatedGroupsAutoconnect(cm))
	assert.Equal(t, cm.SaveCallCount, 1)

	assert.NilError(t, MigrateDeprecatedGroupsAutoconnect(cm))
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

	testDoubleVPNServerTag := "ts-st123"
	testDoubleVPNCountryName := "Test Double VPN Country"
	testDoubleVPNCityName := "Test Double VPN City"

	testDoubleVPNServerLocation := core.Locations{
		core.Location{
			Country: core.Country{
				Name: testDoubleVPNCountryName,
				City: core.City{
					Name: testDoubleVPNCityName,
				},
			},
		},
	}

	testDoubleVPNServer := core.Server{
		Hostname:  testDoubleVPNServerTag + ".nordvpn.com",
		Locations: testDoubleVPNServerLocation,
	}

	testServersList := core.Servers{testServer, testDoubleVPNServer}

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
			expectedTag:         "test_city",
			expectedCountryName: testCountryName,
			expectedCityName:    testCityName,
		},
		{
			name:                "server set to specific double VPN, fallback to country",
			currentServerTag:    testDoubleVPNServerTag,
			serversList:         testServersList,
			expectedTag:         "test_double_vpn_city",
			expectedCountryName: testDoubleVPNCountryName,
			expectedCityName:    testDoubleVPNCityName,
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
			name:             "server set to specific, server not found, fallback to country code",
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

func migrationTestCountries() core.Countries {
	return core.Countries{
		{ID: 1, Name: "Germany", Code: "DE", Cities: core.Cities{
			{ID: 11, Name: "Berlin"},
			{ID: 12, Name: "Frankfurt"},
		}},
		{ID: 2, Name: "United States", Code: "US", Cities: core.Cities{
			{ID: 21, Name: "New York"},
			{ID: 22, Name: "Los Angeles"},
		}},
		{ID: 3, Name: "United Kingdom", Code: "GB", Cities: core.Cities{
			{ID: 31, Name: "London"},
		}},
	}
}

func newServerTagDataManager(t *testing.T, countries core.Countries, servers core.Servers) *DataManager {
	t.Helper()
	dm := &DataManager{}
	dm.serversData = ServersData{Servers: servers}
	if countries != nil {
		path := filepath.Join(t.TempDir(), "countries.dat")
		assert.NilError(t, os.WriteFile(path, []byte{}, internal.PermUserRW))
		dm.countryData = CountryData{filePath: path, Countries: countries}
	} else {
		dm.countryData = CountryData{filePath: filepath.Join(t.TempDir(), "missing.dat")}
	}
	return dm
}

func TestMigrateAutoConnectDataServerTag(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name        string
		countries   core.Countries
		inTag       string
		inCountry   string
		inCity      string
		wantTag     string
		wantCountry string
		wantCity    string
	}{
		{
			name:      "empty tag is a no-op",
			countries: migrationTestCountries(),
			inTag:     "",
			wantTag:   "",
		},
		{
			name:        "empty tag preserves existing country and city",
			countries:   migrationTestCountries(),
			inTag:       "",
			inCountry:   "spain",
			inCity:      "madrid",
			wantTag:     "",
			wantCountry: "spain",
			wantCity:    "madrid",
		},
		{
			name:        "country name sets country and leaves city untouched",
			countries:   migrationTestCountries(),
			inTag:       "germany",
			inCity:      "preexisting_city",
			wantTag:     "",
			wantCountry: "germany",
			wantCity:    "preexisting_city",
		},
		{
			name:        "country code sets country and leaves city untouched",
			countries:   migrationTestCountries(),
			inTag:       "de",
			inCity:      "preexisting_city",
			wantTag:     "",
			wantCountry: "germany",
			wantCity:    "preexisting_city",
		},
		{
			name:        "city overwrites country and sets city",
			countries:   migrationTestCountries(),
			inTag:       "berlin",
			inCountry:   "preexisting_country",
			wantTag:     "",
			wantCountry: "germany",
			wantCity:    "berlin",
		},
		{
			name:        "multi-word city is snake cased",
			countries:   migrationTestCountries(),
			inTag:       "los_angeles",
			wantTag:     "",
			wantCountry: "united_states",
			wantCity:    "los_angeles",
		},
		{
			name:        "matching is case insensitive",
			countries:   migrationTestCountries(),
			inTag:       "GERMANY",
			wantTag:     "",
			wantCountry: "germany",
		},
		{
			name:        "unknown tag is discarded and country/city preserved",
			countries:   migrationTestCountries(),
			inTag:       "atlantis",
			inCountry:   "keep_country",
			inCity:      "keep_city",
			wantTag:     "",
			wantCountry: "keep_country",
			wantCity:    "keep_city",
		},
		{
			name:        "leftover server-style tag is discarded",
			countries:   migrationTestCountries(),
			inTag:       "xx999",
			inCountry:   "keep_country",
			inCity:      "keep_city",
			wantTag:     "",
			wantCountry: "keep_country",
			wantCity:    "keep_city",
		},
		{
			name:        "no country data clears tag and preserves country/city",
			countries:   nil,
			inTag:       "germany",
			inCountry:   "keep_country",
			inCity:      "keep_city",
			wantTag:     "",
			wantCountry: "keep_country",
			wantCity:    "keep_city",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dataManager := newServerTagDataManager(t, test.countries, nil)

			configManager := newMockConfigManager()
			configManager.c = config.Config{
				AutoConnectData: config.AutoConnectData{
					ServerTag: test.inTag,
					Country:   test.inCountry,
					City:      test.inCity,
				},
			}

			assert.NilError(t, MigrateAutoConnectDataServerTag(configManager, dataManager))

			got := configManager.c.AutoConnectData
			assert.Equal(t, test.wantTag, got.ServerTag, "unexpected ServerTag")
			assert.Equal(t, test.wantCountry, got.Country, "unexpected Country")
			assert.Equal(t, test.wantCity, got.City, "unexpected City")
		})
	}
}

func TestMigrateAutoConnectDataServerTag_TwoPhase(t *testing.T) {
	category.Set(t, category.Unit)

	countries := migrationTestCountries()
	servers := core.Servers{
		{
			Hostname: "de123.nordvpn.com",
			Locations: core.Locations{
				core.Location{Country: core.Country{
					Name: "Germany",
					Code: "DE",
					City: core.City{Name: "Berlin"},
				}},
			},
		},
	}

	tests := []struct {
		name  string
		inTag string
		// state expected after phase 1 (specific-server migration)
		// phase 1 is MigrateDeprecatedAutoconnectToSpecificServer
		afterPhase1Tag     string
		afterPhase1Country string
		afterPhase1City    string
		// final state after phase 2 (ServerTag migration)
		wantTag     string
		wantCountry string
		wantCity    string
	}{
		{
			// Specific server: phase 1 resolves the server to its city tag and
			// raw country/city; phase 2 then normalizes and clears the tag
			name:               "specific server resolved by phase 1, finalized by phase 2",
			inTag:              "de123",
			afterPhase1Tag:     "berlin",
			afterPhase1Country: "Germany",
			afterPhase1City:    "Berlin",
			wantTag:            "",
			wantCountry:        "germany",
			wantCity:           "berlin",
		},
		{
			// City target: phase 1 is a no-op (not a server tag), phase 2 does
			// the whole migration
			name:           "city target untouched by phase 1",
			inTag:          "berlin",
			afterPhase1Tag: "berlin",
			wantTag:        "",
			wantCountry:    "germany",
			wantCity:       "berlin",
		},
		{
			// Country target: phase 1 is a no-op, phase 2 sets the country
			name:           "country target untouched by phase 1",
			inTag:          "germany",
			afterPhase1Tag: "germany",
			wantTag:        "",
			wantCountry:    "germany",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dataManager := newServerTagDataManager(t, countries, servers)

			configManager := newMockConfigManager()
			configManager.c = config.Config{
				AutoConnectData: config.AutoConnectData{ServerTag: test.inTag},
			}

			assert.NilError(t, MigrateDeprecatedAutoconnectToSpecificServer(configManager, dataManager))

			afterPhase1 := configManager.c.AutoConnectData
			assert.Equal(t, test.afterPhase1Tag, afterPhase1.ServerTag, "unexpected ServerTag after phase 1")
			assert.Equal(t, test.afterPhase1Country, afterPhase1.Country, "unexpected Country after phase 1")
			assert.Equal(t, test.afterPhase1City, afterPhase1.City, "unexpected City after phase 1")

			assert.NilError(t, MigrateAutoConnectDataServerTag(configManager, dataManager))

			got := configManager.c.AutoConnectData
			assert.Equal(t, test.wantTag, got.ServerTag, "unexpected ServerTag after phase 2")
			assert.Equal(t, test.wantCountry, got.Country, "unexpected Country after phase 2")
			assert.Equal(t, test.wantCity, got.City, "unexpected City after phase 2")
		})
	}
}
