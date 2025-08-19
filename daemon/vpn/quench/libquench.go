//go:build quench

package quench

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/netip"
	"sync"

	quenchBindigns "quench"

	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/tunnel"
)

const (
	quenchPrefix          = "[quench]"
	quenchInterfaceAddr   = "10.3.0.2/16"
	nordWhisperHeaderSize = 80
)

type Logger struct{}

func (l *Logger) Log(logLevel quenchBindigns.LogLevel, message string) {
	logPrefix := ""
	//nolint:exhaustive // We do not use prefixes for other log levels
	switch logLevel {
	case quenchBindigns.LogLevelInfo:
		logPrefix = internal.InfoPrefix
	case quenchBindigns.LogLevelDebug:
		logPrefix = internal.DebugPrefix
	case quenchBindigns.LogLevelError:
		logPrefix = internal.ErrorPrefix
	default:
		logPrefix = internal.DebugPrefix
	}
	log.Println(logPrefix, quenchPrefix, message)
}

type observer struct {
	mu                        sync.Mutex
	currentState              vpn.State
	eventsChan                chan<- vpn.State
	eventsSubscribtionContext context.Context

	eventNotifier *vpn.Events
	// currentServer is used to build vpn event notification
	currentServer vpn.ServerData
	// nicName is used to retrieve tunnel transfer rates to build vpn event notifications
	nicName string
}

func newObserver(eventNotifier *vpn.Events, nicName string) *observer {
	return &observer{
		eventsChan:    nil,
		eventNotifier: eventNotifier,
		nicName:       nicName,
	}
}

func (o *observer) SetServerData(server vpn.ServerData) {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.currentServer = server
}

func (o *observer) SubscribeToEvents(ctx context.Context) <-chan vpn.State {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.eventsChan != nil {
		close(o.eventsChan)
	}

	eventsChan := make(chan vpn.State)
	o.eventsChan = eventsChan
	o.eventsSubscribtionContext = ctx
	return eventsChan
}

func (o *observer) notifyConnectionStateChange(state vpn.State) {
	o.currentState = state
	if o.eventsChan != nil {
		log.Println(internal.DebugPrefix, quenchPrefix, "notifying about connection state change")
		select {
		case o.eventsChan <- state:
		case <-o.eventsSubscribtionContext.Done():
			o.eventsChan = nil
		}
	}
}

func (o *observer) Connecting(uint32) {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Log only when state has changed to ConnectingState from some other state. This will prevent log flood when
	// libquench attempts to reconnect multiple times in no-net scenario.
	if o.currentState != vpn.ConnectingState {
		log.Println(internal.DebugPrefix, quenchPrefix, "connecting to quench server")
	}

	o.notifyConnectionStateChange(vpn.ConnectingState)
	o.eventNotifier.Connected.Publish(vpn.ConnectEvent{
		Status:     events.StatusAttempt,
		TunnelName: o.nicName,
	})
}

func (o *observer) Connected(uint32) {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.notifyConnectionStateChange(vpn.ConnectedState)

	log.Println(internal.DebugPrefix, quenchPrefix, "connected")
	o.eventNotifier.Connected.Publish(vpn.ConnectEvent{
		Status:     events.StatusSuccess,
		TunnelName: o.nicName,
	})
}

func (o *observer) Disconnected(_ uint32, reason quenchBindigns.DisconnectReason) {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.notifyConnectionStateChange(vpn.ExitedState)

	log.Println(internal.DebugPrefix, quenchPrefix, "disconnected:", reason)

	o.eventNotifier.Disconnected.Publish(events.StatusSuccess)
}

type Quench struct {
	mu       sync.Mutex
	fwmark   uint32
	vnicName string
	observer *observer
	logger   *Logger
	state    vpn.State
	server   vpn.ServerData
	vnic     *quenchBindigns.Vnic
	tun      *tunnel.Tunnel
}

func New(fwmark uint32, envIsDev bool, events *vpn.Events) *Quench {
	logLevel := quenchBindigns.LogLevelInfo
	if envIsDev {
		logLevel = quenchBindigns.LogLevelDebug
	}
	logger := Logger{}
	quenchBindigns.SetLogCallback(logLevel, &logger)

	return &Quench{
		fwmark:   fwmark,
		vnicName: internal.NordWhisperInterfaceName,
		observer: newObserver(events, internal.NordWhisperInterfaceName),
		logger:   &logger,
		state:    vpn.ExitedState,
	}
}

func (q *Quench) Start(ctx context.Context, creds vpn.Credentials, server vpn.ServerData) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	opts := quenchBindigns.NewVnicOptions()
	opts.SetFwmark(q.fwmark)

	vnic, err := quenchBindigns.VnicFromName(q.vnicName, q.observer, &opts)
	if err != nil {
		return fmt.Errorf("creating vnic instance: %w", err)
	}

	iface, err := net.InterfaceByName(q.vnicName)
	if err != nil {
		return fmt.Errorf("retrieving the interface: %w", err)
	}

	ip := netip.MustParsePrefix(quenchInterfaceAddr)
	tun := tunnel.New(*iface, ip)

	if err := tun.AddAddrs(); err != nil {
		return fmt.Errorf("setting up vinc: %w", err)
	}

	if err := tun.Up(); err != nil {
		return fmt.Errorf("adding ip address to vnic: %w", err)
	}

	if err := vpn.SetMTU(tun.Interface(), nordWhisperHeaderSize); err != nil {
		return fmt.Errorf("setting MTU for the interface: %w", err)
	}

	addr := fmt.Sprintf("wt://%s:%d/", server.IP, server.NordWhisperPort)

	config := Config{
		Protocol: Protocol{
			Addr: addr,
			Spec: Spec{
				TlsDomain: server.Hostname,
			},
		},
	}

	jsonConfig, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshaling json config: %w", err)
	}

	log.Println(internal.DebugPrefix, "quench config:", string(jsonConfig))

	quenchCreds := quenchBindigns.Credentials{
		User: creds.OpenVPNUsername,
		Pass: creds.OpenVPNPassword,
	}

	eventsContext, eventsCancelFunc := context.WithCancel(ctx)
	eventsChan := q.observer.SubscribeToEvents(eventsContext)
	defer eventsCancelFunc()

	q.observer.SetServerData(server)

	_, err = vnic.Connect(string(jsonConfig), &quenchCreds)
	if err != nil {
		return fmt.Errorf("connecting to a quench server: %w", err)
	}

	q.vnic = vnic
	q.tun = tun
	q.server = server

	log.Println(internal.DebugPrefix, "waiting for connection")
CONNECTION_LOOP:
	for {
		select {
		case <-ctx.Done():
			log.Println(internal.DebugPrefix, "context cancelled before connection was established")
			return ctx.Err()
		case ev := <-eventsChan:
			q.state = ev

			if ev == vpn.ExitedState {
				q.vnic = nil
				q.tun = nil
				q.server = vpn.ServerData{}
				return fmt.Errorf("connection failed")
			}

			if ev == vpn.ConnectedState {
				break CONNECTION_LOOP
			}
		}
	}

	return nil
}

func (q *Quench) Stop() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	ctx, cancelFunc := context.WithCancel(context.Background())
	eventsChan := q.observer.SubscribeToEvents(ctx)
	defer cancelFunc()

	q.state = vpn.ExitingState

	if q.vnic != nil {
		q.vnic.Disconnect()

		for {
			ev := <-eventsChan
			if ev == vpn.ExitedState {
				break
			}
		}

		q.vnic.Destroy()
		q.vnic = nil
	}

	q.tun = nil
	q.state = vpn.ExitedState

	return nil
}

func (q *Quench) State() vpn.State {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.state
}

func (q *Quench) IsActive() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.state == vpn.ConnectedState || q.state == vpn.ConnectingState
}

func (q *Quench) Tun() tunnel.T {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.tun != nil {
		return q.tun
	}
	return nil
}

func (q *Quench) NetworkChanged() error {
	q.mu.Lock()
	defer q.mu.Unlock()
	return fmt.Errorf("not implemented")
}

func (q *Quench) GetConnectionParameters() (vpn.ServerData, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.server, q.state == vpn.ConnectedState
}
