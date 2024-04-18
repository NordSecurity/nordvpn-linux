package libtelio

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

const (
	exampleAppVersion = "3.16.3"
	exampleDeviceID   = "11111"
	exampleEventPath  = "/var/data.db"
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
			isConnectedC := isConnected(ctx, ch, connParameters{pubKey: test.publicKey}, &vpn.Events{})

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

	telioCfg := &telioFeatures{}
	jsn, err := json.Marshal(telioCfg)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(string(jsn))
	}
	var intf interface{}
	err = json.Unmarshal(jsn, &intf)

	assert.NoError(t, err)
}

func Test_TelioConfig(t *testing.T) {
	category.Set(t, category.Integration)

	expectedCfg := telioRemoteTestConfig

	remoteConfigGetter := mockVersionGetter{telioRemoteTestConfig}

	cfg, err := handleTelioConfig(exampleEventPath, exampleDeviceID, exampleAppVersion, true, &remoteConfigGetter)

	assert.NoError(t, err)

	var j1, j2 telioFeatures
	err1 := json.Unmarshal([]byte(expectedCfg), &j1)
	err2 := json.Unmarshal(cfg, &j2)

	assert.NoError(t, err1)
	assert.NoError(t, err2)

	assert.True(t, reflect.DeepEqual(j1, j2))
}

func Test_TelioConfigLanaDisabled(t *testing.T) {
	category.Set(t, category.Integration)

	expectedCfg := telioRemoteTestConfigLanaDisabled

	remoteConfigGetter := mockVersionGetter{telioRemoteTestConfigLanaDisabled}

	cfg, err := handleTelioConfig(exampleEventPath, exampleDeviceID, exampleAppVersion, true, &remoteConfigGetter)

	assert.NoError(t, err)

	var j1, j2 telioFeatures
	err1 := json.Unmarshal([]byte(expectedCfg), &j1)
	err2 := json.Unmarshal(cfg, &j2)

	assert.NoError(t, err1)
	assert.NoError(t, err2)

	assert.Nil(t, j2.Lana)
	assert.Nil(t, j2.Nurse)
}

func Test_TelioConfigAllDisabled(t *testing.T) {
	category.Set(t, category.Integration)

	expectedCfg := telioRemoteTestConfigAllDisabled

	remoteConfigGetter := mockVersionGetter{telioRemoteTestConfigAllDisabled}

	cfg, err := handleTelioConfig(exampleEventPath, exampleDeviceID, exampleAppVersion, true, &remoteConfigGetter)

	assert.NoError(t, err)

	var j1, j2 telioFeatures
	err1 := json.Unmarshal([]byte(expectedCfg), &j1)
	err2 := json.Unmarshal(cfg, &j2)

	assert.NoError(t, err1)
	assert.NoError(t, err2)

	assert.Nil(t, j2.Lana)
	assert.Nil(t, j2.Nurse)
	assert.Nil(t, j2.Derp)
	assert.Nil(t, j2.Direct)
	assert.Nil(t, j2.Wireguard)
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
const telioRemoteTestConfigLanaDisabled string = `
{
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

const telioRemoteTestConfigAllDisabled string = `
{
	"exit-dns": "1.1.1.1"
}
`

type mockVersionGetter struct {
	remoteConfig string
}

func (m *mockVersionGetter) GetConfig(string) (string, error) {
	return m.remoteConfig, nil
}

func Test_maskPublicKey(t *testing.T) {
	eventText := `{
	"type": "node",
	"body": {
		"identifier": "1dd9e096-f420-4afa-bb19-62286a370dc9",
		"public_key": "m1ZvUX5fF5KJA8wQTFukhyxzHDfVQkzKXdi7L7PeVCe=",
		"state": "connected",
		"is_exit": false,
		"is_vpn": false,
		"ip_addresses": [
			"248.146.217.126"
		],
		"allowed_ips": [
			"248.146.217.126/32"
		],
		"endpoint": "65.97.11.97:53434",
		"hostname": "host-andes.nord",
		"allow_incoming_connections": true,
		"allow_peer_send_files": true,
		"path": "direct"
	}
}`

	expectedMaskedEventText := `{
	"type": "node",
	"body": {
		"identifier": "1dd9e096-f420-4afa-bb19-62286a370dc9",
		"public_key": "***",
		"state": "connected",
		"is_exit": false,
		"is_vpn": false,
		"ip_addresses": [
			"248.146.217.126"
		],
		"allowed_ips": [
			"248.146.217.126/32"
		],
		"endpoint": "65.97.11.97:53434",
		"hostname": "host-andes.nord",
		"allow_incoming_connections": true,
		"allow_peer_send_files": true,
		"path": "direct"
	}
}`

	buf := &bytes.Buffer{}
	var err error
	err = json.Compact(buf, []byte(eventText))
	assert.NoError(t, err)
	maskedEventText := maskPublicKey(buf.String())

	buf = &bytes.Buffer{}
	err = json.Compact(buf, []byte(expectedMaskedEventText))
	assert.NoError(t, err)
	assert.Equal(t, buf.String(), maskedEventText)
}
