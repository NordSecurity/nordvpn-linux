package network

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"time"

	"github.com/go-ping/ping"
)

// LookupAddressWithCustomDNS looks up address in a specified DNS server
func LookupAddressWithCustomDNS(addr string, dns string, protocol string) ([]netip.Addr, error) {
	resolver := net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			dialer := net.Dialer{Timeout: time.Second * 7}
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
