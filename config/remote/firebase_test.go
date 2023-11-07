package remote

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"

	"github.com/stretchr/testify/assert"
)

var (
	configLocation = "testdata/settings_with_rc_3.16.5.dat"
	vaultLocation  = "testdata/install_with_rc_3.16.5.dat"
)

func initConfig() {
	cm := getConfigManager()
	if err := cm.SaveWith(func(c config.Config) config.Config {
		c.RCLastUpdate = time.Now().Add(-30 * time.Hour)
		return c
	}); err != nil {
		fmt.Println(err)
	}
}
func getConfigManager() config.Manager {
	salt, _ := os.LookupEnv("SALT")
	return config.NewFilesystemConfigManager(configLocation, vaultLocation, salt, config.LinuxMachineIDGetter{}, config.StdFilesystemHandle{})
}

func TestRemoteConfig_GetValue(t *testing.T) {
	category.Set(t, category.File)
	initConfig()
	rc := NewRConfig(time.Duration(0), &remoteServiceMock{}, getConfigManager())
	welcomeMessage, err := rc.GetValue("welcome_message")
	assert.NoError(t, err)
	assert.Equal(t, "hola", welcomeMessage)
}

func TestRemoteConfig_Caching(t *testing.T) {
	category.Set(t, category.File)
	initConfig()
	rsm := remoteServiceMock{}
	rc := NewRConfig(time.Hour*24, &rsm, getConfigManager())
	_, err := rc.GetValue("welcome_message")
	assert.NoError(t, err)
	assert.Equal(t, 1, rsm.fetchCount)
	welcomeMessage, err := rc.GetValue("welcome_message")
	assert.NoError(t, err)
	assert.Equal(t, "hola", welcomeMessage)
	assert.Equal(t, 1, rsm.fetchCount)
}

func TestRemoteConfig_NoCaching(t *testing.T) {
	category.Set(t, category.File)
	initConfig()
	rsm := remoteServiceMock{}
	rc := NewRConfig(0, &rsm, getConfigManager())
	_, err := rc.GetValue("welcome_message")
	assert.NoError(t, err)
	assert.Equal(t, 1, rsm.fetchCount)
	welcomeMessage, err := rc.GetValue("welcome_message")
	assert.NoError(t, err)
	assert.Equal(t, "hola", welcomeMessage)
	assert.Equal(t, 2, rsm.fetchCount)
}

func TestRemoteConfig_stringToSemVersion(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name          string
		versionString string
		versionPrefix string
		expectError   bool
	}{
		{
			name:          "valid version 1",
			versionString: "libtelio-1.1.1",
			versionPrefix: "libtelio-",
			expectError:   false,
		},
		{
			name:          "valid version 2",
			versionString: "v1.1.1",
			versionPrefix: "",
			expectError:   false,
		},
		{
			name:          "valid version 3",
			versionString: "1.1.1+aefeaf",
			versionPrefix: "",
			expectError:   false,
		},
		{
			name:          "invalid version",
			versionString: "1.1",
			versionPrefix: "",
			expectError:   true,
		},
		{
			name:          "invalid version - empty string",
			versionString: "",
			versionPrefix: "",
			expectError:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := stringToSemVersion(test.versionString, test.versionPrefix)
			assert.Equal(t, test.expectError, err != nil, test.name)
		})
	}
}

func TestRemoteConfig_insertFieldVersion(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name               string
		stringToInsert     string
		versionPrefix      string
		initialStringArray []string //expecting descending sort order
		expectedPosition   int
	}{
		{
			name:               "case1",
			stringToInsert:     "1.1.1",
			versionPrefix:      "",
			initialStringArray: []string{},
			expectedPosition:   0,
		},
		{
			name:               "case2",
			stringToInsert:     "1.1.1",
			versionPrefix:      "",
			initialStringArray: []string{"1.2.1"},
			expectedPosition:   1,
		},
		{
			name:               "case3",
			stringToInsert:     "1.10.1",
			versionPrefix:      "",
			initialStringArray: []string{"1.2.1"},
			expectedPosition:   0,
		},
		{
			name:               "case4",
			stringToInsert:     "1.10.1",
			versionPrefix:      "",
			initialStringArray: []string{"3.15.1", "1.29.1", "0.1.0"},
			expectedPosition:   2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			initialVerArr := []*fieldVersion{}
			for _, s := range test.initialStringArray {
				ver, err := stringToSemVersion(s, test.versionPrefix)
				assert.NoError(t, err, test.name)
				initialVerArr = append(initialVerArr, &fieldVersion{ver, s})
			}
			ver, err := stringToSemVersion(test.stringToInsert, test.versionPrefix)

			assert.NoError(t, err, test.name)

			result := insertFieldVersion(initialVerArr, &fieldVersion{ver, test.stringToInsert})

			pos := 0
			for i, v := range result {
				if v.version.Compare(*ver) == 0 {
					pos = i
				}
			}
			assert.Equal(t, test.expectedPosition, pos, test.name)
		})
	}
}

func TestRemoteConfig_findVersionField(t *testing.T) {
	category.Set(t, category.Integration)

	tests := []struct {
		name               string
		initialStringArray []string
		versionPrefix      string
		appVersion         string
		expectedFieldName  string
		expectError        bool
	}{
		{
			name:               "case1",
			initialStringArray: []string{"telio_3_15_1", "telio_1_29_1", "telio_0_1_0"},
			versionPrefix:      "telio_",
			appVersion:         "v1.1.1",
			expectedFieldName:  "telio_0_1_0",
			expectError:        false,
		},
		{
			name:               "case2",
			initialStringArray: []string{"telio_3_15_1", "telio_1_29_1", "telio_0_1_0"},
			versionPrefix:      "telio_",
			appVersion:         "1.29.1+abcefaa",
			expectedFieldName:  "telio_1_29_1",
			expectError:        false,
		},
		{
			name:               "case3",
			initialStringArray: []string{"telio_3_15_1", "telio_1_29_1", "telio_0_5_0"},
			versionPrefix:      "telio_",
			appVersion:         "v0.1.0",
			expectedFieldName:  "",
			expectError:        true,
		},
		{
			name:               "case4",
			initialStringArray: []string{"telio_3_15_1", "any-other", "telio_1_29_1", "telio_0_5_0"},
			versionPrefix:      "telio_",
			appVersion:         "v10.1.0",
			expectedFieldName:  "telio_3_15_1",
			expectError:        false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			initialVerArr := []*fieldVersion{}
			for _, s := range test.initialStringArray {
				ver, err := stringToSemVersion(s, test.versionPrefix)
				if err != nil {
					continue
				}
				assert.NoError(t, err, test.name)
				initialVerArr = append(initialVerArr, &fieldVersion{ver, s})
			}
			ver, err := stringToSemVersion(test.appVersion, test.versionPrefix)

			assert.NoError(t, err, test.name)

			fieldName, err := findVersionField(initialVerArr, ver)
			assert.Equal(t, test.expectError, err != nil, test.name)
			assert.Equal(t, test.expectedFieldName, fieldName, test.name)
		})
	}
}

func TestRemoteConfig_GetTelioConfig(t *testing.T) {
	category.Set(t, category.Integration)

	expectedCfg := `
	{
		"lana": {
			"event_path": "/var/data.db",
			"prod": true
		},
		"nurse": {
			"fingerprint": "11111",
			"heartbeat_interval": 3600,
			"qos": {
				"rtt_interval": 300,
				"rtt_tries": 3,
				"rtt_types": [
					"Ping"
				],
				"buckets": 5
			}
		},
		"direct": {
			"endpoint_interval_secs": 20,
			"providers": [
				"local",
				"stun"
			]
		},
		"derp": {
			"tcp_keepalive": 15,
			"derp_keepalive": 60
		},
		"wireguard": {
			"persistent_keepalive": {
				"proxying": 25,
				"direct": 5,
				"vpn": 25,
				"stun": 50
			}
		},
		"exit-dns": "1.1.1.1"
	}
	`

	Version := "3.16.2"

	remoteConfigGetter := NewRConfig(time.Duration(0), &remoteServiceMock{}, getConfigManager())
	remoteTelioCfg, err := remoteConfigGetter.GetTelioConfig(Version)

	assert.NoError(t, err)

	var j1, j2 interface{}
	err1 := json.Unmarshal([]byte(expectedCfg), &j1)
	err2 := json.Unmarshal([]byte(remoteTelioCfg), &j2)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
}

func TestRemoteConfig_GetCachedData(t *testing.T) {
	category.Set(t, category.Unit)

	cm := &mock.ConfigManager{}
	rs := &remoteServiceMock{}

	tests := []struct {
		name          string
		expectedValue string
		fetchError    error
		updatePeriod  time.Duration
		remoteConfig  string
		cachedValue   string
	}{
		{
			name:       "fetch fails and nothing is cached",
			fetchError: fmt.Errorf("failed to fetch"),
		},
		{
			name:          "using cached data, no fetching is needed",
			fetchError:    fmt.Errorf("failed to fetch"),
			expectedValue: "hola",
			cachedValue:   remoteConfigString,
			updatePeriod:  time.Hour,
		},
		{
			name:          "use cache data when fetching fails",
			fetchError:    fmt.Errorf("failed to fetch"),
			expectedValue: "hola",
			cachedValue:   remoteConfigString,
		},
		{
			name:          "using fetch remote config",
			expectedValue: "hola",
			remoteConfig:  remoteConfigString,
			cachedValue:   "broke_data",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rc := NewRConfig(time.Duration(test.updatePeriod), rs, cm)

			rs.fetchError = test.fetchError
			rs.response = test.remoteConfig

			cm.Cfg = &config.Config{}
			cm.Cfg.RCLastUpdate = time.Now()
			cm.Cfg.RemoteConfig = test.cachedValue

			value, err := rc.GetValue("welcome_message")

			assert.Equal(t, test.expectedValue, value)
			if test.expectedValue == "" {
				assert.ErrorIs(t, err, test.fetchError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ----------------------------------------------------------------------------------------
type remoteServiceMock struct {
	fetchCount int
	response   string
	fetchError error
}

func (rs *remoteServiceMock) FetchRemoteConfig() ([]byte, error) {
	rs.fetchCount++
	if rs.fetchError != nil {
		return nil, rs.fetchError
	}

	if rs.response == "" {
		return []byte(remoteConfigString), nil
	}

	return []byte(rs.response), nil
}

var remoteConfigString = `
{
	"parameters": {
			"file_sharing_min_version": {
					"defaultValue": {
							"value": "3.15.5"
					}
			},
			"fileshare_min_version": {
					"defaultValue": {
							"value": "3.15.5"
					},
					"description": "Apps with lower version than min_version will have Fileshare feature disabled."
			},
			"mesh_enabled": {
					"defaultValue": {
							"value": "false"
					},
					"description": "Flag for remote meshnet enabling/disabling"
			},
			"meshnet_min_version": {
					"defaultValue": {
							"value": "3.12.0"
					},
					"description": "Minimal version for meshnet"
			},
			"min_version": {
					"defaultValue": {
							"value": "3.8.0-2"
					},
					"description": "Apps with lower version than min_version will not function and will require an upgrade."
			},
			"nat_traversal_enabled": {
					"defaultValue": {
							"value": "true"
					}
			},
			"nat_traversal_min_version": {
					"defaultValue": {
							"value": "3.15.1"
					},
					"description": "Apps with lower version will have NAT Traversal disabled"
			},
			"telio_analytics_enabled": {
					"defaultValue": {
							"value": "true"
					}
			},
			"telio_analytics_min_version": {
					"defaultValue": {
							"value": "3.15.2"
					}
			},
			"telio_config_3_16_2": {
					"defaultValue": {
							"value": "{    \"lana\": {},    \"nurse\": {        \"heartbeat_interval\": 600,        \"qos\": {            \"rtt_interval\": 300,            \"rtt_tries\": 3,            \"rtt_types\": [                \"Ping\"            ],            \"buckets\": 5        }    },    \"direct\": {        \"endpoint_interval_secs\": 20,        \"providers\": [            \"local\",            \"stun\"        ]    },    \"derp\": {        \"tcp_keepalive\": 15,        \"derp_keepalive\": 60    },    \"wireguard\": {        \"persistent_keepalive\": {            \"proxying\": 25,            \"direct\": 5,            \"vpn\": 25,            \"stun\": 50        }    }}"
					}
			},
			"telio_config_3_16_3": {
					"defaultValue": {
							"value": "{    \"lana\": {},    \"nurse\": {        \"heartbeat_interval\": 1200,        \"qos\": {            \"rtt_interval\": 300,            \"rtt_tries\": 3,            \"rtt_types\": [                \"Ping\"            ],            \"buckets\": 5        }    },    \"direct\": {        \"endpoint_interval_secs\": 20,        \"providers\": [            \"local\",            \"stun\"        ]    },    \"derp\": {        \"tcp_keepalive\": 15,        \"derp_keepalive\": 60    },    \"wireguard\": {        \"persistent_keepalive\": {            \"proxying\": 25,            \"direct\": 5,            \"vpn\": 25,            \"stun\": 50        }    }}"
					}
			},
			"test_min_version": {
					"defaultValue": {
							"value": "3.8.0-2"
					},
					"description": "Version for testing remote config"
			},
			"welcome_message": {
					"defaultValue": {
							"value": "hola"
					}
			},
			"welcome_message_caps": {
					"defaultValue": {
							"value": "false"
					}
			}
	}
}
`
