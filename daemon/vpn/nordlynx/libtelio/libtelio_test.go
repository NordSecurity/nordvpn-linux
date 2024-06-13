package libtelio

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	teliogo "github.com/NordSecurity/libtelio-go/v5"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

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
		name          string
		state         state
		publicKey     string
		channelClosed bool // channel will be closed when connection is established
	}{
		{
			name: "connecting",
			state: state{
				State:     teliogo.NodeStateConnecting,
				PublicKey: "123",
				IsExit:    true,
			},
			publicKey: "123",
		},
		{
			name: "connected",
			state: state{
				State:     teliogo.NodeStateConnected,
				PublicKey: "123",
				IsExit:    true,
			},
			publicKey:     "123",
			channelClosed: true,
		},
		{
			name: "different pubkey",
			state: state{
				State:     teliogo.NodeStateConnected,
				PublicKey: "321",
				IsExit:    true,
			},
			publicKey: "123",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ch := make(chan state)
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				ch <- test.state
				wg.Done()
			}()

			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
			defer cancel()
			isConnectedC := isConnected(ctx, ch, connParameters{pubKey: test.publicKey}, vpn.NewInternalVPNEvents())

			wg.Wait()
			select {
			case _, ok := <-isConnectedC:
				assert.Equal(t, test.channelClosed, !ok)
			case <-time.After(1 * time.Second):
				assert.False(t, test.channelClosed, "Channel was not closed when state changed to connected")
			}
		})
	}
}

func TestEventCallback_DoesntBlock(t *testing.T) {
	stateC := make(chan state)
	cb := eventCallback(stateC)
	var event teliogo.Event

	returnedC := make(chan any)
	go func() {
		cb(event)
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

	telioCfg := &teliogo.Features{}
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

	expectedCfg := toTelioFeatures(t, telioRemoteTestConfig)

	remoteConfigGetter := mockVersionGetter{telioRemoteTestConfig}

	actualCfg, err := handleTelioConfig(exampleEventPath, exampleDeviceID, exampleAppVersion, true, &remoteConfigGetter)

	assert.NoError(t, err)

	if diff := cmp.Diff(actualCfg, &expectedCfg, cmpopts.SortSlices(func(a, b teliogo.EndpointProvider) bool {
		return a < b
	})); diff != "" {
		t.Errorf("Telio Config mismatch (-want +got):\n%s", diff)
	}
}

func toTelioFeatures(t *testing.T, cfg string) teliogo.Features {
	features, err := teliogo.DeserializeFeatureConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return features
}

func Test_TelioConfigLanaDisabled(t *testing.T) {
	category.Set(t, category.Integration)

	remoteConfigGetter := mockVersionGetter{telioRemoteTestConfigLanaDisabled}

	cfg, err := handleTelioConfig(exampleEventPath, exampleDeviceID, exampleAppVersion, true, &remoteConfigGetter)

	assert.NoError(t, err)

	assert.Nil(t, cfg.Lana)
	assert.Nil(t, cfg.Nurse)
}

func Test_TelioConfigAllDisabled(t *testing.T) {
	category.Set(t, category.Integration)

	remoteConfigGetter := mockVersionGetter{telioRemoteTestConfigAllDisabled}

	cfg, err := handleTelioConfig(exampleEventPath, exampleDeviceID, exampleAppVersion, true, &remoteConfigGetter)

	assert.NoError(t, err)

	assert.Nil(t, cfg.Lana)
	assert.Nil(t, cfg.Nurse)
	assert.Nil(t, cfg.Derp)
	assert.Nil(t, cfg.Direct)

	// defaults from libtelio
	assert.NotNil(t, cfg.Wireguard)
	assert.NotNil(t, cfg.Wireguard.PersistentKeepalive)

	assert.NotNil(t, cfg.Wireguard.PersistentKeepalive.Vpn)
	assert.Equal(t, uint32(25), *cfg.Wireguard.PersistentKeepalive.Vpn)

	assert.Equal(t, uint32(5), cfg.Wireguard.PersistentKeepalive.Direct)

	assert.NotNil(t, cfg.Wireguard.PersistentKeepalive.Proxying)
	assert.Equal(t, uint32(25), *cfg.Wireguard.PersistentKeepalive.Proxying)

	assert.NotNil(t, cfg.Wireguard.PersistentKeepalive.Stun)
	assert.Equal(t, uint32(25), *cfg.Wireguard.PersistentKeepalive.Stun)
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
