/*
Package libtelio wraps generated Go bindings so that the rest of the
project would not need C dependencies to run unit tests.
*/
package libtelio

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/netip"
	"os/exec"
	"regexp"
	"sync"
	"time"

	teliogo "github.com/NordSecurity/libtelio/ffi/bindings/linux/go"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn/nordlynx"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/tunnel"
)

const (
	// TelioLocalConfigName defines env key for local config value
	TelioLocalConfigName = "TELIO_LOCAL_CFG"
)

type state struct {
	State     string `json:"state"`
	PublicKey string `json:"public_key"`
}

type event struct {
	Body state `json:"body"`
}

type eventFn func(string)

func maskPublicKey(event string) string {
	expr := regexp.MustCompile(`"public_key":(\s)*"(.*?)"`)
	return expr.ReplaceAllString(event, `"public_key":"***"`)
}

func eventCallback(states chan<- state) eventFn {
	return func(s string) {
		log.Println(internal.InfoPrefix + maskPublicKey(s))
		var e event
		if err := json.Unmarshal([]byte(s), &e); err != nil {
			return
		}

		select {
		case states <- e.Body:
		default: // drop if nobody is listening
		}
	}
}

// Libtelio wrapper around generated Go bindings.
// Libtelio has 4 related methods and their combinations as following
// 1. Mesh disabled, calling Start - tunnel must be created with the
// API provided private key and the default IP address (10.5.0.2)
// 2. Mesh enabled, calling Start - tunnel must be re-used and
// connection to the VPN must be done with the meshnet private key and
// IP address
// 3. Mesh enabled, calling Stop - tunnel must stay as it is
// 4. Mesh disabled, calling Stop - tunnel must be destroyed
// 5. VPN connected, calling Enable - tunnel must be re-initiated
// with the meshnet private key and IP address, VPN connection must be
// re-initiated
// 6. VPN disconnected, calling Enable - tunnel must be initiated with
// the meshnet private key and IP address
// 7. VPN connected, calling Disable - tunnel must be re-initiated with the originally saved values provided to Start
// 8. VPN disconnected, calling Disable - tunnel must be destroyed
type Libtelio struct {
	state                   vpn.State
	lib                     teliogo.Telio
	events                  <-chan state
	cancelConnectionMonitor func()
	active                  bool
	tun                     *tunnel.Tunnel
	// This must be the one given from the public interface and
	// retrieved from the API
	currentServer     vpn.ServerData
	currentPrivateKey string
	isMeshEnabled     bool
	meshnetMap        string
	isKernelDisabled  bool
	fwmark            uint32
	eventsPublisher   *vpn.Events
	mu                sync.Mutex
}

var defaultIP = netip.MustParseAddr("10.5.0.2")

type telioFeatures struct {
	Lana      *lanaConfig      `json:"lana,omitempty"`
	Nurse     *nurseConfig     `json:"nurse,omitempty"`
	Direct    *directConfig    `json:"direct,omitempty"`
	Derp      *derpConfig      `json:"derp,omitempty"`
	Wireguard *wireguardConfig `json:"wireguard,omitempty"`
	ExitDns   string           `json:"exit-dns,omitempty"`
}

type lanaConfig struct {
	EventPath string `json:"event_path"`
	Prod      bool   `json:"prod"`
}

type directConfig struct {
	EndpointIntervalSecs int      `json:"endpoint_interval_secs,omitempty"`
	Providers            []string `json:"providers,omitempty"`
}

type nurseConfig struct {
	Fingerprint       string     `json:"fingerprint"`
	HeartbeatInterval int        `json:"heartbeat_interval,omitempty"`
	Qos               *qosConfig `json:"qos,omitempty"`
}

type qosConfig struct {
	RttInterval int      `json:"rtt_interval,omitempty"`
	RttTries    int      `json:"rtt_tries,omitempty"`
	RttTypes    []string `json:"rtt_types,omitempty"`
	Buckets     int      `json:"buckets,omitempty"`
}

type derpConfig struct {
	TcpKeepalive  int `json:"tcp_keepalive,omitempty"`
	DerpKeepalive int `json:"derp_keepalive,omitempty"`
}

type wireguardConfig struct {
	PersistentKeepAlive *persistentKeepAliveConfig `json:"persistent_keepalive,omitempty"`
}

type persistentKeepAliveConfig struct {
	Proxying int `json:"proxying,omitempty"`
	Direct   int `json:"direct,omitempty"`
	Vpn      int `json:"vpn,omitempty"`
	Stun     int `json:"stun,omitempty"`
}

func handleTelioConfig(eventPath, deviceID, version string, prod bool, vpnLibCfg vpn.LibConfigGetter) ([]byte, error) {
	telioConfig := &telioFeatures{}
	cfgString, err := vpnLibCfg.GetConfig(version)
	if err != nil {
		return nil, fmt.Errorf("getting telio config json string: %w", err)
	} else {
		err := json.Unmarshal([]byte(cfgString), &telioConfig)
		if err != nil {
			return nil, fmt.Errorf("unmarshaling telio config json string: %w", err)
		}
	}
	if telioConfig.Lana != nil {
		telioConfig.Lana.EventPath = eventPath
		telioConfig.Lana.Prod = prod
		if telioConfig.Nurse != nil { // nurse depends on lana
			telioConfig.Nurse.Fingerprint = deviceID
		}
	} else {
		telioConfig.Nurse = nil
	}
	return json.Marshal(telioConfig)
}

func New(prod bool, eventPath string, fwmark uint32,
	vpnLibCfg vpn.LibConfigGetter, deviceID, appVersion string, eventsPublisher *vpn.Events) *Libtelio {
	events := make(chan state)
	logLevel := teliogo.TELIOLOGINFO

	var telioConfigString string
	cfg, err := handleTelioConfig(eventPath, deviceID, appVersion, prod, vpnLibCfg)
	if err != nil {
		log.Println(internal.ErrorPrefix, "Failed to get telio config:", err)

		defaultTelioConfig := &telioFeatures{}
		defaultTelioConfig.Lana = &lanaConfig{
			Prod:      prod,
			EventPath: eventPath,
		}
		defaultTelioConfig.Direct = &directConfig{}
		defaultTelioConfig.Nurse = &nurseConfig{
			Fingerprint: deviceID,
		}

		fallbackTelioConfig, err := json.Marshal(defaultTelioConfig)
		if err != nil {
			log.Println(internal.ErrorPrefix, "Couldn't encode default telio config:", err)
			fallbackTelioConfig = []byte(`{"direct":{}}`)
		}
		cfg = fallbackTelioConfig
	}
	telioConfigString = string(cfg)

	log.Println(internal.InfoPrefix, "Telio final config:", telioConfigString)

	return &Libtelio{
		lib: teliogo.NewTelio(
			telioConfigString,
			eventCallback(events),
			logLevel, func(i int, s string) {
				log.Println(
					logLevelToPrefix(teliogo.Enum_SS_telio_log_level(i)),
					"TELIO("+teliogo.TelioGetVersionTag()+"): "+s,
				)
			}),
		events:          events,
		state:           vpn.ExitedState,
		fwmark:          fwmark,
		eventsPublisher: eventsPublisher,
	}
}

func logLevelToPrefix(level teliogo.Enum_SS_telio_log_level) string {
	switch level {
	case teliogo.TELIOLOGCRITICAL, teliogo.TELIOLOGERROR:
		return internal.ErrorPrefix
	case teliogo.TELIOLOGWARNING:
		return internal.WarningPrefix
	case teliogo.TELIOLOGDEBUG, teliogo.TELIOLOGTRACE:
		return internal.DebugPrefix
	default:
		return internal.InfoPrefix
	}
}

// Start initiates the tunnel if it is not yet initiated and initiates
// the connection with the VPN server.
// If only VPN feature is used, tunnel should never be initiated when
// Start is called. If meshnet was enabled before, tunnel already
// exists and this function should re-use that and just initiate the
// connection
func (l *Libtelio) Start(
	creds vpn.Credentials,
	serverData vpn.ServerData,
) (err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	log.Println(internal.InfoPrefix, "libtelio version:", teliogo.TelioGetVersionTag())

	if err = l.openTunnel(defaultIP, creds.NordLynxPrivateKey); err != nil {
		return fmt.Errorf("opening the tunnel: %w", err)
	}

	l.currentServer = serverData
	if err = l.connect(serverData.IP, serverData.NordLynxPublicKey); err != nil {
		return err
	}

	// Remember the values used for connection. This is necessary
	// in case meshnet is enabled and disabled before calling Stop
	l.currentPrivateKey = creds.NordLynxPrivateKey
	return nil
}

// connect to the VPN server
func (l *Libtelio) connect(serverIP netip.Addr, serverPublicKey string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	l.cancelConnectionMonitor = cancel

	// Start monitoring connection events before connecting to not miss any
	isConnectedC := isConnected(ctx, l.events, connParameters{pubKey: serverPublicKey, server: l.currentServer}, l.eventsPublisher)

	if err := toError(l.lib.ConnectToExitNode(
		serverPublicKey,
		"0.0.0.0/0",
		net.JoinHostPort(serverIP.String(), "51820"),
	)); err != nil {
		if !l.isMeshEnabled {
			// only close the tunnel when there was VPN connect problem
			// and meshnet is not active
			// #nosec G104 -- errors.Join would be useful here
			l.closeTunnel()
		}
		cancel()
		return fmt.Errorf("libtelio connect: %w", err)
	}

	// Check if the connection actually happened. Disconnect if
	// no actual connection was created within the timeout
	isConnected := <-isConnectedC
	if !isConnected {
		// #nosec G104 -- errors.Join would be useful here
		l.disconnect()
		return errors.New("connected to nordlynx server but there is no internet as a result")
	}

	l.active = true
	l.state = vpn.ConnectedState
	return nil
}

// Stop breaks the connection with the VPN server.
// After that it checks if the meshnet is enabled or not. In case
// Meshnet is still enabled, it should not destroy the tunnel because
// it is used for meshnet connections. If meshnet is not enabled,
func (l *Libtelio) Stop() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if err := l.disconnect(); err != nil {
		return fmt.Errorf("disconnecting from libtelio: %w", err)
	}

	if !l.isMeshEnabled {
		if err := l.closeTunnel(); err != nil {
			return fmt.Errorf("closing the tunnel: %w", err)
		}
	}
	return nil
}

// disconnect from all the exit nodes, including VPN server
func (l *Libtelio) disconnect() error {
	if l.cancelConnectionMonitor != nil {
		l.cancelConnectionMonitor()
	}

	if err := toError(l.lib.DisconnectFromExitNodes()); err != nil {
		return fmt.Errorf("stopping libtelio: %w", err)
	}
	l.active = false
	l.state = vpn.ExitedState
	return nil
}

func (l *Libtelio) IsActive() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.active
}

func (l *Libtelio) State() vpn.State {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.state
}

func (l *Libtelio) Tun() tunnel.T {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.tun != nil {
		return l.tun
	}

	return nil
}

// Enable initiates the tunnel if it is not initiated yet. It can be
// initiated in case Start method was called before.
// If the tunnel is initiated and VPN is active, this function
// re-initiates the tunnel - sets the meshnet private key to libtelio
// and sets the meshnet IP address and re-creates a connection to VPN
// server with the new private key and IP. These should be supported
// by the VPN server if device is properly registered to meshnet map.
func (l *Libtelio) Enable(ip netip.Addr, privateKey string) (err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	defer func() {
		// Err defer. Revert changes in case something failed
		if err != nil {
			log.Println(internal.ErrorPrefix, "enabling libtelio:", err)
			// #nosec G104 -- errors.Join would be useful here
			l.disable()
		}
	}()

	if err = l.openTunnel(ip, privateKey); err != nil {
		return fmt.Errorf("opening the tunnel: %w", err)
	}

	// If VPN is active, tunnel must be re-initiated in order to
	// use new address and private key. Because of this, VPN
	// connection must be re-created as well
	if l.active {
		if err = l.disconnect(); err != nil {
			return fmt.Errorf("disconnecting from libtelio: %w", err)
		}

		if err = l.updateTunnel(privateKey, ip); err != nil {
			return fmt.Errorf("updating the tunnel: %w", err)
		}

		// Re-connect to the VPN server
		if err = l.connect(l.currentServer.IP, l.currentServer.NordLynxPublicKey); err != nil {
			return fmt.Errorf("reconnecting to server: %w", err)
		}
	}

	// remember that mesh is enabled so we could check for the
	// value during Stop()
	l.isMeshEnabled = true
	return nil
}

// Disable the meshnet for libtelio. If VPN is not active, disable also
// destroys the tunnel. However, if it is active, original private key
// and IP address must be re-set to the ones given by the API because
// device is likely to be removed from the meshnet map and VPN servers
// will not recognize mesh IP and private key anymore.
func (l *Libtelio) Disable() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.disable()
}

func (l *Libtelio) disable() error {
	if err := toError(l.lib.SetMeshnetOff()); err != nil {
		return fmt.Errorf("disabling mesh: %w", err)
	}
	l.isMeshEnabled = false

	if !l.active {
		if err := l.closeTunnel(); err != nil {
			return fmt.Errorf("closing the tunnel: %w", err)
		}
	}

	return nil
}

func (l *Libtelio) NetworkChanged() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if result := l.lib.NotifyNetworkChange(""); result != teliogo.TELIORESOK {
		log.Println(internal.ErrorPrefix, "failed to notify network change:", toError(result))

		if l.active {
			serverIP := l.currentServer.IP
			serverPublicKey := l.currentServer.NordLynxPublicKey
			if err := l.disconnect(); err != nil {
				return err
			}

			if err := l.connect(serverIP, serverPublicKey); err != nil {
				return err
			}
		}

		if l.isMeshEnabled {
			if err := toError(l.lib.SetMeshnetOff()); err != nil {
				return err
			}

			if err := toError(l.lib.SetMeshnet(l.meshnetMap)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (l *Libtelio) Refresh(c mesh.MachineMap) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.tun == nil {
		return nil
	}

	meshnetMap := string(c.Raw)

	result := teliogo.TELIORESOK
	for i := 0; i < 10; i++ {
		if result = l.lib.SetMeshnet(meshnetMap); result == teliogo.TELIORESOK {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}

	if result != teliogo.TELIORESOK {
		return fmt.Errorf("failed to refresh meshnet: %d", result)
	}

	l.meshnetMap = meshnetMap

	return nil
}

type peer struct {
	PublicKey string `json:"public_key"`
	State     string `json:"state"`
}

func (l *Libtelio) StatusMap() (map[string]string, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	var peers []peer
	if err := json.Unmarshal([]byte(l.lib.GetStatusMap()), &peers); err != nil {
		return nil, fmt.Errorf("unmarshalling peer list: %w", err)
	}

	m := map[string]string{}
	for _, p := range peers {
		m[p.PublicKey] = p.State
	}
	return m, nil
}

// openTunnel if not opened already
func (l *Libtelio) openTunnel(ip netip.Addr, privateKey string) (err error) {
	if l.tun != nil {
		return nil
	}

	// clean the network interface from the previous program run
	if _, err := net.InterfaceByName(nordlynx.InterfaceName); err == nil {
		// #nosec G204 -- input is properly sanitized
		if err := exec.Command("ip", "link", "del", nordlynx.InterfaceName).Run(); err != nil {
			log.Println(internal.WarningPrefix, err)
		}
	}

	adapter := teliogo.TELIOADAPTERLINUXNATIVETUN
	if l.isKernelDisabled {
		adapter = teliogo.TELIOADAPTERBORINGTUN
	}

	if err := toError(l.lib.StartNamed(
		privateKey,
		adapter,
		nordlynx.InterfaceName,
	)); err != nil {
		if l.isKernelDisabled {
			return fmt.Errorf("starting libtelio: %w", err)
		}
		adapter = teliogo.TELIOADAPTERBORINGTUN
		if err := toError(l.lib.StartNamed(
			privateKey,
			adapter,
			nordlynx.InterfaceName,
		)); err != nil {
			return fmt.Errorf("starting libtelio on retry with boring-tun: %w", err)
		}
		l.isKernelDisabled = true
	}

	defer func() {
		if err != nil {
			l.lib.Stop()
		}
	}()

	if err = toError(l.lib.SetFwmark(uint(l.fwmark))); err != nil {
		return fmt.Errorf("setting fwmark: %w", err)
	}

	iface, err := net.InterfaceByName(nordlynx.InterfaceName)
	if err != nil {
		return fmt.Errorf("retrieving the interface: %w", err)
	}

	tun := tunnel.New(*iface, []netip.Addr{ip})

	err = tun.AddAddrs()
	if err != nil {
		return fmt.Errorf("adding addresses to the interface: %w", err)
	}

	err = tun.Up()
	if err != nil {
		return fmt.Errorf("upping the interface: %w", err)
	}

	err = nordlynx.SetMTU(tun.Interface())
	if err != nil {
		return fmt.Errorf("setting mtu for the interface: %w", err)
	}

	l.tun = tun
	return nil
}

func (l *Libtelio) closeTunnel() error {
	if l.tun == nil {
		return nil
	}
	if err := toError(l.lib.Stop()); err != nil {
		return fmt.Errorf("stopping libtelio: %w", err)
	}
	l.tun = nil
	return nil
}

func (l *Libtelio) updateTunnel(privateKey string, ip netip.Addr) error {
	if err := l.tun.DelAddrs(); err != nil {
		return fmt.Errorf("deleting interface addrs: %w", err)
	}
	tun := tunnel.New(l.tun.Interface(), []netip.Addr{ip})
	if err := tun.AddAddrs(); err != nil {
		return fmt.Errorf("adding interface addrs: %w", err)
	}

	if err := toError(l.lib.SetPrivateKey(
		privateKey,
	)); err != nil {
		return fmt.Errorf("setting private key: %w", err)
	}

	l.tun = tun
	return nil
}

// Private key generation.
func (l *Libtelio) Private() string {
	return l.lib.GenerateSecretKey()
}

// Public key extraction from private.
func (l *Libtelio) Public(private string) string {
	return l.lib.GeneratePublicKey(private)
}

// isConnected function designed to be called before performing an action which trigger events.
// libtelio is sending back events via callback, to properly catch event from libtelio, event
// is being received in goroutine, but this goroutine has to be 100% started before invoking
// libtelio function (e.g. ConnectToExitNode).
// There was a problem observed on VM (Fedora36 and Ubuntu22) when event from libtelio function
// is not caught, because receiving goroutine is not started yet. So, extra WaitGroup is used
// to make sure this function is exited only after event receiving goroutine has started.
func isConnected(ctx context.Context,
	stateCh <-chan state,
	connParams connParameters,
	eventsPublisher *vpn.Events) <-chan bool {
	// we need waitgroup just to make sure goroutine has started
	var wg sync.WaitGroup
	wg.Add(1)

	connectedCh := make(chan bool)
	go func() {
		wg.Done() // signal that goroutine has started
		monitorConnection(ctx, stateCh, connectedCh, connParams, eventsPublisher)
	}()

	wg.Wait() // wait until goroutine is started

	return connectedCh
}

func publishConnecting(publisher *vpn.Events, server vpn.ServerData) {
	publisher.Connected.Publish(events.DataConnect{
		Type:                events.ConnectAttempt,
		TargetServerIP:      server.IP.String(),
		TargetServerCountry: server.Country,
		TargetServerCity:    server.City,
	})
}

func publishConnected(publisher *vpn.Events, server vpn.ServerData) {
	publisher.Connected.Publish(events.DataConnect{
		Type:                events.ConnectSuccess,
		TargetServerIP:      server.IP.String(),
		TargetServerCountry: server.Country,
		TargetServerCity:    server.City,
	})
}

func publishDisconnected(publisher *vpn.Events) {
	publisher.Disconnected.Publish(events.DataDisconnect{})
}

type connParameters struct {
	pubKey string
	server vpn.ServerData
}

// monitorConnection awaits for incoming state changes from the states chan and publishes appropriate events. Upon
// detecting the 'connected' state for the first time it will send true via isConnected channel and close it afterwards.
// If goroutine is canceled before detecting 'connected', false will be sent via isConnected channel.
func monitorConnection(
	ctx context.Context,
	states <-chan state,
	isConnected chan<- bool,
	connParameters connParameters,
	eventsPublisher *vpn.Events) {
	connectedPublished := false

	for {
		select {
		case state := <-states:
			switch state.State {
			case "connecting":
				publishConnecting(eventsPublisher, connParameters.server)
			case "connected":
				if state.PublicKey == connParameters.pubKey &&
					state.State == "connected" {
					if !connectedPublished {
						isConnected <- true
						connectedPublished = true
						close(isConnected)
					}
				}
				publishConnected(eventsPublisher, connParameters.server)
			case "disconnected":
				publishDisconnected(eventsPublisher)
			}
		case <-ctx.Done():
			publishDisconnected(eventsPublisher)
			if !connectedPublished {
				isConnected <- false
			}
			return
		}
	}
}
