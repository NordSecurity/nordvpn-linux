// Package tunnel provides an extension over standard library's net.Interface type.
package tunnel

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

var (
	// ErrNotFound is returned when no tunnel matches the search parameters.
	ErrNotFound = errors.New("tunnel not found")
)

// T describes tunnel behavior
// probably needs a better name, though
type T interface {
	Interface() net.Interface
	IPs() []netip.Addr
	TransferRates() (Statistics, error)
}

// Tunnel encrypts and decrypts network traffic.
type Tunnel struct {
	// might be a good idea to change this to a pointer now
	// so that we could see changes to the interface at real time
	// but this would need testing first to check if it actually works
	iface net.Interface
	ips   []netip.Addr
}

func New(iface net.Interface, ips []netip.Addr) *Tunnel {
	return &Tunnel{iface: iface, ips: ips}
}

// Interface returns the underlying network interface.
func (t *Tunnel) Interface() net.Interface { return t.iface }

// IPs attached to the tunnel.
func (t *Tunnel) IPs() []netip.Addr { return t.ips }

// Statistics defines what information can be collected about the tunnel
type Statistics struct {
	Tx uint64
	Rx uint64
}

// Find a tunnel with given IPs.
func Find(ipAddrs ...netip.Addr) (Tunnel, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return Tunnel{}, err
	}
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			return Tunnel{}, err
		}

		for _, addr := range addrs {
			subnet, err := netip.ParsePrefix(addr.String())
			if err != nil {
				continue
			}

			var ips []netip.Addr
			for _, ip := range ipAddrs {
				if subnet.Contains(ip) {
					ips = append(ips, ip)
				}
			}

			if len(ips) == 0 {
				continue
			}

			return Tunnel{
				iface: iface,
				ips:   ips,
			}, nil
		}
	}
	return Tunnel{}, ErrNotFound
}

func (t *Tunnel) cmdAddrs(cmd string) error {
	for _, ip := range t.ips {
		// #nosec G204 -- input is properly sanitized
		cmd := exec.Command(
			"ip",
			"address",
			cmd,
			fmt.Sprintf("%s/%d", ip.String(), ip.BitLen()),
			"dev",
			t.iface.Name,
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("%s IP address to interface: %s : %w", cmd, string(out), err)
		}
	}
	return nil
}

// AddAddrs to a tunnel interface.
func (t *Tunnel) AddAddrs() error {
	return t.cmdAddrs("add")
}

// DelAddrs from a tunnel interface.
func (t *Tunnel) DelAddrs() error {
	return t.cmdAddrs("del")
}

// Up sets tunnel state to up.
func (t *Tunnel) Up() error {
	fd, err := unix.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_IP)
	if err != nil {
		return err
	}
	defer unix.Close(fd)

	req, err := unix.NewIfreq(t.iface.Name)
	if err != nil {
		return err
	}
	req.SetUint16(req.Uint16() | unix.IFF_UP)

	return unix.IoctlIfreq(fd, unix.SIOCSIFFLAGS, req)
}

// TransferRates collects data transfer statistics.
func (t Tunnel) TransferRates() (Statistics, error) {
	out, err := os.ReadFile("/sys/class/net/" + t.iface.Name + "/statistics/rx_bytes")
	if err != nil {
		return Statistics{}, err
	}

	rx, err := strconv.ParseUint(strings.TrimSpace(string(out)), 10, 64)
	if err != nil {
		return Statistics{}, err
	}

	out, err = os.ReadFile("/sys/class/net/" + t.iface.Name + "/statistics/tx_bytes")
	if err != nil {
		return Statistics{}, err
	}

	tx, err := strconv.ParseUint(strings.TrimSpace(string(out)), 10, 64)
	if err != nil {
		return Statistics{}, err
	}

	return Statistics{Tx: tx, Rx: rx}, nil
}
