package nordlynx

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math"
	"net"
	"net/netip"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/tunnel"

	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/ipc"
	"golang.zx2c4.com/wireguard/tun"
)

type UserSpace struct {
	state  vpn.State
	active bool
	fwmark uint32
	tun    *tunnel.Tunnel
	conn   int32
	sync.Mutex
}

func NewUserSpace(fwmark uint32) *UserSpace {
	return &UserSpace{
		state:  vpn.ExitedState,
		fwmark: fwmark,
	}
}

// uapiTemplate is a template for wg-go
const uapiTemplate = `private_key=%s
fwmark=%d
replace_peers=true
public_key=%s
replace_allowed_ips=true
allowed_ip=0.0.0.0/0
allowed_ip=::/0
endpoint=%s
persistent_keepalive_interval=25`

func uapiConfig(
	privateKey string,
	fwmark uint32,
	publicKey string,
	serverIP netip.Addr,
) (string, error) {
	// UAPI requires keys as hex encoded raw bytes
	rawPrivKey, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		return "", fmt.Errorf("decoding private key: %w", err)
	}
	rawPubKey, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return "", fmt.Errorf("decoding public key: %w", err)
	}
	return fmt.Sprintf(uapiTemplate,
		hex.EncodeToString(rawPrivKey),
		fwmark,
		hex.EncodeToString(rawPubKey),
		net.JoinHostPort(
			serverIP.String(),
			strconv.Itoa(defaultPort),
		),
	), nil
}

func (u *UserSpace) Start(
	_ context.Context,
	creds vpn.Credentials,
	serverData vpn.ServerData,
) error {
	u.Lock()
	defer u.Unlock()
	if u.active {
		return vpn.ErrVPNAIsAlreadyStarted
	}

	conf, err := uapiConfig(
		creds.NordLynxPrivateKey,
		u.fwmark,
		serverData.NordLynxPublicKey,
		serverData.IP,
	)
	if err != nil {
		return fmt.Errorf("generating uapi config: %w", err)
	}
	log.Println("UAPI CONFIG:", conf)

	// check if wireguard interface is not up already
	if _, err := exec.Command("ip", "link", "show", "dev", InterfaceName).Output(); err == nil {
		return vpn.ErrTunnelAlreadyExists
	}

	conn, err := wgGoTurnOn(InterfaceName, conf)
	if err != nil {
		return fmt.Errorf("turning on nordlynx: %w", err)
	}

	iface, err := net.InterfaceByName(InterfaceName)
	if err != nil {
		if err := u.stop(); err != nil {
			log.Println(internal.DeferPrefix, err)
		}
		return err
	}

	interfaceIps := []netip.Addr{netip.MustParseAddr("10.5.0.2")}
	ipv6, err := vpn.InterfaceIPv6(serverData.IP, interfaceID())
	if err == nil {
		interfaceIps = append(interfaceIps, ipv6)
	}

	u.conn = conn

	tun := tunnel.New(*iface, interfaceIps, netip.Prefix{})
	u.tun = tun
	if err := tun.AddAddrs(); err != nil {
		if err := u.stop(); err != nil {
			log.Println(internal.DeferPrefix, err)
		}
		return err
	}

	if err := tun.Up(); err != nil {
		if err := u.stop(); err != nil {
			log.Println(internal.DeferPrefix, err)
		}
		return err
	}

	if err := SetMTU(tun.Interface()); err != nil {
		if err := u.stop(); err != nil {
			log.Println(internal.DeferPrefix, err)
		}
		return fmt.Errorf("setting MTU for nordlynx interface: %w", err)
	}

	u.active = true
	u.state = vpn.ConnectedState
	return nil
}

// Stop is used by disconnect command
func (u *UserSpace) Stop() error {
	u.Lock()
	defer u.Unlock()
	if u.state == vpn.ConnectingState {
		u.state = vpn.ExitingState
		return nil
	}
	return u.stop()
}

// stop is used on errors
func (u *UserSpace) stop() error {
	if u.conn >= 0 {
		if err := wgGoTurnOff(u.conn); err != nil {
			return err
		}
	}
	u.conn = 0
	u.active = false
	u.tun = nil
	u.state = vpn.ExitedState
	return nil
}

func (u *UserSpace) IsActive() bool {
	u.Lock()
	defer u.Unlock()
	return u.active
}

func (u *UserSpace) State() vpn.State {
	u.Lock()
	defer u.Unlock()
	return u.state
}

func (u *UserSpace) Tun() tunnel.T {
	u.Lock()
	defer u.Unlock()
	return u.tun
}

type tunnelHandle struct {
	device *device.Device
	uapi   net.Listener
	log    *device.Logger
}

func (t tunnelHandle) Close() error {
	t.device.Close()
	if t.uapi != nil {
		return t.uapi.Close()
	}
	return nil
}

var tHandles map[int32]tunnelHandle = make(map[int32]tunnelHandle)

func wgGoTurnOn(iface string, settings string) (int32, error) {
	var i int32
	for i = 0; i < math.MaxInt32; i++ {
		if _, exists := tHandles[i]; !exists {
			break
		}
	}
	if i == math.MaxInt32 {
		return i, errors.New("out of file descriptors")
	}

	interfaceName := iface

	logLevel := func() int {
		switch os.Getenv("LOG_LEVEL") {
		case "silent":
			return device.LogLevelSilent
		case "verbose":
			return device.LogLevelVerbose
		}
		return device.LogLevelError
	}()

	logger := device.NewLogger(
		logLevel,
		fmt.Sprintf("(%s) ", interfaceName),
	)

	tun, err := tun.CreateTUN(interfaceName, device.DefaultMTU)
	if err != nil {
		return i, fmt.Errorf("creating tun device: %w", err)
	}
	logger.Verbosef("Attaching to interface")
	device := device.NewDevice(tun, conn.NewStdNetBind(), logger)

	err = device.IpcSetOperation(bufio.NewReader(strings.NewReader(settings)))
	if err != nil {
		device.Close()
		return i, fmt.Errorf("setting IPC operation: %w", err)
	}

	var uapi net.Listener
	uapiFile, err := ipc.UAPIOpen(interfaceName)
	if err != nil {
		logger.Errorf("%s", err)
	}

	uapi, err = ipc.UAPIListen(interfaceName, uapiFile)
	if err != nil {
		if err := uapiFile.Close(); err != nil {
			log.Println(internal.DeferPrefix, err)
		}
		return i, fmt.Errorf("listening for UAPI: %w", err)
	}

	go func() {
		for {
			conn, err := uapi.Accept()
			if err != nil {
				return
			}
			go device.IpcHandle(conn)
		}
	}()

	if err = device.Up(); err != nil {
		return i, fmt.Errorf("upping the device: %w", err)
	}

	logger.Verbosef("Device started")

	tHandles[i] = tunnelHandle{device, uapi, logger}
	return i, nil
}

func wgGoTurnOff(tHandle int32) error {
	handle, ok := tHandles[tHandle]
	if !ok {
		return nil
	}
	delete(tHandles, tHandle)
	return handle.Close()
}
