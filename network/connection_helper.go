package network

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"time"
)

const noFwMark uint32 = 0

func lookupAddress(addr string, dns string, protocol string, fwmark uint32) ([]netip.Addr, error) {
	resolver := net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			dialer := &net.Dialer{
				Timeout: time.Second * 7,
			}

			if fwmark != noFwMark {
				dialer.Control = NewFwmarkControlFn(fwmark)
			}
			// if the server address doesn't have port number then add port 53
			hostAndPortAddress := dns
			if _, _, err := net.SplitHostPort(hostAndPortAddress); err != nil {
				hostAndPortAddress = net.JoinHostPort(dns, "53")
			}
			return dialer.DialContext(ctx, protocol, hostAndPortAddress)
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
