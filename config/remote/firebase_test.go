package remote

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

const firebaseTokenEnvKey = "FIREBASE_TOKEN"

func TestRemoteConfig_GetValue(t *testing.T) {
	category.Set(t, category.Integration)
	rc := NewRConfig(time.Duration(0), os.Getenv(firebaseTokenEnvKey))
	welcomeMessage, err := rc.GetValue("welcome_message")
	assert.NoError(t, err)
	assert.Equal(t, "hola", welcomeMessage)
}

func TestRemoteConfig_Caching(t *testing.T) {
	category.Set(t, category.Integration)
	rc := NewRConfig(time.Hour*24, os.Getenv(firebaseTokenEnvKey))
	_, err := rc.GetValue("welcome_message")
	assert.NoError(t, err)
	rc.config = nil // imitate incorrectly received config

	_, err = rc.GetValue("welcome_message")
	assert.Error(t, err)

	rc.lastUpdate = time.Now().Add(-time.Hour * 48)

	welcomeMessage, err := rc.GetValue("welcome_message")
	assert.NoError(t, err)
	assert.Equal(t, "hola", welcomeMessage)
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

	remoteConfigGetter := NewRConfig(time.Duration(0), os.Getenv(firebaseTokenEnvKey))
	remoteTelioCfg, err := remoteConfigGetter.GetTelioConfig(Version)

	assert.NoError(t, err)

	var j1, j2 interface{}
	err1 := json.Unmarshal([]byte(expectedCfg), &j1)
	err2 := json.Unmarshal([]byte(remoteTelioCfg), &j2)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
}
