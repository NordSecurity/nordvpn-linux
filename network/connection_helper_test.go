package network

import (
	"net"
	"net/netip"
	"sync"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/miekg/dns"

	"github.com/stretchr/testify/assert"
)

func startTestDNSServer(t *testing.T, domainName string, ip netip.Addr) (addr string, shutdown func()) {
	t.Helper()

	mux := dns.NewServeMux()
	mux.HandleFunc(".", func(w dns.ResponseWriter, r *dns.Msg) {
		msg := new(dns.Msg)
		msg.SetReply(r)

		for _, q := range r.Question {
			if q.Name == domainName+"." && q.Qtype == dns.TypeA {
				msg.Answer = append(msg.Answer, &dns.A{
					Hdr: dns.RR_Header{
						Name:   q.Name,
						Rrtype: dns.TypeA,
						Class:  dns.ClassINET,
						Ttl:    60,
					},
					A: ip.AsSlice(),
				})
			}
		}

		_ = w.WriteMsg(msg)
	})

	pc, err := net.ListenPacket("udp", "127.0.0.1:5000")
	if err != nil {
		t.Fatalf("listen udp: %v", err)
	}

	server := &dns.Server{
		PacketConn: pc,
		Handler:    mux,
	}

	var wg sync.WaitGroup
	wg.Go(func() {
		_ = server.ActivateAndServe()
	})

	return pc.LocalAddr().String(), func() {
		_ = server.Shutdown()
		_ = pc.Close()
	}
}

func TestResolveHostUsingLocalDnsServer(t *testing.T) {
	category.Set(t, category.Root)

	domainName := "nordvpn.com"
	ipAddress := netip.MustParseAddr("1.2.3.4")

	dnsAddr, shutdown := startTestDNSServer(t, domainName, ipAddress)
	defer shutdown()

	result, err := LookupAddressWithCustomDNS(domainName, dnsAddr, "udp", 0xe1f1)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	if len(result) > 0 {
		assert.Equal(t, ipAddress, result[0])
	}
}

func TestResolveHost(t *testing.T) {
	category.Set(t, category.Root)

	result, err := LookupAddressWithCustomDNS("google.com", "1.1.1.1", "udp", 0x1234)
	assert.NoError(t, err)
	assert.NotEmpty(t, result[0])
	assert.NotEmpty(t, result[1])
}
