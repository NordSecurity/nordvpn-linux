package network

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"syscall"
	"time"

	"github.com/NordSecurity/nordvpn-linux/request"
	"github.com/go-ping/ping"
	"golang.org/x/sys/unix"
)

// LookupAddressWithCustomDNS looks up address in a specified DNS server
func LookupAddressWithCustomDNS(addr string, dns string, protocol string, fwmark uint32) ([]netip.Addr, error) {
	resolver := net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			var operr error
			fwmarkFn := func(fd uintptr) {
				operr = syscall.SetsockoptInt(
					int(fd),
					unix.SOL_SOCKET,
					unix.SO_MARK,
					int(fwmark),
				)
			}
			// dialer := net.Dialer{Timeout: time.Second * 7}
			dialer := &net.Dialer{
				Control: func(network, address string, conn syscall.RawConn) error {
					if err := conn.Control(fwmarkFn); err != nil {
						return err
					}
					return operr
				},
				Timeout: request.DefaultTimeout,
			}
			return dialer.DialContext(ctx, protocol, net.JoinHostPort(dns, "53"))
		},
	}
	ipAddrs, err := resolver.LookupIPAddr(context.Background(), addr)
	if err != nil {
		return nil, fmt.Errorf("looking addr ip up: %w", err)
	}
	var ips []netip.Addr
	for _, ipAddr := range ipAddrs {
		ip, ok := netip.AddrFromSlice(ipAddr.IP)
		if ok {
			ips = append(ips, ip)
		}
	}
	return ips, nil
}

func Ping(addr string, count int) error {
	pinger, err := ping.NewPinger(addr)
	pinger.Timeout = 500 * time.Millisecond
	pinger.SetPrivileged(true)
	if err != nil {
		return fmt.Errorf("unable resolve %s to ping: %w", addr, err)
	}
	pinger.Count = count
	err = pinger.Run()
	if err != nil {
		return fmt.Errorf("unable to ping: %w", err)
	}
	stats := pinger.Statistics()
	if stats.PacketsRecv > 0 {
		return nil
	}
	return fmt.Errorf("no ping response received")
}
