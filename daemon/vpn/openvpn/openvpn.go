// Package openvpn provides OpenVPN technology.
package openvpn

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/netip"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	gopenvpn "github.com/NordSecurity/gopenvpn/openvpn"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/tunnel"
)

const (
	scriptSecurityLevel = "1"
	connectRetryScale   = "5"
	connectRetryMax     = "5"
	authFailureDesc     = "auth-failure"
	timeoutDesc         = "server_poll"
)

var (
	errAccountExpired = errors.New("account expired")
	errServerTimeout  = errors.New("server timeout")
	ErrServerVersion  = errors.New("invalid openvpn server version")
	errExited         = errors.New("exited")
	errNotImplemented = errors.New("not implemented")

	openvpnManagementSocket = filepath.Join(internal.RunDir, "nordvpn-openvpn.sock")
)

type OpenVPN struct {
	process         *exec.Cmd
	manager         *gopenvpn.MgmtClient
	state           vpn.State
	substate        vpn.Substate
	tun             *tunnel.Tunnel
	active          bool
	fwmark          uint32
	eventsPublisher *vpn.Events
	serverData      vpn.ServerData
	// sync.Mutex is used all over the place due to how OpenVPN
	// is managed over the management interface.
	// Simple Lock(); defer Unlock() results in deadlocks, since
	// substates updates get stuck waiting for Mutex.
	sync.Mutex
}

func New(fwmark uint32, eventsPublisher *vpn.Events) *OpenVPN {
	return &OpenVPN{
		state:           vpn.ExitedState,
		substate:        vpn.UnknownSubstate,
		fwmark:          fwmark,
		eventsPublisher: eventsPublisher,
	}
}

// Start starts openvpn process
func (ovpn *OpenVPN) Start(
	creds vpn.Credentials,
	serverData vpn.ServerData,
) error {
	ovpn.Lock()
	ovpn.serverData = serverData

	if ovpn.active {
		ovpn.Unlock()
		return vpn.ErrVPNAIsAlreadyStarted
	}

	if !creds.IsOpenVPNDefined() {
		ovpn.Unlock()
		return errors.New("server credentials not provided")
	}

	err := setOpenVPNConfig(
		serverData.Protocol,
		serverData.IP,
		serverData.Obfuscated,
		serverData.OpenVPNVersion,
	)
	if err != nil {
		ovpn.Unlock()
		return fmt.Errorf("setting openvpn server to connect to: %w", err)
	}

	mgmtCh := make(chan gopenvpn.Event, 10) // closed by demux in NewManagementClient
	clientCh, mErrCh, err := newManagementClient(mgmtCh)
	if err != nil {
		ovpn.Unlock()
		return fmt.Errorf("creating openvpn management client: %w", err)
	}
	defer close(clientCh)
	defer close(mErrCh)

	ovpn.active = true
	// #nosec G204 -- input is properly sanitized
	ovpn.process = exec.Command(
		openVPNExec,
		"--config", openVPNConfigFileName, // path to openVpnConfig to be used
		"--management-client",
		"--management", openvpnManagementSocket, "unix", // enable openvpn management
		"--pull-filter", "ignore", "redirect-gateway", // disable automatic routing
		"--script-security", scriptSecurityLevel,
		"--connect-retry", connectRetryScale, connectRetryMax,
		"--auth-retry", "nointeract",
		"--management-query-passwords",
		"--verify-x509-name", fmt.Sprintf("CN=%s", serverData.Hostname), // certificate validation
		"--mark", strconv.Itoa(int(ovpn.fwmark)),
		"--dev-type", interfaceType,
		"--dev", InterfaceName,
	)
	ovpn.Unlock()

	err = ovpn.startOpenVPN()
	if err != nil {
		// #nosec G104 -- errors.Join would be useful here
		ovpn.stop()
		return fmt.Errorf("starting openvpn: %w", err)
	}

	select {
	case client := <-clientCh:
		ovpn.Lock()
		ovpn.manager = client
		ovpn.Unlock()
	case err := <-mErrCh:
		return err
	case <-time.After(3 * time.Second):
		return errors.New("management timeout")
	}

	err = ovpn.manager.SetStateEvents(true)
	if err != nil {
		// #nosec G104 -- errors.Join would be useful here
		ovpn.stop()
		return fmt.Errorf("setting openvpn state events: %w", err)
	}

	err = stage1Handler(
		ovpn,
		mgmtCh,
		creds.OpenVPNUsername,
		creds.OpenVPNPassword,
	)
	if err != nil {
		if err == errExited {
			// don't call stop() in case of ErrExited
			// it occurs when user invokes the disconnect action while connection is still processing
			// which means that an stop() is already executing
			// running it a second time used to lead to nil dereferences
			return fmt.Errorf("disconnected while previous connection was still being established: %w", err)
		}

		// #nosec G104 -- errors.Join would be useful here
		ovpn.stop()
		return err
	}

	go stage2Handler(
		ovpn,
		mgmtCh,
		creds.OpenVPNUsername,
		creds.OpenVPNPassword,
	)
	return nil
}

// Stop stops openvpn process
func (ovpn *OpenVPN) Stop() error {
	ovpn.Lock()
	ovpn.publishDisconnected()
	if ovpn.active {
		ovpn.Unlock()
		return ovpn.stop()
	}
	ovpn.Unlock()
	return errors.New("not active")
}

// stop actually stops openvpn process
func (ovpn *OpenVPN) stop() error {
	if ovpn.manager != nil {
		if pid, _ := ovpn.manager.Pid(); pid != 0 {
			err := ovpn.manager.SendSignal("SIGINT")
			if err != nil {
				return err
			}
		}
	}

	if ovpn.process.Process != nil {
		if err := ovpn.process.Wait(); err != nil {
			// Don't return here as this could've failed due to non zero return code.
			// This would result in broken ovpn state
			log.Println(internal.ErrorPrefix, "Error on waiting for OpenVPN to stop:", err)
		}

		if err := ovpn.manager.Close(); err != nil {
			return err
		}
	}

	ovpn.Lock()
	defer ovpn.Unlock()
	ovpn.manager = nil
	ovpn.tun = nil
	ovpn.active = false
	ovpn.state = vpn.ExitedState
	return nil
}

func (ovpn *OpenVPN) NetworkChanged() error {
	// application uses the flag --persist-tun and it should not close the TUN interface at restarts
	// but because the servers sends different configuration at pull, then the TUN is closed + recreated
	// https://github.com/OpenVPN/openvpn/blob/a1cb1b47b138b9f654cd0bca5de6d08dbca61888/src/openvpn/init.c#L2421

	return errNotImplemented

	// the following should be the correct implementation for it
	// if !ovpn.IsActive() {
	// 	return nil
	// }

	// ovpn.Lock()
	// defer ovpn.Unlock()

	// if ovpn.manager != nil {
	// 	if pid, _ := ovpn.manager.Pid(); pid != 0 {
	// 		err := ovpn.manager.SendSignal("SIGUSR1")
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}
	// }
	// return nil
}

// IsActive checks if openvpn process is running
func (ovpn *OpenVPN) IsActive() bool {
	ovpn.Lock()
	defer ovpn.Unlock()
	return ovpn.active
}

func (ovpn *OpenVPN) State() vpn.State {
	ovpn.Lock()
	defer ovpn.Unlock()
	return ovpn.state
}

func (ovpn *OpenVPN) startOpenVPN() error {
	stdout, err := ovpn.process.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := ovpn.process.StderrPipe()
	if err != nil {
		return err
	}

	stdoutCh := make(chan struct{})
	stderrCh := make(chan struct{})
	go vpnMonitor(stdout, "INFO", stdoutCh)
	go vpnMonitor(stderr, "ERROR", stderrCh)

	err = ovpn.process.Start()
	if err != nil {
		return err
	}

	select {
	case <-stdoutCh:
		close(stdoutCh)
		return nil
	case <-stderrCh:
		close(stderrCh)
		return nil
	default:
		return nil
	}
}

func (ovpn *OpenVPN) setTun(tun tunnel.Tunnel) {
	ovpn.Lock()
	defer ovpn.Unlock()
	ovpn.tun = &tun
}

func (ovpn *OpenVPN) Tun() tunnel.T {
	ovpn.Lock()
	defer ovpn.Unlock()
	if ovpn.tun != nil {
		return ovpn.tun
	}
	return nil
}

func (ovpn *OpenVPN) setState(arg string) {
	ovpn.Lock()
	defer ovpn.Unlock()
	ovpn.state, _ = vpn.StringToState(arg)

	switch ovpn.state {
	case vpn.ConnectedState:
		ovpn.publishConnected()
	case vpn.ExitingState:
		fallthrough
	case vpn.ExitedState:
		ovpn.publishDisconnected()
	default:
		ovpn.publishConnecting()
	}
}

func (ovpn *OpenVPN) setSubstate(substate vpn.Substate) {
	ovpn.Lock()
	defer ovpn.Unlock()
	ovpn.substate = substate
}

func (ovpn *OpenVPN) getSubstate() vpn.Substate {
	ovpn.Lock()
	defer ovpn.Unlock()
	return ovpn.substate
}

// publishConnecting publishes Connecting event using current stored server data. Thread unsafe.
func (ovpn *OpenVPN) publishConnecting() {
	ovpn.eventsPublisher.Connected.Publish(events.DataConnect{
		Type:                events.ConnectAttempt,
		TargetServerIP:      ovpn.serverData.IP.String(),
		TargetServerCountry: ovpn.serverData.Country,
		TargetServerCity:    ovpn.serverData.City,
	})
}

// publishConnecting publishes Connecting event using current stored server data. Thread unsafe.
func (ovpn *OpenVPN) publishConnected() {
	ovpn.eventsPublisher.Connected.Publish(events.DataConnect{
		Type:                events.ConnectSuccess,
		TargetServerIP:      ovpn.serverData.IP.String(),
		TargetServerCountry: ovpn.serverData.Country,
		TargetServerCity:    ovpn.serverData.City,
	})
}

// publishDisconnected publishes Connecting event using current stored server data. Thread unsafe.
func (ovpn *OpenVPN) publishDisconnected() {
	ovpn.eventsPublisher.Disconnected.Publish(events.DataDisconnect{})
}

// stage1Handler handles events until first successful connection or timeout
func stage1Handler(
	ovpn *OpenVPN,
	eventCh chan gopenvpn.Event,
	username string,
	password string,
) error {
	timeout := time.NewTimer(time.Second * 30)
	for {
		var e gopenvpn.Event
		select {
		case <-timeout.C:
			return errServerTimeout
		case e = <-eventCh:
		}

		switch event := e.(type) {
		case *gopenvpn.FatalEvent:
			return errors.New(e.String())
		case *gopenvpn.PasswordEvent:
			if !strings.Contains(e.String(), "Need 'Auth' username/password") {
				continue
			}
			err := ovpn.manager.Auth(
				username, password,
			)
			if err != nil {
				return err
			}
		case *gopenvpn.StateEvent:
			state := event.NewState()
			ovpn.setState(state)
			switch vpn.State(state) { //nolint:exhaustive
			case vpn.ReconnectingState:
				switch event.Description() {
				case authFailureDesc:
					switch ovpn.getSubstate() { //nolint:exhaustive
					case vpn.UnknownSubstate:
						ovpn.setSubstate(vpn.AuthFlukeSubstate)
					case vpn.AuthFlukeSubstate:
						ovpn.setSubstate(vpn.AuthBadSubstate)
					case vpn.AuthBadSubstate:
						ovpn.setSubstate(vpn.UnknownSubstate)
						return errAccountExpired
					}
				case timeoutDesc:
					switch ovpn.getSubstate() { //nolint:exhaustive
					case vpn.UnknownSubstate:
						ovpn.setSubstate(vpn.TimeoutFlukeSubstate)
					case vpn.TimeoutFlukeSubstate:
						ovpn.setSubstate(vpn.TimeoutSubstate)
					case vpn.TimeoutSubstate:
						ovpn.setSubstate(vpn.UnknownSubstate)
						return errServerTimeout
					}
				}
			case vpn.ExitingState:
				return errExited
			case vpn.ConnectedState:
				ip, err := netip.ParseAddr(event.LocalTunnelAddr())
				if err != nil {
					return err
				}
				tunnel, err := tunnel.Find(ip)
				if err != nil {
					return err
				}
				ovpn.setTun(tunnel)
				// #nosec G104 -- it's okay to ignore an error here
				internal.FileDelete(openVPNConfigFileName)
				return nil
			}
		}
	}
}

// stage2Handler
func stage2Handler(
	ovpn *OpenVPN,
	eventCh chan gopenvpn.Event,
	username string,
	password string,
) {
	for e := range eventCh {
		switch e := e.(type) {
		case *gopenvpn.FatalEvent:
			log.Println(e.String())
			ovpn.setSubstate(vpn.UnknownSubstate)
			return
		case *gopenvpn.PasswordEvent:
			if !strings.Contains(e.String(), "Need 'Auth' username/password") {
				continue
			}
			err := ovpn.manager.Auth(username, password)
			if err != nil {
				log.Println(internal.ErrorPrefix, err)
			}
		case *gopenvpn.StateEvent:
			event := e
			state := event.NewState()
			ovpn.setState(state)
			switch vpn.State(state) { //nolint:exhaustive
			case vpn.ReconnectingState:
				switch event.Description() {
				case timeoutDesc:
					switch ovpn.getSubstate() { //nolint:exhaustive
					case vpn.UnknownSubstate:
						ovpn.setSubstate(vpn.TimeoutFlukeSubstate)
					case vpn.TimeoutFlukeSubstate:
						ovpn.setSubstate(vpn.TimeoutSubstate)
					case vpn.TimeoutSubstate:
						ovpn.setSubstate(vpn.UnknownSubstate)
						return
					}
				}
			case vpn.ConnectedState:
				ip, err := netip.ParseAddr(event.LocalTunnelAddr())
				if err != nil {
					log.Println(internal.ErrorPrefix, err)
				}

				tunnel, err := tunnel.Find(ip)
				if err != nil {
					log.Println(internal.ErrorPrefix, err)
				}
				ovpn.setTun(tunnel) // might set to nil and crash
			}
		}
	}
}

// VPNMonitor reads from the reader and logs the output.
// It may also signal to retry on certain errors.
func vpnMonitor(reader io.Reader, prefix string, inform chan struct{}) {
	cipherErr := "cipher final failed"
	tlsErr := "keys are out of sync"
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		txt := scanner.Text()
		if containsSeveral(txt, []string{cipherErr, tlsErr}) {
			inform <- struct{}{}
		}
		log.Println(prefix, txt)
	}
}

func containsSeveral(s string, ss []string) bool {
	for _, val := range ss {
		if strings.Contains(s, val) {
			return true
		}
	}
	return false
}

func newManagementClient(eventCh chan<- gopenvpn.Event) (chan *gopenvpn.MgmtClient, chan error, error) {
	// free up socket from the previous daemon process
	// #nosec G104 -- it's okay to ignore an error here
	internal.FileDelete(openvpnManagementSocket)

	listener, err := net.Listen("unix", openvpnManagementSocket)
	if err != nil {
		return nil, nil, err
	}

	var clientCh = make(chan *gopenvpn.MgmtClient)
	var errorCh = make(chan error)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			errorCh <- err
		}
		clientCh <- gopenvpn.NewClient(conn, eventCh)
	}()
	return clientCh, errorCh, nil
}
