package libtelio

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/netip"
	"sync"
	"testing"
	"time"

	teliogo "github.com/NordSecurity/libtelio-go/v5"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/stretchr/testify/assert"
)

const (
	exampleAppVersion                  = "3.16.3"
	exampleDeviceID                    = "11111"
	exampleEventPath                   = "/var/data.db"
	vpnPeersPersistentKeepaliveSeconds = uint32(25)
	directPersistentKeepaliveSeconds   = uint32(5)
	proxyingPersistentKeepaliveSeconds = uint32(25)
	stunPersistentKeepaliveSeconds     = uint32(25)
)

type mockLib struct{}

func (mockLib) ConnectToExitNode(teliogo.PublicKey, *[]teliogo.IpNet, *teliogo.SocketAddr) error {
	return nil
}
func (mockLib) ConnectToExitNodePostquantum(
	*string, teliogo.PublicKey, *[]teliogo.IpNet, teliogo.SocketAddr) error {
	return nil
}
func (mockLib) DisconnectFromExitNodes() error                                       { return nil }
func (mockLib) SetMeshnetOff() error                                                 { return nil }
func (mockLib) NotifyNetworkChange(string) error                                     { return nil }
func (mockLib) SetMeshnet(teliogo.Config) error                                      { return nil }
func (mockLib) GetStatusMap() []teliogo.TelioNode                                    { return nil }
func (mockLib) StartNamed(teliogo.SecretKey, teliogo.TelioAdapterType, string) error { return nil }
func (mockLib) Stop() error                                                          { return nil }
func (mockLib) SetFwmark(uint32) error                                               { return nil }
func (mockLib) SetSecretKey(teliogo.SecretKey) error                                 { return nil }

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
			ch := make(chan state, 1)
			var connectionEstablishedWG sync.WaitGroup
			connectionEstablishedWG.Add(1)
			go func() {
				ch <- test.state
				connectionEstablishedWG.Done()
			}()

			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
			defer cancel()
			isConnectedC := isConnected(ctx, ch, connParameters{pubKey: test.publicKey}, vpn.NewInternalVPNEvents())

			connectionEstablishedWG.Wait()
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

	actualCfg, err := handleTelioConfig(exampleEventPath, exampleAppVersion, true, &remoteConfigGetter)

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

	cfg, err := handleTelioConfig(exampleEventPath, exampleAppVersion, true, &remoteConfigGetter)

	assert.NoError(t, err)

	assert.Nil(t, cfg.Lana)
	assert.Nil(t, cfg.Nurse)
}

func Test_TelioConfigAllDisabled(t *testing.T) {
	category.Set(t, category.Integration)

	remoteConfigGetter := mockVersionGetter{telioRemoteTestConfigAllDisabled}

	cfg, err := handleTelioConfig(exampleEventPath, exampleAppVersion, true, &remoteConfigGetter)

	assert.NoError(t, err)

	assert.Nil(t, cfg.Lana)
	assert.Nil(t, cfg.Nurse)
	assert.Nil(t, cfg.Derp)
	assert.Nil(t, cfg.Direct)

	// defaults from libtelio
	assert.NotNil(t, cfg.Wireguard)
	assert.NotNil(t, cfg.Wireguard.PersistentKeepalive)

	assert.NotNil(t, cfg.Wireguard.PersistentKeepalive.Vpn)
	assert.Equal(t, vpnPeersPersistentKeepaliveSeconds, *cfg.Wireguard.PersistentKeepalive.Vpn)

	assert.Equal(t, directPersistentKeepaliveSeconds, cfg.Wireguard.PersistentKeepalive.Direct)

	assert.NotNil(t, cfg.Wireguard.PersistentKeepalive.Proxying)
	assert.Equal(t, proxyingPersistentKeepaliveSeconds, *cfg.Wireguard.PersistentKeepalive.Proxying)

	assert.NotNil(t, cfg.Wireguard.PersistentKeepalive.Stun)
	assert.Equal(t, stunPersistentKeepaliveSeconds, *cfg.Wireguard.PersistentKeepalive.Stun)
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
	"Type": "node",
	"Body": {
		"Identifier": "1dd9e096-f420-4afa-bb19-62286a370dc9",
		"PublicKey": "m1ZvUX5fF5KJA8wQTFukhyxzHDfVQkzKXdi7L7PeVCe=",
		"State": "connected",
		"IsExit": false,
		"IsVpn": false,
		"IpAddresses": [
			"248.146.217.126"
		],
		"AllowedIps": [
			"248.146.217.126/32"
		],
		"Endpoint": "65.97.11.97:53434",
		"Hostname": "host-andes.nord",
		"AllowIncomingConnections": true,
		"AllowPeerSendFiles": true,
		"Path": "direct"
	}
}`

	expectedMaskedEventText := `{
	"Type": "node",
	"Body": {
		"Identifier": "1dd9e096-f420-4afa-bb19-62286a370dc9",
		"PublicKey": "***",
		"State": "connected",
		"IsExit": false,
		"IsVpn": false,
		"IpAddresses": [
			"248.146.217.126"
		],
		"AllowedIps": [
			"248.146.217.126/32"
		],
		"Endpoint": "65.97.11.97:53434",
		"Hostname": "host-andes.nord",
		"AllowIncomingConnections": true,
		"AllowPeerSendFiles": true,
		"Path": "direct"
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

type subscriber struct {
	mu      sync.RWMutex
	counter int
}

func (s *subscriber) NotifyConnect(events.DataConnect) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counter++
	return nil
}
func (s *subscriber) NotifyDisconnect(events.DataDisconnect) error {
	return s.NotifyConnect(events.DataConnect{})
}

func (s *subscriber) Counter() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.counter
}

func TestLibtelio_connect(t *testing.T) {
	category.Set(t, category.Unit)
	for _, tt := range []struct {
		name   string
		err    error
		active bool
		events int
		body   func(context.CancelFunc, chan<- state, *sync.WaitGroup, *sync.WaitGroup)
	}{
		{
			name:   "ctx done before connection established",
			err:    context.Canceled,
			active: false,
			events: 0,
			body: func(cf context.CancelFunc, _ chan<- state, _ *sync.WaitGroup, eventsSentWG *sync.WaitGroup) {
				cf()
				eventsSentWG.Done()
			},
		},
		{
			name:   "successful connection",
			err:    nil,
			active: true,
			events: 1,
			body: func(cf context.CancelFunc, events chan<- state, _ *sync.WaitGroup, eventsSentWG *sync.WaitGroup) {
				events <- state{
					State:  teliogo.NodeStateConnected,
					IsExit: true,
				}
				eventsSentWG.Done()
			},
		},
		{
			name:   "ctx done after connection established has no impact",
			err:    nil,
			active: true,
			events: 4,
			body: func(cf context.CancelFunc, events chan<- state, connectionEstablishedWG *sync.WaitGroup, eventsSentWG *sync.WaitGroup) {
				events <- state{
					State:  teliogo.NodeStateConnected,
					IsExit: true,
				}
				// make sure context is canceled after function exited
				connectionEstablishedWG.Wait()
				cf()

				// Check that events are still received
				events <- state{
					State:  teliogo.NodeStateDisconnected,
					IsExit: true,
				}
				events <- state{
					State:  teliogo.NodeStateConnected,
					IsExit: true,
				}
				events <- state{
					State:  teliogo.NodeStateDisconnected,
					IsExit: true,
				}
				// Send one more state which will be ignored just to make sure the
				// earlier states arrived
				events <- state{
					IsExit: false,
				}

				eventsSentWG.Done()
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			events := make(chan state, 1)
			pub := vpn.NewInternalVPNEvents()
			sub := &subscriber{}
			pub.Subscribe(sub)
			// Create instance
			libtelio := &Libtelio{
				lib:             mockLib{},
				events:          events,
				state:           vpn.ExitedState,
				fwmark:          123,
				eventsPublisher: pub,
			}

			// connect ctx
			ctx, cancel := context.WithCancel(context.Background())
			connectionInitiatedWG := sync.WaitGroup{}
			connectionEstablishedWG := sync.WaitGroup{}
			connectionInitiatedWG.Add(1)
			connectionEstablishedWG.Add(1)
			var err error
			go func() {
				connectionInitiatedWG.Done()
				err = libtelio.connect(ctx, netip.Addr{}, "", false)
				connectionEstablishedWG.Done()
			}()

			// Wait until goroutine starts
			connectionInitiatedWG.Wait()

			eventsSentWG := sync.WaitGroup{}
			eventsSentWG.Add(1)
			tt.body(cancel, events, &connectionEstablishedWG, &eventsSentWG)

			// Wait until goroutine stops
			connectionEstablishedWG.Wait()
			assert.ErrorIs(t, tt.err, err)
			assert.Equal(t, tt.active, libtelio.active)

			eventsSentWG.Wait()
			assert.Equal(t, tt.events, sub.Counter())
		})
	}
}
