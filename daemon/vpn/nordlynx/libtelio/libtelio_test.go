package libtelio

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/netip"
	"sync"
	"testing"
	"time"

	teliogo "github.com/NordSecurity/libtelio-go/v5"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/tunnel"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/stretchr/testify/assert"
)

const (
	exampleDeviceID                    = "11111"
	exampleEventPath                   = "/var/data.db"
	vpnPeersPersistentKeepaliveSeconds = uint32(25)
	directPersistentKeepaliveSeconds   = uint32(5)
	proxyingPersistentKeepaliveSeconds = uint32(25)
	stunPersistentKeepaliveSeconds     = uint32(25)
)

type callbackHandlerStub struct{}

func (callbackHandlerStub) handleEvent(teliogo.Event) *teliogo.TelioError { return nil }
func (callbackHandlerStub) setMonitoringContext(ctx context.Context)      {}

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

type mockTunnel struct{}

func (mockTunnel) TransferRates() (tunnel.Statistics, error) { return tunnel.Statistics{}, nil }
func (mockTunnel) Interface() net.Interface                  { return net.Interface{Name: "nordlynx"} }
func (mockTunnel) IP() (netip.Addr, bool)                    { return netip.Addr{}, true }
func (mockTunnel) AddAddrs() error                           { return nil }
func (mockTunnel) DelAddrs() error                           { return nil }

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
			isConnectedC := isConnected(ctx, ch, connParameters{pubKey: test.publicKey}, vpn.NewInternalVPNEvents(), mockTunnel{})

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
	callbackHandler := newTelioCallbackHandler(stateC)
	cb := eventCallbackWrap(callbackHandler)
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

	actualCfg, err := handleTelioConfig(exampleEventPath, true, &remoteConfigGetter)

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

	cfg, err := handleTelioConfig(exampleEventPath, true, &remoteConfigGetter)

	assert.NoError(t, err)

	assert.Nil(t, cfg.Lana)
	assert.Nil(t, cfg.Nurse)
}

func Test_TelioConfigAllDisabled(t *testing.T) {
	category.Set(t, category.Integration)

	remoteConfigGetter := mockVersionGetter{telioRemoteTestConfigAllDisabled}

	cfg, err := handleTelioConfig(exampleEventPath, true, &remoteConfigGetter)

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

func (m *mockVersionGetter) GetConfig() (string, error) {
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
	mu               sync.RWMutex
	counter          int
	eventsReceivedWG *sync.WaitGroup
}

func NewSubscriber(eventsReceivedWG *sync.WaitGroup) subscriber {
	return subscriber{
		eventsReceivedWG: eventsReceivedWG,
	}
}

func (s *subscriber) ConnectionStatusNotifyConnect(data events.DataConnect) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counter++
	s.eventsReceivedWG.Done()
	return nil
}

func (s *subscriber) ConnectionStatusNotifyDisconnect(_ events.DataDisconnect) error {
	return s.ConnectionStatusNotifyConnect(events.DataConnect{})
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
		body   func(context.CancelFunc, chan<- state, *sync.WaitGroup)
	}{
		{
			name:   "ctx done before connection established",
			err:    context.Canceled,
			active: false,
			events: 0,
			body: func(cf context.CancelFunc, _ chan<- state, _ *sync.WaitGroup) {
				cf()
			},
		},
		{
			name:   "successful connection",
			err:    nil,
			active: true,
			events: 1,
			body: func(cf context.CancelFunc, events chan<- state, _ *sync.WaitGroup) {
				events <- state{
					State:  teliogo.NodeStateConnected,
					IsExit: true,
				}
			},
		},
		{
			name:   "ctx done after connection established has no impact",
			err:    nil,
			active: true,
			events: 4,
			body: func(cf context.CancelFunc, events chan<- state, connectionEstablishedWG *sync.WaitGroup) {
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
				time.Sleep(3 * time.Second)
				events <- state{
					State:  teliogo.NodeStateDisconnected,
					IsExit: true,
				}
				// Send one more state which will be ignored just to make sure the
				// earlier states arrived
				events <- state{
					IsExit: false,
				}
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			events := make(chan state, 1)
			pub := vpn.NewInternalVPNEvents()

			eventsReceivedWG := sync.WaitGroup{}
			eventsReceivedWG.Add(tt.events)
			sub := NewSubscriber(&eventsReceivedWG)
			pub.Subscribe(&sub)
			// Create instance
			libtelio := &Libtelio{
				lib:             mockLib{},
				events:          events,
				state:           vpn.ExitedState,
				fwmark:          123,
				eventsPublisher: pub,
				tun:             mockTunnel{},
				callbackHandler: callbackHandlerStub{},
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

			tt.body(cancel, events, &connectionEstablishedWG)

			// Wait until goroutine stops
			connectionEstablishedWG.Wait()
			assert.ErrorIs(t, tt.err, err)
			assert.Equal(t, tt.active, libtelio.active)

			if tt.events > 0 {
				eventsReceivedChan := make(chan interface{})
				go func() {
					eventsReceivedWG.Wait()
					close(eventsReceivedChan)
				}()
				select {
				case <-eventsReceivedChan:
				case <-time.After(time.Millisecond * 10):
					assert.Fail(t, "timeout when waiting for events")
				}
			}

			assert.Equal(t, tt.events, sub.Counter())
		})
	}
}

func Test_EventCallback(t *testing.T) {
	category.Set(t, category.Unit)

	stateChan := make(chan state)
	callbackHandler := newTelioCallbackHandler(stateChan)

	var wg sync.WaitGroup
	callbackHandlerEventNonBlocking := func() {
		callbackHandler.handleEvent(teliogo.EventNode{})
		// in this case Event should not block, as no context was provided, so we can wait until the function exits
		wg.Done()
	}

	wg.Add(1)
	go callbackHandlerEventNonBlocking()
	wg.Wait()

	select {
	case <-stateChan:
		assert.Fail(t, "Event sent when no context was provided.")
	default:
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	// add a context, callbackHandled should start sending events via stateChan
	callbackHandler.setMonitoringContext(ctx)

	wg.Add(1)
	go func() {
		wg.Done()
		callbackHandler.handleEvent(teliogo.EventNode{})
	}()
	wg.Wait()

	select {
	case <-stateChan:
	case <-time.After(5 * time.Second):
		cancelFunc()
		assert.Fail(t, "Event not sent when context was provieded")
	}

	cancelFunc()

	wg.Add(1)
	go callbackHandlerEventNonBlocking()
	wg.Wait()

	select {
	case <-stateChan:
		assert.Fail(t, "Event sent when no context was provided.")
	default:
	}
}
