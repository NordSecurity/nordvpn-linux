package libtelio

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestIsConnected(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name      string
		state     state
		publicKey string
		expected  bool
	}{
		{
			name: "connecting",
			state: state{
				State:     "connecting",
				PublicKey: "123",
			},
			publicKey: "123",
		},
		{
			name: "connected",
			state: state{
				State:     "connected",
				PublicKey: "123",
			},
			publicKey: "123",
			expected:  true,
		},
		{
			name: "misbehaving",
			state: state{
				State:     "misbehaving",
				PublicKey: "123",
			},
			publicKey: "123",
		},
		{
			name: "different pubkey",
			state: state{
				State:     "connected",
				PublicKey: "321",
			},
			publicKey: "123",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ch := make(chan state)
			go func() { ch <- test.state }()

			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
			defer cancel()
			isConnectedC := isConnected(ctx, ch, test.publicKey)

			assert.Equal(t, test.expected, <-isConnectedC)
		})
	}
}

func TestEventCallback_DoesntBlock(t *testing.T) {
	stateC := make(chan state)
	cb := eventCallback(stateC)
	event, err := json.Marshal(state{})
	assert.NoError(t, err)

	returnedC := make(chan any)
	go func() {
		cb(string(event))
		returnedC <- nil
	}()

	condition := func() bool {
		select {
		case <-returnedC:
			return true
		default:
			return false
		}
	}
	assert.Eventually(t, condition, time.Millisecond*100, time.Millisecond*10)
}

func Test_TelioDefaultConfig(t *testing.T) {
	category.Set(t, category.Integration)

	expectedCfg := `
	{
		"lana": {
			"event_path": "",
			"prod": false
		},
		"nurse": {
			"fingerprint": "",
			"qos": {}
		},
		"direct": {},
		"derp": {},
		"wireguard": {
			"persistent_keepalive": {}
		}
	}
	`

	telioCfg := newTelioFeatures()
	jsn, err := json.Marshal(telioCfg)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(string(jsn))
	}
	var j1, j2 interface{}
	err1 := json.Unmarshal([]byte(expectedCfg), &j1)
	err2 := json.Unmarshal(jsn, &j2)

	assert.NoError(t, err1)
	assert.NoError(t, err2)

	assert.True(t, reflect.DeepEqual(j1, j2))
}

const telioRemoteTestConfig string = `
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

type mockVersionGetter struct{}

func (mockVersionGetter) GetValue(key string) (string, error) {
	return "", nil
}

func (mockVersionGetter) GetTelioConfig(string) (string, error) {
	return telioRemoteTestConfig, nil
}

func Test_TelioConfig(t *testing.T) {
	category.Set(t, category.Integration)

	expectedCfg := telioRemoteTestConfig

	deviceID := "11111"
	appVersion := "3.16.3"
	eventPath := "/var/data.db"
	prod := true

	remoteConfigGetter := mockVersionGetter{}

	cfg, err := handleTelioConfig(eventPath, deviceID, appVersion, prod, remoteConfigGetter)

	assert.NoError(t, err)

	var j1, j2 interface{}
	err1 := json.Unmarshal([]byte(expectedCfg), &j1)
	err2 := json.Unmarshal(cfg, &j2)

	assert.NoError(t, err1)
	assert.NoError(t, err2)

	assert.True(t, reflect.DeepEqual(j1, j2))
}
